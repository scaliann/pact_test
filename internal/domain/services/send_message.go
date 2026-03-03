package services

import (
	"context"
	"time"

	"github.com/scaliann/pact_test/internal/domain"
)

func (s *TelegramService) SendMessage(ctx context.Context, sessionId string, peer string, text string) (int64, error) {
	if sessionId == "" || peer == "" || text == "" {
		return 0, domain.ErrInvalidArgument
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	return s.telegramGateway.SendMessage(ctx, sessionId, peer, text)
}
