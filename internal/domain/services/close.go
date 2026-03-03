package services

import "context"

func (s *TelegramService) Close(ctx context.Context) error {
	return s.telegramGateway.Close(ctx)
}
