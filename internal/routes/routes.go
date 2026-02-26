package routes

import (
	"github.com/go-chi/chi/v5"
	"recipebox-backend-go/internal/controller"
	"recipebox-backend-go/internal/service"
)

func RegisterAll(r chi.Router, authController *controller.AuthController, authService *service.AuthService, authRateLimitPerMinute int) {
	RegisterAuthRoutes(r, authController, authService, authRateLimitPerMinute)
}
