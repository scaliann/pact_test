package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GRPCListenAddr  string
	TelegramAPIID   int
	TelegramAPIHash string
}

func New() (Config, error) {
	var cfg Config

	// .env is optional in containerized/runtime environments.
	_ = godotenv.Load(".env")

	cfg.GRPCListenAddr = envOrDefault("GRPC_LISTEN_ADDR", ":50051")

	apiIDStr := os.Getenv("TELEGRAM_API_ID")
	if apiIDStr == "" {
		return cfg, fmt.Errorf("TELEGRAM_API_ID is not set")
	}

	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil {
		return cfg, fmt.Errorf("invalid TELEGRAM_API_ID: %w", err)
	}
	cfg.TelegramAPIID = apiID

	cfg.TelegramAPIHash = os.Getenv("TELEGRAM_API_HASH")
	if cfg.TelegramAPIHash == "" {
		return cfg, fmt.Errorf("TELEGRAM_API_HASH is not set")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
