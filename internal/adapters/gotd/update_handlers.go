package gotd

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/scaliann/pact_test/internal/domain"
)

func (g *TelegramGateway) registerUpdateHandlers(session *managedSession) {
	session.dispatcher.OnNewMessage(func(_ context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
		g.handleIncomingMessage(session, entities, update.Message)
		return nil
	})
	session.dispatcher.OnNewChannelMessage(func(_ context.Context, entities tg.Entities, update *tg.UpdateNewChannelMessage) error {
		g.handleIncomingMessage(session, entities, update.Message)
		return nil
	})
}

func (g *TelegramGateway) handleIncomingMessage(session *managedSession, entities tg.Entities, message tg.MessageClass) {
	msg, ok := message.(*tg.Message)
	if !ok || msg == nil || msg.Out || msg.Message == "" {
		return
	}

	fromPeer, hasFrom := msg.GetFromID()
	if !hasFrom {
		fromPeer = msg.GetPeerID()
	}

	session.publish(domain.MessageUpdate{
		MessageID: int64(msg.ID),
		From:      formatPeerDisplay(fromPeer, entities),
		Text:      msg.Message,
		Timestamp: int64(msg.Date),
	})
}

func formatPeerDisplay(peer tg.PeerClass, entities tg.Entities) string {
	switch p := peer.(type) {
	case *tg.PeerUser:
		if user := entities.Users[p.UserID]; user != nil {
			if username, ok := user.GetUsername(); ok && username != "" {
				return "@" + username
			}
			first, _ := user.GetFirstName()
			last, _ := user.GetLastName()
			fullName := strings.TrimSpace(first + " " + last)
			if fullName != "" {
				return fullName
			}
		}
		return fmt.Sprintf("user:%d", p.UserID)
	case *tg.PeerChat:
		if chat := entities.Chats[p.ChatID]; chat != nil {
			if title := chat.GetTitle(); title != "" {
				return title
			}
		}
		return fmt.Sprintf("chat:%d", p.ChatID)
	case *tg.PeerChannel:
		if channel := entities.Channels[p.ChannelID]; channel != nil {
			if title := channel.GetTitle(); title != "" {
				return title
			}
		}
		return fmt.Sprintf("channel:%d", p.ChannelID)
	default:
		return "unknown"
	}
}
