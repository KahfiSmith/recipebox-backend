package repository

import (
	"context"

	"recipebox-backend-go/internal/models"
)

type RecipeBoxRepository interface {
	ListRecipes(ctx context.Context, userID int64) ([]models.Recipe, error)
	CreateRecipe(ctx context.Context, userID int64, recipe models.Recipe) (models.Recipe, error)
	UpdateRecipe(ctx context.Context, userID, recipeID int64, recipe models.Recipe) (models.Recipe, error)
	DeleteRecipe(ctx context.Context, userID, recipeID int64) error

	ListMealPlans(ctx context.Context, userID int64) ([]models.MealPlan, error)
	CreateMealPlan(ctx context.Context, userID int64, mealPlan models.MealPlan) (models.MealPlan, error)
	UpdateMealPlan(ctx context.Context, userID, mealPlanID int64, mealPlan models.MealPlan) (models.MealPlan, error)
	DeleteMealPlan(ctx context.Context, userID, mealPlanID int64) error

	ListShoppingItems(ctx context.Context, userID int64) ([]models.ShoppingItem, error)
	CreateShoppingItem(ctx context.Context, userID int64, item models.ShoppingItem) (models.ShoppingItem, error)
	UpdateShoppingItem(ctx context.Context, userID, itemID int64, item models.ShoppingItem) (models.ShoppingItem, error)
	DeleteShoppingItem(ctx context.Context, userID, itemID int64) error
}
