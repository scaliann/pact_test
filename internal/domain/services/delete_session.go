package services

import (
	"context"
	"time"

	"github.com/scaliann/pact_test/internal/domain"
)

func (s *TelegramService) DeleteSession(ctx context.Context, sessionId string) error {
	if sessionId == "" {
		return domain.ErrInvalidArgument
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return s.telegramGateway.DeleteSession(ctx, sessionId)
}
