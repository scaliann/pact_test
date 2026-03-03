package services

import (
	"context"

	"github.com/scaliann/pact_test/internal/domain"
)

type TelegramService struct {
	telegramGateway TelegramGateway
}

func NewTelegramService(telegramGateway TelegramGateway) *TelegramService {
	return &TelegramService{
		telegramGateway: telegramGateway,
	}
}

type TelegramGateway interface {
	CreateSession(ctx context.Context, sessionId string) (*domain.Session, error)
	RefreshSession(ctx context.Context, sessionId string) (string, error)
	DeleteSession(ctx context.Context, sessionId string) error
	SendMessage(ctx context.Context, sessionId, peer, text string) (int64, error)
	SubscribeMessages(ctx context.Context, sessionId string) (<-chan domain.MessageUpdate, error)
	Close(ctx context.Context) error
}
