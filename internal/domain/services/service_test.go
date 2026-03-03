package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/scaliann/pact_test/internal/domain"
)

type mockGateway struct {
	createSessionFn     func(ctx context.Context, sessionId string) (*domain.Session, error)
	refreshSessionFn    func(ctx context.Context, sessionId string) (string, error)
	deleteSessionFn     func(ctx context.Context, sessionId string) error
	sendMessageFn       func(ctx context.Context, sessionId, peer, text string) (int64, error)
	subscribeMessagesFn func(ctx context.Context, sessionId string) (<-chan domain.MessageUpdate, error)
	closeFn             func(ctx context.Context) error
}

func (m *mockGateway) CreateSession(ctx context.Context, sessionId string) (*domain.Session, error) {
	if m.createSessionFn != nil {
		return m.createSessionFn(ctx, sessionId)
	}
	return nil, nil
}

func (m *mockGateway) RefreshSession(ctx context.Context, sessionId string) (string, error) {
	if m.refreshSessionFn != nil {
		return m.refreshSessionFn(ctx, sessionId)
	}
	return "", nil
}

func (m *mockGateway) DeleteSession(ctx context.Context, sessionId string) error {
	if m.deleteSessionFn != nil {
		return m.deleteSessionFn(ctx, sessionId)
	}
	return nil
}

func (m *mockGateway) SendMessage(ctx context.Context, sessionId, peer, text string) (int64, error) {
	if m.sendMessageFn != nil {
		return m.sendMessageFn(ctx, sessionId, peer, text)
	}
	return 0, nil
}

func (m *mockGateway) SubscribeMessages(ctx context.Context, sessionId string) (<-chan domain.MessageUpdate, error) {
	if m.subscribeMessagesFn != nil {
		return m.subscribeMessagesFn(ctx, sessionId)
	}
	return nil, nil
}

func (m *mockGateway) Close(ctx context.Context) error {
	if m.closeFn != nil {
		return m.closeFn(ctx)
	}
	return nil
}

func TestCreateSessionDelegatesToGateway(t *testing.T) {
	t.Parallel()

	called := false
	service := NewTelegramService(&mockGateway{
		createSessionFn: func(_ context.Context, sessionId string) (*domain.Session, error) {
			called = true
			if _, err := uuid.Parse(sessionId); err != nil {
				t.Fatalf("session id must be uuid, got %q", sessionId)
			}
			return &domain.Session{SessionId: sessionId, QrCode: "tg://login?token=abc"}, nil
		},
	})

	sessionID, qrCode, err := service.CreateSession(context.Background())
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if !called {
		t.Fatal("gateway CreateSession was not called")
	}
	if sessionID == "" {
		t.Fatal("sessionID must not be empty")
	}
	if qrCode == "" {
		t.Fatal("qrCode must not be empty")
	}
}

func TestSendMessageRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	service := NewTelegramService(&mockGateway{
		sendMessageFn: func(_ context.Context, _, _, _ string) (int64, error) {
			t.Fatal("gateway should not be called for invalid input")
			return 0, nil
		},
	})

	_, err := service.SendMessage(context.Background(), "", "@durov", "hello")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestRefreshSessionRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	service := NewTelegramService(&mockGateway{
		refreshSessionFn: func(_ context.Context, _ string) (string, error) {
			t.Fatal("gateway should not be called for invalid input")
			return "", nil
		},
	})

	_, err := service.RefreshSession(context.Background(), "")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestDeleteSessionRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	service := NewTelegramService(&mockGateway{
		deleteSessionFn: func(_ context.Context, _ string) error {
			t.Fatal("gateway should not be called for invalid input")
			return nil
		},
	})

	err := service.DeleteSession(context.Background(), "")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestSubscribeMessagesRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	service := NewTelegramService(&mockGateway{
		subscribeMessagesFn: func(_ context.Context, _ string) (<-chan domain.MessageUpdate, error) {
			t.Fatal("gateway should not be called for invalid input")
			return nil, nil
		},
	})

	_, err := service.SubscribeMessages(context.Background(), "")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
