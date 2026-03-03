package grpcapi

import (
	"context"

	"github.com/scaliann/pact_test/internal/domain"
)

type TelegramService interface {
	CreateSession(ctx context.Context) (string, string, error)
	RefreshSession(ctx context.Context, sessionId string) (string, error)
	DeleteSession(ctx context.Context, sessionId string) error
	SendMessage(ctx context.Context, sessionId, peer, text string) (int64, error)
	SubscribeMessages(ctx context.Context, sessionId string) (<-chan domain.MessageUpdate, error)
}
