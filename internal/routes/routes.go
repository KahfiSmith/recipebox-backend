package routes

import (
	"net"

	"github.com/go-chi/chi/v5"
	"recipebox-backend-go/internal/controller"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/service"
)

func RegisterAll(r chi.Router, authController *controller.AuthController, dashboardController *controller.DashboardController, authService *service.AuthService, authRateLimitStore middleware.AuthRateLimitStore, authRateLimitPerMinute int, trustedProxies []*net.IPNet) {
	RegisterAuthRoutes(r, authController, authService, authRateLimitStore, authRateLimitPerMinute, trustedProxies)
	RegisterDashboardRoutes(r, dashboardController, authService)
}
