package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *TelegramService) CreateSession(ctx context.Context) (string, string, error) {
	sessionId := uuid.NewString()
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	session, err := s.telegramGateway.CreateSession(ctx, sessionId)
	if err != nil {
		return "", "", fmt.Errorf("telegramGateway.CreateSession: %w", err)
	}

	return session.SessionId, session.QrCode, nil

}
