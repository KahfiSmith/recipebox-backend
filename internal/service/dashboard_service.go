package service

import (
	"context"
	"errors"

	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/repository"
)

type DashboardService struct {
	repo repository.RecipeBoxRepository
}

func NewDashboardService(repo repository.RecipeBoxRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) ListRecipes(ctx context.Context, userID int64) ([]dto.DashboardRecipe, error) {
	if s.repo == nil {
		return nil, errors.New("recipebox repository is not configured")
	}

	recipes, err := s.repo.ListRecipes(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.DashboardRecipe, 0, len(recipes))
	for _, recipe := range recipes {
		result = append(result, dto.DashboardRecipe{
			ID:        recipe.ID,
			Name:      recipe.Name,
			Category:  recipe.Category,
			PrepTime:  recipe.PrepTime,
			UpdatedAt: recipe.UpdatedAt,
		})
	}

	return result, nil
}

func (s *DashboardService) ListMealPlans(ctx context.Context, userID int64) ([]dto.DashboardMealPlan, error) {
	if s.repo == nil {
		return nil, errors.New("recipebox repository is not configured")
	}

	mealPlans, err := s.repo.ListMealPlans(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.DashboardMealPlan, 0, len(mealPlans))
	for _, mealPlan := range mealPlans {
		result = append(result, dto.DashboardMealPlan{
			ID:          mealPlan.ID,
			Day:         mealPlan.Day,
			MealName:    mealPlan.MealName,
			Servings:    mealPlan.Servings,
			Ingredients: append([]string(nil), mealPlan.Ingredients...),
		})
	}

	return result, nil
}

func (s *DashboardService) ListShoppingItems(ctx context.Context, userID int64) ([]dto.DashboardShoppingItem, error) {
	if s.repo == nil {
		return nil, errors.New("recipebox repository is not configured")
	}

	items, err := s.repo.ListShoppingItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.DashboardShoppingItem, 0, len(items))
	for _, item := range items {
		result = append(result, dto.DashboardShoppingItem{
			ID:       item.ID,
			MenuName: item.MenuName,
			Name:     item.Name,
			Qty:      item.Qty,
			Checked:  item.Checked,
		})
	}

	return result, nil
}

func (s *DashboardService) GetDashboard(ctx context.Context, userID int64) (dto.DashboardResponse, error) {
	recipes, err := s.ListRecipes(ctx, userID)
	if err != nil {
		return dto.DashboardResponse{}, err
	}

	mealPlans, err := s.ListMealPlans(ctx, userID)
	if err != nil {
		return dto.DashboardResponse{}, err
	}

	shoppingItems, err := s.ListShoppingItems(ctx, userID)
	if err != nil {
		return dto.DashboardResponse{}, err
	}

	pendingCount := 0
	for _, item := range shoppingItems {
		if !item.Checked {
			pendingCount++
		}
	}

	return dto.DashboardResponse{
		Summary: dto.DashboardSummary{
			RecipeCount:              len(recipes),
			UpcomingMealPlanCount:    len(mealPlans),
			PendingShoppingItemCount: pendingCount,
		},
		Recipes:       recipes,
		MealPlans:     mealPlans,
		ShoppingItems: shoppingItems,
	}, nil
}
