package service

import (
	"context"
	"errors"
	"strings"

	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/models"
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
		result = append(result, toDashboardRecipe(recipe))
	}

	return result, nil
}

func (s *DashboardService) CreateRecipe(ctx context.Context, userID int64, input dto.UpsertRecipeRequest) (dto.DashboardRecipe, error) {
	if s.repo == nil {
		return dto.DashboardRecipe{}, errors.New("recipebox repository is not configured")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return dto.DashboardRecipe{}, models.ValidationError{Message: "name is required"}
	}

	if input.PrepTime < 0 {
		return dto.DashboardRecipe{}, models.ValidationError{Message: "prepTime must be >= 0"}
	}

	created, err := s.repo.CreateRecipe(ctx, userID, models.Recipe{
		Name:     name,
		Category: strings.TrimSpace(input.Category),
		PrepTime: input.PrepTime,
	})
	if err != nil {
		return dto.DashboardRecipe{}, err
	}

	return toDashboardRecipe(created), nil
}

func (s *DashboardService) UpdateRecipe(ctx context.Context, userID, recipeID int64, input dto.UpsertRecipeRequest) (dto.DashboardRecipe, error) {
	if s.repo == nil {
		return dto.DashboardRecipe{}, errors.New("recipebox repository is not configured")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return dto.DashboardRecipe{}, models.ValidationError{Message: "name is required"}
	}
	if input.PrepTime < 0 {
		return dto.DashboardRecipe{}, models.ValidationError{Message: "prepTime must be >= 0"}
	}

	updated, err := s.repo.UpdateRecipe(ctx, userID, recipeID, models.Recipe{
		Name:     name,
		Category: strings.TrimSpace(input.Category),
		PrepTime: input.PrepTime,
	})
	if err != nil {
		return dto.DashboardRecipe{}, err
	}

	return toDashboardRecipe(updated), nil
}

func (s *DashboardService) DeleteRecipe(ctx context.Context, userID, recipeID int64) error {
	if s.repo == nil {
		return errors.New("recipebox repository is not configured")
	}
	return s.repo.DeleteRecipe(ctx, userID, recipeID)
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
		result = append(result, toDashboardMealPlan(mealPlan))
	}

	return result, nil
}

func (s *DashboardService) CreateMealPlan(ctx context.Context, userID int64, input dto.UpsertMealPlanRequest) (dto.DashboardMealPlan, error) {
	if s.repo == nil {
		return dto.DashboardMealPlan{}, errors.New("recipebox repository is not configured")
	}

	day := strings.TrimSpace(input.Day)
	mealName := strings.TrimSpace(input.MealName)
	if day == "" {
		return dto.DashboardMealPlan{}, models.ValidationError{Message: "day is required"}
	}
	if mealName == "" {
		return dto.DashboardMealPlan{}, models.ValidationError{Message: "mealName is required"}
	}
	if input.Servings < 1 {
		return dto.DashboardMealPlan{}, models.ValidationError{Message: "servings must be >= 1"}
	}

	created, err := s.repo.CreateMealPlan(ctx, userID, models.MealPlan{
		Day:         day,
		MealName:    mealName,
		Servings:    input.Servings,
		Ingredients: compactStrings(input.Ingredients),
	})
	if err != nil {
		return dto.DashboardMealPlan{}, err
	}

	return toDashboardMealPlan(created), nil
}

func (s *DashboardService) UpdateMealPlan(ctx context.Context, userID, mealPlanID int64, input dto.UpsertMealPlanRequest) (dto.DashboardMealPlan, error) {
	if s.repo == nil {
		return dto.DashboardMealPlan{}, errors.New("recipebox repository is not configured")
	}

	day := strings.TrimSpace(input.Day)
	mealName := strings.TrimSpace(input.MealName)
	if day == "" {
		return dto.DashboardMealPlan{}, models.ValidationError{Message: "day is required"}
	}
	if mealName == "" {
		return dto.DashboardMealPlan{}, models.ValidationError{Message: "mealName is required"}
	}
	if input.Servings < 1 {
		return dto.DashboardMealPlan{}, models.ValidationError{Message: "servings must be >= 1"}
	}

	updated, err := s.repo.UpdateMealPlan(ctx, userID, mealPlanID, models.MealPlan{
		Day:         day,
		MealName:    mealName,
		Servings:    input.Servings,
		Ingredients: compactStrings(input.Ingredients),
	})
	if err != nil {
		return dto.DashboardMealPlan{}, err
	}

	return toDashboardMealPlan(updated), nil
}

func (s *DashboardService) DeleteMealPlan(ctx context.Context, userID, mealPlanID int64) error {
	if s.repo == nil {
		return errors.New("recipebox repository is not configured")
	}
	return s.repo.DeleteMealPlan(ctx, userID, mealPlanID)
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
		result = append(result, toDashboardShoppingItem(item))
	}

	return result, nil
}

func (s *DashboardService) CreateShoppingItem(ctx context.Context, userID int64, input dto.UpsertShoppingItemRequest) (dto.DashboardShoppingItem, error) {
	if s.repo == nil {
		return dto.DashboardShoppingItem{}, errors.New("recipebox repository is not configured")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return dto.DashboardShoppingItem{}, models.ValidationError{Message: "name is required"}
	}

	created, err := s.repo.CreateShoppingItem(ctx, userID, models.ShoppingItem{
		MenuName: strings.TrimSpace(input.MenuName),
		Name:     name,
		Qty:      strings.TrimSpace(input.Qty),
		Checked:  input.Checked,
	})
	if err != nil {
		return dto.DashboardShoppingItem{}, err
	}

	return toDashboardShoppingItem(created), nil
}

func (s *DashboardService) UpdateShoppingItem(ctx context.Context, userID, itemID int64, input dto.UpsertShoppingItemRequest) (dto.DashboardShoppingItem, error) {
	if s.repo == nil {
		return dto.DashboardShoppingItem{}, errors.New("recipebox repository is not configured")
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return dto.DashboardShoppingItem{}, models.ValidationError{Message: "name is required"}
	}

	updated, err := s.repo.UpdateShoppingItem(ctx, userID, itemID, models.ShoppingItem{
		MenuName: strings.TrimSpace(input.MenuName),
		Name:     name,
		Qty:      strings.TrimSpace(input.Qty),
		Checked:  input.Checked,
	})
	if err != nil {
		return dto.DashboardShoppingItem{}, err
	}

	return toDashboardShoppingItem(updated), nil
}

func (s *DashboardService) DeleteShoppingItem(ctx context.Context, userID, itemID int64) error {
	if s.repo == nil {
		return errors.New("recipebox repository is not configured")
	}
	return s.repo.DeleteShoppingItem(ctx, userID, itemID)
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

func toDashboardRecipe(recipe models.Recipe) dto.DashboardRecipe {
	return dto.DashboardRecipe{
		ID:        recipe.ID,
		Name:      recipe.Name,
		Category:  recipe.Category,
		PrepTime:  recipe.PrepTime,
		UpdatedAt: recipe.UpdatedAt,
	}
}

func toDashboardMealPlan(mealPlan models.MealPlan) dto.DashboardMealPlan {
	return dto.DashboardMealPlan{
		ID:          mealPlan.ID,
		Day:         mealPlan.Day,
		MealName:    mealPlan.MealName,
		Servings:    mealPlan.Servings,
		Ingredients: append([]string(nil), mealPlan.Ingredients...),
	}
}

func toDashboardShoppingItem(item models.ShoppingItem) dto.DashboardShoppingItem {
	return dto.DashboardShoppingItem{
		ID:       item.ID,
		MenuName: item.MenuName,
		Name:     item.Name,
		Qty:      item.Qty,
		Checked:  item.Checked,
	}
}

func compactStrings(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}

	out := make([]string, 0, len(in))
	for _, s := range in {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
