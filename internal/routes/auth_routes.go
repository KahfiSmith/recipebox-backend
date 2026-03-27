package routes

import (
	"net"

	"github.com/go-chi/chi/v5"
	"recipebox-backend-go/internal/controller"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/service"
)

func RegisterAuthRoutes(r chi.Router, authController *controller.AuthController, authService *service.AuthService, authRateLimitStore middleware.AuthRateLimitStore, authRateLimitPerMinute int, trustedProxies []*net.IPNet) {
	authSensitiveRateLimit := middleware.NewAuthRateLimit(authRateLimitStore, authRateLimitPerMinute, trustedProxies)

	r.Route("/auth", func(r chi.Router) {
		r.With(authSensitiveRateLimit).Post("/register", authController.Register)

		r.With(authSensitiveRateLimit).Post("/login", authController.Login)
		r.With(authSensitiveRateLimit).Post("/verify-email/request", authController.RequestEmailVerification)
		r.Get("/verify-email/confirm", authController.VerifyEmailLink)
		r.Post("/verify-email/confirm", authController.VerifyEmail)
		r.With(authSensitiveRateLimit).Post("/password/forgot", authController.ForgotPassword)
		r.Post("/password/reset", authController.ResetPassword)
		r.With(authSensitiveRateLimit).Post("/refresh", authController.Refresh)
		r.Post("/logout", authController.Logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthJWT(authService))
			r.Get("/me", authController.Me)
		})
	})
}
