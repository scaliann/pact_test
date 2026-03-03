package gotd

import (
	"context"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/scaliann/pact_test/config"
	"github.com/scaliann/pact_test/internal/domain"
)

const subscriberBufferSize = 64

type managedSession struct {
	id string

	mu sync.RWMutex

	qrCode     string
	qrVersion  uint64
	authorized bool
	authErr    error

	sender     *message.Sender
	dispatcher tg.UpdateDispatcher
	client     *telegram.Client

	runCtx    context.Context
	runCancel context.CancelFunc
	done      chan error

	subscribers      map[uint64]chan domain.MessageUpdate
	nextSubscriberID uint64

	lastMsgID int64
}

func (s *managedSession) snapshotState() (qrCode string, qrVersion uint64, authorized bool, authErr error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.qrCode, s.qrVersion, s.authorized, s.authErr
}

func (s *managedSession) removeSubscriber(subscriberID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[subscriberID]; ok {
		delete(s.subscribers, subscriberID)
		close(ch)
	}
}

func (s *managedSession) closeSubscribers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for subscriberID, ch := range s.subscribers {
		delete(s.subscribers, subscriberID)
		close(ch)
	}
}

func (s *managedSession) publish(update domain.MessageUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, subscriber := range s.subscribers {
		select {
		case subscriber <- update:
		default:
			// Drop oldest item to keep stream non-blocking under slow consumers.
			select {
			case <-subscriber:
			default:
			}
			select {
			case subscriber <- update:
			default:
			}
		}
	}
}

type TelegramGateway struct {
	apiID   int
	apiHash string

	mu       sync.RWMutex
	sessions map[string]*managedSession
}

func NewTelegramGateway(cfg config.Config) *TelegramGateway {
	return &TelegramGateway{
		apiID:    cfg.TelegramAPIID,
		apiHash:  cfg.TelegramAPIHash,
		sessions: make(map[string]*managedSession),
	}
}

func (g *TelegramGateway) getSession(sessionID string) (*managedSession, error) {
	g.mu.RLock()
	session, exists := g.sessions[sessionID]
	g.mu.RUnlock()

	if !exists {
		return nil, domain.ErrSessionNotFound
	}

	return session, nil
}

func (g *TelegramGateway) setSession(sessionID string, session *managedSession) {
	g.mu.Lock()
	g.sessions[sessionID] = session
	g.mu.Unlock()
}

func (g *TelegramGateway) removeSession(sessionID string) {
	g.mu.Lock()
	delete(g.sessions, sessionID)
	g.mu.Unlock()
}

func (g *TelegramGateway) snapshotSessions() []*managedSession {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*managedSession, 0, len(g.sessions))
	for _, session := range g.sessions {
		result = append(result, session)
	}
	return result
}

func extractMessageID(updates tg.UpdatesClass) int64 {
	switch u := updates.(type) {
	case *tg.UpdateShortSentMessage:
		return int64(u.ID)
	case *tg.Updates:
		return findMessageIDInUpdates(u.Updates)
	case *tg.UpdatesCombined:
		return findMessageIDInUpdates(u.Updates)
	case *tg.UpdateShort:
		return findMessageIDInUpdates([]tg.UpdateClass{u.Update})
	default:
		return 0
	}
}

func findMessageIDInUpdates(updates []tg.UpdateClass) int64 {
	for _, update := range updates {
		switch u := update.(type) {
		case *tg.UpdateMessageID:
			if u.ID > 0 {
				return int64(u.ID)
			}
		case *tg.UpdateNewMessage:
			if msg, ok := u.Message.(*tg.Message); ok && msg.ID > 0 {
				return int64(msg.ID)
			}
		case *tg.UpdateNewChannelMessage:
			if msg, ok := u.Message.(*tg.Message); ok && msg.ID > 0 {
				return int64(msg.ID)
			}
		}
	}
	return 0
}
