package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	telegramv1 "github.com/scaliann/pact_test/gen/telegram/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed index.html
var assets embed.FS

type uiServer struct {
	grpcAddr string
}

type sendMessageInput struct {
	SessionID string `json:"session_id"`
	Peer      string `json:"peer"`
	Text      string `json:"text"`
}

type sessionInput struct {
	SessionID string `json:"session_id"`
}

func main() {
	httpAddr := getenv("TELEGRAM_UI_ADDR", ":8085")
	grpcAddr := getenv("TELEGRAM_GRPC_ADDR", "127.0.0.1:50051")

	srv := &uiServer{grpcAddr: grpcAddr}
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/api/create-session", srv.handleCreateSession)
	mux.HandleFunc("/api/refresh-session", srv.handleRefreshSession)
	mux.HandleFunc("/api/delete-session", srv.handleDeleteSession)
	mux.HandleFunc("/api/send-message", srv.handleSendMessage)
	mux.HandleFunc("/api/subscribe-messages", srv.handleSubscribeMessages)

	log.Printf("ui: http=%s grpc=%s", httpAddr, grpcAddr)
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal(err)
	}
}

func (s *uiServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	content, err := assets.ReadFile("index.html")
	if err != nil {
		http.Error(w, "cannot load page", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(content)
}

func (s *uiServer) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	client, closeConn, err := s.grpcClient(ctx, 5*time.Second)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	defer closeConn()

	resp, err := client.CreateSession(ctx, &telegramv1.CreateSessionRequest{})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session_id": resp.GetSessionId(),
		"qr_code":    resp.GetQrCode(),
	})
}

func (s *uiServer) handleRefreshSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in sessionInput
	if err := decodeJSONBody(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(in.SessionID) == "" {
		writeError(w, http.StatusBadRequest, errors.New("session_id is required"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	client, closeConn, err := s.grpcClient(ctx, 5*time.Second)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	defer closeConn()

	resp, err := client.RefreshSession(ctx, &telegramv1.RefreshSessionRequest{SessionId: in.SessionID})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session_id": resp.GetSessionId(),
		"qr_code":    resp.GetQrCode(),
	})
}

func (s *uiServer) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in sessionInput
	if err := decodeJSONBody(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(in.SessionID) == "" {
		writeError(w, http.StatusBadRequest, errors.New("session_id is required"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	client, closeConn, err := s.grpcClient(ctx, 5*time.Second)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	defer closeConn()

	if _, err := client.DeleteSession(ctx, &telegramv1.DeleteSessionRequest{SessionId: in.SessionID}); err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *uiServer) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in sendMessageInput
	if err := decodeJSONBody(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if in.SessionID == "" || in.Peer == "" || in.Text == "" {
		writeError(w, http.StatusBadRequest, errors.New("session_id, peer and text are required"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	client, closeConn, err := s.grpcClient(ctx, 5*time.Second)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	defer closeConn()

	resp, err := client.SendMessage(ctx, &telegramv1.SendMessageRequest{
		SessionId: in.SessionID,
		Peer:      in.Peer,
		Text:      in.Text,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message_id": resp.GetMessageId()})
}

func (s *uiServer) handleSubscribeMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := strings.TrimSpace(r.URL.Query().Get("session_id"))
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, errors.New("session_id is required"))
		return
	}

	client, closeConn, err := s.grpcClient(r.Context(), 5*time.Second)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	defer closeConn()

	stream, err := client.SubscribeMessages(r.Context(), &telegramv1.SubscribeMessagesRequest{SessionId: sessionID})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	for {
		update, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return
			}
			writeSSEError(w, flusher, err)
			return
		}

		payload, err := json.Marshal(map[string]any{
			"message_id": update.GetMessageId(),
			"from":       update.GetFrom(),
			"text":       update.GetText(),
			"timestamp":  update.GetTimestamp(),
		})
		if err != nil {
			writeSSEError(w, flusher, err)
			return
		}

		if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
			return
		}
		flusher.Flush()
	}
}

func (s *uiServer) grpcClient(ctx context.Context, dialTimeout time.Duration) (telegramv1.TelegramServiceClient, func(), error) {
	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	conn, err := grpc.DialContext(
		dialCtx,
		s.grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	return telegramv1.NewTelegramServiceClient(conn), func() {
		cancel()
		_ = conn.Close()
	}, nil
}

func decodeJSONBody(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}

	return nil
}

func writeSSEError(w http.ResponseWriter, flusher http.Flusher, err error) {
	payload, marshalErr := json.Marshal(map[string]string{"error": err.Error()})
	if marshalErr != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "event: error\ndata: %s\n\n", payload)
	flusher.Flush()
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
