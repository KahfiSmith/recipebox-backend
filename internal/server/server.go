package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"recipebox-backend-go/internal/config"
	"recipebox-backend-go/internal/controller"
	"recipebox-backend-go/internal/repository"
	"recipebox-backend-go/internal/service"
	"gorm.io/gorm"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg config.Config, database *gorm.DB) *Server {
	authRepo := repository.NewAuthGormRepository(database)
	authService := service.NewAuthService(authRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.BcryptCost)
	authController := controller.NewAuthController(authService, cfg.Env == "production", cfg.RefreshTokenTTL)

	router := NewRouter(authController, authService)
	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{httpServer: httpServer}
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start http server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
