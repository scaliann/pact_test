package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/scaliann/pact_test/config"
	telegramv1 "github.com/scaliann/pact_test/gen/telegram/v1"
	"github.com/scaliann/pact_test/internal/adapters/gotd"
	"github.com/scaliann/pact_test/internal/domain/services"
	"github.com/scaliann/pact_test/internal/transport/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config.New failed: %v", err)
	}

	grpcServer := grpc.NewServer()
	telegramGateway := gotd.NewTelegramGateway(cfg)
	telegramService := services.NewTelegramService(telegramGateway)
	telegramServer := grpcapi.NewTelegramServer(telegramService)

	telegramv1.RegisterTelegramServiceServer(grpcServer, telegramServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := telegramService.Close(shutdownCtx); err != nil {
			log.Printf("service close failed: %v", err)
		}

		grpcServer.GracefulStop()
	}()

	log.Printf("grpc server listening on %s (backend: gotd)", cfg.GRPCListenAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc serve failed: %v", err)
	}
}
