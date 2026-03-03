package services

import (
	"context"
	"time"

	"github.com/scaliann/pact_test/internal/domain"
)

func (s *TelegramService) RefreshSession(ctx context.Context, sessionId string) (string, error) {
	if sessionId == "" {
		return "", domain.ErrInvalidArgument
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	return s.telegramGateway.RefreshSession(ctx, sessionId)
}
