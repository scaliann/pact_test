# Telegram Multi-Session Service (Go + gotd + gRPC)

Home assignment implementation for a Go Engineer role.

## Features

- gRPC API:
  - `CreateSession`
  - `RefreshSession`
  - `DeleteSession`
  - `SendMessage`
  - `SubscribeMessages` (server stream)
- Multiple isolated Telegram sessions.
- QR authorization flow.
- Send text messages.
- Receive incoming text messages (`message_id`, `from`, `text`, `timestamp`).
- Web UI for manual testing:
  - create session
  - refresh QR
  - send message
  - subscribe to incoming messages
  - delete session
- Unit tests for service layer and gRPC transport.

## Stack

- Go `1.26`
- Telegram client: `github.com/gotd/td`
- API: gRPC + Protobuf
- State storage: in-memory

## Configuration

Environment variables:

- `TELEGRAM_API_ID` (required)
- `TELEGRAM_API_HASH` (required)
- `GRPC_LISTEN_ADDR` (default `:50051`)
- `TELEGRAM_UI_ADDR` (default `:8085`)
- `TELEGRAM_GRPC_ADDR` (default `127.0.0.1:50051`)

See `.env.example`.

## Local run

1. Create `.env` from `.env.example`.

Linux/macOS:

```bash
cp .env.example .env
```

PowerShell:

```powershell
Copy-Item .env.example .env
```

2. Fill `TELEGRAM_API_ID` and `TELEGRAM_API_HASH`.

3. Start gRPC server:

```bash
go run ./cmd/app
```

4. In another terminal start UI:

```bash
go run ./cmd/ui
```

5. Open UI: `http://localhost:8085`

## Docker Compose run

1. Create `.env` from `.env.example` and fill Telegram credentials.
2. Run:

```bash
docker compose up --build
```

After startup:

- gRPC: `localhost:50051`
- UI: `http://localhost:8085`

## grpcurl examples

gRPC reflection is enabled.

Create session:

```bash
grpcurl -plaintext -d '{}' localhost:50051 telegram.v1.TelegramService/CreateSession
```

Refresh session QR:

```bash
grpcurl -plaintext -d '{"session_id":"<SESSION_ID>"}' localhost:50051 telegram.v1.TelegramService/RefreshSession
```

Send message:

```bash
grpcurl -plaintext -d '{"session_id":"<SESSION_ID>","peer":"@durov","text":"hello"}' localhost:50051 telegram.v1.TelegramService/SendMessage
```

Subscribe to incoming messages:

```bash
grpcurl -plaintext -d '{"session_id":"<SESSION_ID>"}' localhost:50051 telegram.v1.TelegramService/SubscribeMessages
```

Delete session:

```bash
grpcurl -plaintext -d '{"session_id":"<SESSION_ID>"}' localhost:50051 telegram.v1.TelegramService/DeleteSession
```

## Architecture

- `cmd/app`: gRPC server bootstrap and graceful shutdown.
- `cmd/ui`: HTTP UI + gRPC proxy handlers.
- `internal/domain`: entities and domain errors.
- `internal/domain/services`: use-cases and validation.
- `internal/adapters/gotd`: Telegram integration, session lifecycle, updates.
- `internal/transport/grpcapi`: gRPC handlers and error mapping.

Key decisions:

- In-memory session storage.
- Buffered per-session subscribers for incoming messages.
- Non-blocking fan-out to avoid slow consumer blocking update handling.
- Delete session stops client, closes subscriptions, and performs best-effort `auth.logOut` for authorized sessions.

## Tests

```bash
go test ./...
```

Covered packages:

- `internal/domain/services`
- `internal/transport/grpcapi`

## Current limitations

- In-memory storage only (sessions are lost after restart).
- UI is for debugging/demo, not production.
