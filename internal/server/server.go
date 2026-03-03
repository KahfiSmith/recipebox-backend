package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"recipebox-backend-go/internal/config"
	"recipebox-backend-go/internal/controller"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/notification"
	"recipebox-backend-go/internal/redisx"
	"recipebox-backend-go/internal/repository"
	"recipebox-backend-go/internal/service"
	"gorm.io/gorm"
)

type Server struct {
	httpServer  *http.Server
	redisClient *redisx.Client
}

func NewServer(cfg config.Config, database *gorm.DB) (*Server, error) {
	redisClient := redisx.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(pingCtx); err != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	authRepo := repository.NewAuthGormRepository(database)
	authService := service.NewAuthService(authRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL, cfg.BcryptCost)
	if cfg.SMTPHost != "" {
		sender := notification.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFromEmail, cfg.SMTPFromName)
		authService.ConfigureEmailDelivery(sender, cfg.FrontendBaseURL, cfg.AuthDebugExposeTokens)
	} else {
		if !cfg.AuthDebugExposeTokens {
			log.Printf("warning: email delivery unavailable while token exposure is disabled")
		}
		authService.ConfigureEmailDelivery(nil, cfg.FrontendBaseURL, cfg.AuthDebugExposeTokens)
	}
	authController := controller.NewAuthController(authService, cfg.Env == "production", cfg.RefreshTokenTTL, cfg.TrustedProxyCIDRs)

	authRateLimitStore := middleware.NewRedisAuthRateLimitStore(redisClient)
	router := NewRouter(authController, authService, authRateLimitStore, cfg.AuthRateLimitPerMinute, cfg.TrustedProxyCIDRs)
	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{httpServer: httpServer, redisClient: redisClient}, nil
}

func (s *Server) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start http server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	if s.redisClient != nil {
		return s.redisClient.Close()
	}
	return nil
}
