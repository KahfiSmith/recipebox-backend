package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"recipebox-backend-go/internal/config"
	"recipebox-backend-go/internal/db"
	"recipebox-backend-go/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	database, err := db.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("get sql db: %v", err)
	}
	defer sqlDB.Close()

	apiServer, err := server.NewServer(cfg, database)
	if err != nil {
		log.Fatalf("bootstrap server: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("starting API on %s", cfg.HTTPAddr)
		errCh <- apiServer.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("received signal: %s", sig)
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.GracefulShutdownMs)*time.Millisecond)
	defer cancel()

	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
