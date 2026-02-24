package routes

import (
	"github.com/go-chi/chi/v5"
	"recipebox-backend-go/internal/controller"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/service"
)

func RegisterAuthRoutes(r chi.Router, authController *controller.AuthController, authService *service.AuthService) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authController.Register)
		r.Post("/login", authController.Login)
		r.Post("/verify-email/request", authController.RequestEmailVerification)
		r.Post("/verify-email/confirm", authController.VerifyEmail)
		r.Post("/password/forgot", authController.ForgotPassword)
		r.Post("/password/reset", authController.ResetPassword)
		r.Post("/refresh", authController.Refresh)
		r.Post("/logout", authController.Logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthJWT(authService))
			r.Get("/me", authController.Me)
		})
	})
}
