package gotd

import (
	"context"

	"github.com/scaliann/pact_test/internal/domain"
)

func (g *TelegramGateway) SubscribeMessages(ctx context.Context, sessionID string) (<-chan domain.MessageUpdate, error) {
	session, err := g.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	select {
	case <-session.done:
		return nil, domain.ErrSessionNotReady
	default:
	}

	ch := make(chan domain.MessageUpdate, subscriberBufferSize)

	session.mu.Lock()
	subscriberID := session.nextSubscriberID
	session.nextSubscriberID++
	session.subscribers[subscriberID] = ch
	session.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
		case <-session.done:
		}
		session.removeSubscriber(subscriberID)
	}()

	return ch, nil
}
