package services

import (
	"context"

	"github.com/scaliann/pact_test/internal/domain"
)

func (s *TelegramService) SubscribeMessages(ctx context.Context, sessionId string) (<-chan domain.MessageUpdate, error) {
	if sessionId == "" {
		return nil, domain.ErrInvalidArgument
	}

	return s.telegramGateway.SubscribeMessages(ctx, sessionId)
}
