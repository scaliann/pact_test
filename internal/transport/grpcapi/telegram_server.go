package grpcapi

import (
	"context"
	"errors"
	"log"

	telegramv1 "github.com/scaliann/pact_test/gen/telegram/v1"
	"github.com/scaliann/pact_test/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TelegramServer struct {
	telegramv1.UnimplementedTelegramServiceServer
	service TelegramService
}

func NewTelegramServer(service TelegramService) *TelegramServer {
	return &TelegramServer{
		service: service,
	}
}

func (s *TelegramServer) CreateSession(ctx context.Context, _ *telegramv1.CreateSessionRequest) (*telegramv1.CreateSessionResponse, error) {
	session, qr, err := s.service.CreateSession(ctx)
	if err != nil {
		log.Printf("CreateSession failed: %v", err)
		return nil, mapDomainError(err)
	}

	return &telegramv1.CreateSessionResponse{
		SessionId: session,
		QrCode:    qr,
	}, nil
}

func (s *TelegramServer) RefreshSession(ctx context.Context, request *telegramv1.RefreshSessionRequest) (*telegramv1.RefreshSessionResponse, error) {
	qrCode, err := s.service.RefreshSession(ctx, request.GetSessionId())
	if err != nil {
		log.Printf("RefreshSession failed: %v", err)
		return nil, mapDomainError(err)
	}

	return &telegramv1.RefreshSessionResponse{
		SessionId: request.GetSessionId(),
		QrCode:    qrCode,
	}, nil
}

func (s *TelegramServer) DeleteSession(ctx context.Context, request *telegramv1.DeleteSessionRequest) (*telegramv1.DeleteSessionResponse, error) {
	if err := s.service.DeleteSession(ctx, request.GetSessionId()); err != nil {
		log.Printf("DeleteSession failed: %v", err)
		return nil, mapDomainError(err)
	}

	return &telegramv1.DeleteSessionResponse{}, nil
}

func (s *TelegramServer) SendMessage(ctx context.Context, request *telegramv1.SendMessageRequest) (*telegramv1.SendMessageResponse, error) {
	messageID, err := s.service.SendMessage(ctx, request.GetSessionId(), request.GetPeer(), request.GetText())
	if err != nil {
		log.Printf("SendMessage failed: %v", err)
		return nil, mapDomainError(err)
	}
	return &telegramv1.SendMessageResponse{
		MessageId: messageID,
	}, nil
}

func (s *TelegramServer) SubscribeMessages(request *telegramv1.SubscribeMessagesRequest, stream grpc.ServerStreamingServer[telegramv1.MessageUpdate]) error {
	updates, err := s.service.SubscribeMessages(stream.Context(), request.GetSessionId())
	if err != nil {
		log.Printf("SubscribeMessages failed on subscribe: %v", err)
		return mapDomainError(err)
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if err := stream.Send(&telegramv1.MessageUpdate{
				MessageId: update.MessageID,
				From:      update.From,
				Text:      update.Text,
				Timestamp: update.Timestamp,
			}); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return err
			}
		}
	}
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrSessionNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrSessionNotReady), errors.Is(err, domain.ErrSessionAuthorized):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
