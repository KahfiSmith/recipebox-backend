package dto

import "recipebox-backend-go/internal/models"

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type RegisterEnvelope struct {
	Data RegisterResponse `json:"data"`
}

type AuthEnvelope struct {
	Data AuthResponse `json:"data"`
}

type TokenEnvelope struct {
	Data TokenPair `json:"data"`
}

type UserEnvelope struct {
	Data models.User `json:"data"`
}

type OneTimeTokenEnvelope struct {
	Message string               `json:"message"`
	Data    OneTimeTokenResponse `json:"data,omitempty"`
}

type DashboardEnvelope struct {
	Data DashboardResponse `json:"data"`
}

type RecipesEnvelope struct {
	Data []DashboardRecipe `json:"data"`
}

type MealPlansEnvelope struct {
	Data []DashboardMealPlan `json:"data"`
}

type ShoppingItemsEnvelope struct {
	Data []DashboardShoppingItem `json:"data"`
}
