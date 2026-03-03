# Telegram Multi-Session Service (Go + gotd + gRPC)



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


## Docker Compose run

1. Create `.env` from `.env.example` and fill Telegram credentials.
2. Run:

```bash
docker compose up --build
```

After startup:

- gRPC: `localhost:50051`
- UI: `http://localhost:8085`


## Tests

```bash
go test ./...
```


