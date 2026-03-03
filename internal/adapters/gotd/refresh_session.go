package gotd

import (
	"context"

	"github.com/scaliann/pact_test/internal/domain"
)

func (g *TelegramGateway) RefreshSession(ctx context.Context, sessionID string) (string, error) {
	session, err := g.getSession(sessionID)
	if err != nil {
		return "", err
	}

	qrCode, _, authorized, authErr := session.snapshotState()
	if authErr != nil {
		return "", authErr
	}
	if authorized {
		return "", domain.ErrSessionAuthorized
	}
	if qrCode != "" {
		return qrCode, nil
	}

	return g.waitForQRCode(ctx, session, 0, true)
}
