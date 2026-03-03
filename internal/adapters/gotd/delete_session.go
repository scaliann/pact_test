package gotd

import (
	"context"
	"time"

	"github.com/gotd/td/tg"
)

func (g *TelegramGateway) DeleteSession(ctx context.Context, sessionID string) error {
	session, err := g.getSession(sessionID)
	if err != nil {
		return err
	}

	session.mu.RLock()
	authorized := session.authorized
	session.mu.RUnlock()

	if authorized {
		logoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		_, _ = tg.NewClient(session.client).AuthLogOut(logoutCtx)
		cancel()
	}

	session.runCancel()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-session.done:
	}

	g.removeSession(sessionID)
	return nil
}

func (g *TelegramGateway) Close(ctx context.Context) error {
	sessions := g.snapshotSessions()
	for _, session := range sessions {
		session.runCancel()
	}

	for _, session := range sessions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-session.done:
		}
		g.removeSession(session.id)
	}

	return nil
}
