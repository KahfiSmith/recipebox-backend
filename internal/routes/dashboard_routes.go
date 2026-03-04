package routes

import (
	"github.com/go-chi/chi/v5"
	"recipebox-backend-go/internal/controller"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/service"
)

func RegisterDashboardRoutes(r chi.Router, dashboardController *controller.DashboardController, authService *service.AuthService) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthJWT(authService))
		r.Get("/dashboard", dashboardController.GetDashboard)
		r.Get("/recipes", dashboardController.GetRecipes)
		r.Get("/meal-plans", dashboardController.GetMealPlans)
		r.Get("/shopping-items", dashboardController.GetShoppingItems)
	})
}
