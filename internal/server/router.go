package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"recipebox-backend-go/internal/controller"
	"recipebox-backend-go/internal/routes"
	"recipebox-backend-go/internal/service"
	"recipebox-backend-go/internal/utils"
)

func NewRouter(authController *controller.AuthController, authService *service.AuthService, authRateLimitPerMinute int) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Logger)
	r.Use(chimw.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		routes.RegisterAll(r, authController, authService, authRateLimitPerMinute)
	})

	return r
}
