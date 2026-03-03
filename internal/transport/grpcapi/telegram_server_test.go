package grpcapi

import (
	"context"
	"errors"
	"testing"

	telegramv1 "github.com/scaliann/pact_test/gen/telegram/v1"
	"github.com/scaliann/pact_test/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeTelegramService struct {
	createSessionFn     func(ctx context.Context) (string, string, error)
	refreshSessionFn    func(ctx context.Context, sessionId string) (string, error)
	deleteSessionFn     func(ctx context.Context, sessionId string) error
	sendMessageFn       func(ctx context.Context, sessionId, peer, text string) (int64, error)
	subscribeMessagesFn func(ctx context.Context, sessionId string) (<-chan domain.MessageUpdate, error)
}

func (s *fakeTelegramService) CreateSession(ctx context.Context) (string, string, error) {
	if s.createSessionFn != nil {
		return s.createSessionFn(ctx)
	}
	return "", "", nil
}

func (s *fakeTelegramService) RefreshSession(ctx context.Context, sessionId string) (string, error) {
	if s.refreshSessionFn != nil {
		return s.refreshSessionFn(ctx, sessionId)
	}
	return "", nil
}

func (s *fakeTelegramService) DeleteSession(ctx context.Context, sessionId string) error {
	if s.deleteSessionFn != nil {
		return s.deleteSessionFn(ctx, sessionId)
	}
	return nil
}

func (s *fakeTelegramService) SendMessage(ctx context.Context, sessionId, peer, text string) (int64, error) {
	if s.sendMessageFn != nil {
		return s.sendMessageFn(ctx, sessionId, peer, text)
	}
	return 0, nil
}

func (s *fakeTelegramService) SubscribeMessages(ctx context.Context, sessionId string) (<-chan domain.MessageUpdate, error) {
	if s.subscribeMessagesFn != nil {
		return s.subscribeMessagesFn(ctx, sessionId)
	}
	return nil, nil
}

type mockMessageStream struct {
	ctx context.Context

	sent []*telegramv1.MessageUpdate
}

func (m *mockMessageStream) Send(update *telegramv1.MessageUpdate) error {
	m.sent = append(m.sent, update)
	return nil
}

func (m *mockMessageStream) SetHeader(_ metadata.MD) error {
	return nil
}

func (m *mockMessageStream) SendHeader(_ metadata.MD) error {
	return nil
}

func (m *mockMessageStream) SetTrailer(_ metadata.MD) {}

func (m *mockMessageStream) Context() context.Context {
	return m.ctx
}

func (m *mockMessageStream) SendMsg(any) error {
	return nil
}

func (m *mockMessageStream) RecvMsg(any) error {
	return nil
}

func TestMapDomainError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "invalid argument", err: domain.ErrInvalidArgument, code: codes.InvalidArgument},
		{name: "not found", err: domain.ErrSessionNotFound, code: codes.NotFound},
		{name: "not ready", err: domain.ErrSessionNotReady, code: codes.FailedPrecondition},
		{name: "authorized", err: domain.ErrSessionAuthorized, code: codes.FailedPrecondition},
		{name: "canceled", err: context.Canceled, code: codes.Canceled},
		{name: "deadline", err: context.DeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "internal", err: errors.New("boom"), code: codes.Internal},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mapped := mapDomainError(test.err)
			if status.Code(mapped) != test.code {
				t.Fatalf("expected code %v, got %v", test.code, status.Code(mapped))
			}
		})
	}
}

func TestSubscribeMessagesStreamsUpdates(t *testing.T) {
	t.Parallel()

	service := &fakeTelegramService{
		subscribeMessagesFn: func(_ context.Context, _ string) (<-chan domain.MessageUpdate, error) {
			updates := make(chan domain.MessageUpdate, 1)
			updates <- domain.MessageUpdate{
				MessageID: 42,
				From:      "@durov",
				Text:      "hello",
				Timestamp: 1700000000,
			}
			close(updates)
			return updates, nil
		},
	}

	server := NewTelegramServer(service)
	stream := &mockMessageStream{ctx: context.Background()}

	err := server.SubscribeMessages(&telegramv1.SubscribeMessagesRequest{SessionId: "session-1"}, stream)
	if err != nil {
		t.Fatalf("SubscribeMessages() error = %v", err)
	}

	if len(stream.sent) != 1 {
		t.Fatalf("expected 1 update, got %d", len(stream.sent))
	}

	got := stream.sent[0]
	if got.GetMessageId() != 42 || got.GetFrom() != "@durov" || got.GetText() != "hello" || got.GetTimestamp() != 1700000000 {
		t.Fatalf("unexpected update payload: %+v", got)
	}
}

func TestSubscribeMessagesMapsServiceError(t *testing.T) {
	t.Parallel()

	service := &fakeTelegramService{
		subscribeMessagesFn: func(_ context.Context, _ string) (<-chan domain.MessageUpdate, error) {
			return nil, domain.ErrSessionNotFound
		},
	}

	server := NewTelegramServer(service)
	stream := &mockMessageStream{ctx: context.Background()}

	err := server.SubscribeMessages(&telegramv1.SubscribeMessagesRequest{SessionId: "missing"}, stream)
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", status.Code(err))
	}
}
