package repository

import (
	"context"

	"recipebox-backend-go/internal/models"
)

type RecipeBoxRepository interface {
	ListRecipes(ctx context.Context, userID int64) ([]models.Recipe, error)
	ListMealPlans(ctx context.Context, userID int64) ([]models.MealPlan, error)
	ListShoppingItems(ctx context.Context, userID int64) ([]models.ShoppingItem, error)
}
