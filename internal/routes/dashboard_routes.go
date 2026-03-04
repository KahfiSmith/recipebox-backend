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
		r.Post("/recipes", dashboardController.CreateRecipe)
		r.Put("/recipes/{id}", dashboardController.UpdateRecipe)
		r.Delete("/recipes/{id}", dashboardController.DeleteRecipe)

		r.Get("/meal-plans", dashboardController.GetMealPlans)
		r.Post("/meal-plans", dashboardController.CreateMealPlan)
		r.Put("/meal-plans/{id}", dashboardController.UpdateMealPlan)
		r.Delete("/meal-plans/{id}", dashboardController.DeleteMealPlan)

		r.Get("/shopping-items", dashboardController.GetShoppingItems)
		r.Post("/shopping-items", dashboardController.CreateShoppingItem)
		r.Put("/shopping-items/{id}", dashboardController.UpdateShoppingItem)
		r.Delete("/shopping-items/{id}", dashboardController.DeleteShoppingItem)
	})
}
