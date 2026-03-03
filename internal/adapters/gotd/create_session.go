package gotd

import (
	"context"
	"errors"
	"fmt"
	"time"

	tgsession "github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/scaliann/pact_test/internal/domain"
)

const (
	createWaitTimeout = 20 * time.Second
	qrPollInterval    = 150 * time.Millisecond
)

func (g *TelegramGateway) CreateSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	if sessionID == "" || g.apiID <= 0 || g.apiHash == "" {
		return nil, domain.ErrInvalidArgument
	}

	runCtx, runCancel := context.WithCancel(context.Background())
	dispatcher := tg.NewUpdateDispatcher()

	managed := &managedSession{
		id:          sessionID,
		dispatcher:  dispatcher,
		runCtx:      runCtx,
		runCancel:   runCancel,
		done:        make(chan error, 1),
		subscribers: make(map[uint64]chan domain.MessageUpdate),
	}

	g.registerUpdateHandlers(managed)

	managed.client = telegram.NewClient(g.apiID, g.apiHash, telegram.Options{
		SessionStorage: new(tgsession.StorageMemory),
		UpdateHandler:  dispatcher,
	})

	g.setSession(sessionID, managed)
	go g.runSession(managed)

	waitCtx, cancel := context.WithTimeout(ctx, createWaitTimeout)
	defer cancel()

	qrCode, err := g.waitForQRCode(waitCtx, managed, 0, true)
	if err != nil {
		managed.runCancel()
		g.removeSession(sessionID)
		select {
		case <-managed.done:
		case <-time.After(2 * time.Second):
		}
		return nil, fmt.Errorf("wait for first QR: %w", err)
	}

	return &domain.Session{
		SessionId: sessionID,
		QrCode:    qrCode,
	}, nil
}

func (g *TelegramGateway) runSession(session *managedSession) {
	defer g.removeSession(session.id)

	err := session.client.Run(session.runCtx, func(ctx context.Context) error {
		session.mu.Lock()
		session.sender = message.NewSender(tg.NewClient(session.client))
		session.mu.Unlock()

		qr := session.client.QR()
		loggedIn := qrlogin.OnLoginToken(session.dispatcher)

		_, authErr := qr.Auth(ctx, loggedIn, func(_ context.Context, token qrlogin.Token) error {
			session.mu.Lock()
			session.qrCode = token.URL()
			session.qrVersion++
			session.mu.Unlock()
			return nil
		})
		if authErr != nil && !errors.Is(authErr, context.Canceled) {
			session.setAuthError(authErr)
			return authErr
		}

		if authErr == nil {
			session.mu.Lock()
			session.authorized = true
			session.mu.Unlock()
		}

		<-ctx.Done()
		return ctx.Err()
	})

	if auth.IsUnauthorized(err) {
		session.setAuthError(err)
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		session.setAuthError(err)
	}

	session.closeSubscribers()
	session.done <- err
	close(session.done)
}

func (g *TelegramGateway) waitForQRCode(ctx context.Context, session *managedSession, minVersion uint64, allowCurrent bool) (string, error) {
	ticker := time.NewTicker(qrPollInterval)
	defer ticker.Stop()

	for {
		qrCode, qrVersion, authorized, authErr := session.snapshotState()
		if authErr != nil {
			return "", authErr
		}
		if authorized {
			return "", domain.ErrSessionAuthorized
		}
		if qrCode != "" && (allowCurrent || qrVersion > minVersion) {
			return qrCode, nil
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case err := <-session.done:
			if err == nil || errors.Is(err, context.Canceled) {
				return "", errors.New("session stopped before QR was exported")
			}
			return "", err
		case <-ticker.C:
		}
	}
}

func (s *managedSession) setAuthError(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	s.authErr = err
	s.mu.Unlock()
}
