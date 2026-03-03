package gotd

import (
	"context"
	"errors"
	"strings"

	"github.com/scaliann/pact_test/internal/domain"
)

func (g *TelegramGateway) SendMessage(ctx context.Context, sessionId string, peer string, text string) (int64, error) {
	s, err := g.getSession(sessionId)
	if err != nil {
		return 0, err
	}

	s.mu.RLock()
	sender := s.sender
	s.mu.RUnlock()

	if sender == nil {
		return 0, domain.ErrSessionNotReady
	}

	select {
	case <-s.done:
		return 0, domain.ErrSessionNotReady
	default:
	}

	updates, err := sender.Resolve(peer).Text(ctx, text)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "engine was closed") {
			return 0, domain.ErrSessionNotReady
		}
		return 0, err
	}

	msgID := extractMessageID(updates)
	if msgID > 0 {
		return msgID, nil
	}

	// Fallback if Telegram did not return explicit message_id.
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastMsgID++
	return s.lastMsgID, nil
}
