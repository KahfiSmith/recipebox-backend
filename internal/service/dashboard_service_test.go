package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/models"
	"recipebox-backend-go/internal/repository"
)

func TestDashboardServiceBuildsOverviewPayload(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 3, 8, 0, 0, 0, time.UTC)
	svc := NewDashboardService(fakeRecipeBoxRepository{
		recipes: []models.Recipe{
			{ID: 1, UserID: 42, Name: "Ayam Bakar Madu", Category: "Dinner", PrepTime: 35, UpdatedAt: now},
			{ID: 2, UserID: 42, Name: "Nasi Goreng Kampung", Category: "Breakfast", PrepTime: 15, UpdatedAt: now.Add(-time.Hour)},
			{ID: 3, UserID: 42, Name: "Tumis Brokoli Jamur", Category: "Lunch", PrepTime: 20, UpdatedAt: now.Add(-2 * time.Hour)},
		},
		mealPlans: []models.MealPlan{
			{ID: 10, UserID: 42, Day: "Monday", MealName: "Ayam Bakar Madu", Servings: 2, Ingredients: []string{"Ayam", "Madu"}},
			{ID: 11, UserID: 42, Day: "Tuesday", MealName: "Tumis Brokoli Jamur", Servings: 3, Ingredients: []string{"Brokoli", "Jamur"}},
		},
		shoppingItems: []models.ShoppingItem{
			{ID: 20, UserID: 42, MenuName: "Ayam Bakar Madu", Name: "Dada ayam", Qty: "500 g", Checked: false},
			{ID: 21, UserID: 42, MenuName: "Ayam Bakar Madu", Name: "Madu", Qty: "3 sdm", Checked: false},
			{ID: 22, UserID: 42, MenuName: "Tumis Brokoli Jamur", Name: "Brokoli", Qty: "1 ikat", Checked: true},
			{ID: 23, UserID: 42, MenuName: "Tumis Brokoli Jamur", Name: "Jamur kancing", Qty: "200 g", Checked: false},
		},
	})

	resp, err := svc.GetDashboard(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetDashboard returned error: %v", err)
	}

	if resp.Summary.RecipeCount != len(resp.Recipes) {
		t.Fatalf("expected recipe count to match recipes length")
	}
	if resp.Summary.UpcomingMealPlanCount != len(resp.MealPlans) {
		t.Fatalf("expected meal plan count to match meal plans length")
	}
	if resp.Summary.PendingShoppingItemCount != 3 {
		t.Fatalf("expected 3 pending shopping items, got %d", resp.Summary.PendingShoppingItemCount)
	}
	if len(resp.Recipes) != 3 || len(resp.MealPlans) != 2 || len(resp.ShoppingItems) != 4 {
		t.Fatalf("unexpected dashboard list sizes")
	}
}

func TestDashboardServiceProvidesIndividualMenuPayloads(t *testing.T) {
	t.Parallel()

	repo := fakeRecipeBoxRepository{
		recipes: []models.Recipe{
			{ID: 1, UserID: 42, Name: "Ayam Bakar Madu", Category: "Dinner", PrepTime: 35, UpdatedAt: time.Now().UTC()},
			{ID: 2, UserID: 42, Name: "Nasi Goreng Kampung", Category: "Breakfast", PrepTime: 15, UpdatedAt: time.Now().UTC()},
			{ID: 3, UserID: 42, Name: "Tumis Brokoli Jamur", Category: "Lunch", PrepTime: 20, UpdatedAt: time.Now().UTC()},
		},
		mealPlans: []models.MealPlan{
			{ID: 10, UserID: 42, Day: "Monday", MealName: "Ayam Bakar Madu", Servings: 2, Ingredients: []string{"Ayam", "Madu"}},
			{ID: 11, UserID: 42, Day: "Tuesday", MealName: "Tumis Brokoli Jamur", Servings: 3, Ingredients: []string{"Brokoli", "Jamur"}},
		},
		shoppingItems: []models.ShoppingItem{
			{ID: 20, UserID: 42, MenuName: "Ayam Bakar Madu", Name: "Dada ayam", Qty: "500 g", Checked: false},
			{ID: 21, UserID: 42, MenuName: "Ayam Bakar Madu", Name: "Madu", Qty: "3 sdm", Checked: false},
			{ID: 22, UserID: 42, MenuName: "Tumis Brokoli Jamur", Name: "Brokoli", Qty: "1 ikat", Checked: true},
			{ID: 23, UserID: 42, MenuName: "Tumis Brokoli Jamur", Name: "Jamur kancing", Qty: "200 g", Checked: false},
		},
	}
	svc := NewDashboardService(repo)

	recipes, err := svc.ListRecipes(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListRecipes returned error: %v", err)
	}
	if len(recipes) != 3 {
		t.Fatalf("expected 3 recipes, got %d", len(recipes))
	}

	mealPlans, err := svc.ListMealPlans(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListMealPlans returned error: %v", err)
	}
	if len(mealPlans) != 2 {
		t.Fatalf("expected 2 meal plans, got %d", len(mealPlans))
	}
	if mealPlans[0].Day != "Monday" || mealPlans[0].MealName == "" {
		t.Fatalf("expected transformed meal plan payload")
	}
	if len(mealPlans[0].Ingredients) != 2 {
		t.Fatalf("expected ingredients array to be preserved")
	}

	shoppingItems, err := svc.ListShoppingItems(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListShoppingItems returned error: %v", err)
	}
	if len(shoppingItems) != 4 {
		t.Fatalf("expected 4 shopping items, got %d", len(shoppingItems))
	}
	if shoppingItems[0].MenuName == "" || shoppingItems[0].Qty == "" {
		t.Fatalf("expected shopping items to expose menuName and qty")
	}
}

func TestDashboardServicePropagatesRepositoryErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	svc := NewDashboardService(fakeRecipeBoxRepository{err: expectedErr})

	_, err := svc.GetDashboard(context.Background(), 42)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

type fakeRecipeBoxRepository struct {
	recipes       []models.Recipe
	mealPlans     []models.MealPlan
	shoppingItems []models.ShoppingItem
	err           error
}

func (f fakeRecipeBoxRepository) ListRecipes(_ context.Context, _ int64) ([]models.Recipe, error) {
	if f.err != nil {
		return nil, f.err
	}
	return cloneRecipes(f.recipes), nil
}

func TestDashboardServiceCreateRecipeValidatesInput(t *testing.T) {
	t.Parallel()

	svc := NewDashboardService(fakeRecipeBoxRepository{})
	_, err := svc.CreateRecipe(context.Background(), 42, dto.UpsertRecipeRequest{
		Name:     "",
		Category: "Dinner",
		PrepTime: 10,
	})
	if !models.IsValidationError(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func (f fakeRecipeBoxRepository) CreateRecipe(_ context.Context, userID int64, recipe models.Recipe) (models.Recipe, error) {
	if f.err != nil {
		return models.Recipe{}, f.err
	}
	recipe.ID = 1001
	recipe.UserID = userID
	return recipe, nil
}

func (f fakeRecipeBoxRepository) UpdateRecipe(_ context.Context, userID, recipeID int64, recipe models.Recipe) (models.Recipe, error) {
	if f.err != nil {
		return models.Recipe{}, f.err
	}
	recipe.ID = recipeID
	recipe.UserID = userID
	return recipe, nil
}

func (f fakeRecipeBoxRepository) DeleteRecipe(_ context.Context, _, _ int64) error {
	return f.err
}

func (f fakeRecipeBoxRepository) ListMealPlans(_ context.Context, _ int64) ([]models.MealPlan, error) {
	if f.err != nil {
		return nil, f.err
	}
	return cloneMealPlans(f.mealPlans), nil
}

func (f fakeRecipeBoxRepository) CreateMealPlan(_ context.Context, userID int64, mealPlan models.MealPlan) (models.MealPlan, error) {
	if f.err != nil {
		return models.MealPlan{}, f.err
	}
	mealPlan.ID = 2001
	mealPlan.UserID = userID
	return mealPlan, nil
}

func (f fakeRecipeBoxRepository) UpdateMealPlan(_ context.Context, userID, mealPlanID int64, mealPlan models.MealPlan) (models.MealPlan, error) {
	if f.err != nil {
		return models.MealPlan{}, f.err
	}
	mealPlan.ID = mealPlanID
	mealPlan.UserID = userID
	return mealPlan, nil
}

func (f fakeRecipeBoxRepository) DeleteMealPlan(_ context.Context, _, _ int64) error {
	return f.err
}

func (f fakeRecipeBoxRepository) ListShoppingItems(_ context.Context, _ int64) ([]models.ShoppingItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return cloneShoppingItems(f.shoppingItems), nil
}

func (f fakeRecipeBoxRepository) CreateShoppingItem(_ context.Context, userID int64, item models.ShoppingItem) (models.ShoppingItem, error) {
	if f.err != nil {
		return models.ShoppingItem{}, f.err
	}
	item.ID = 3001
	item.UserID = userID
	return item, nil
}

func (f fakeRecipeBoxRepository) UpdateShoppingItem(_ context.Context, userID, itemID int64, item models.ShoppingItem) (models.ShoppingItem, error) {
	if f.err != nil {
		return models.ShoppingItem{}, f.err
	}
	item.ID = itemID
	item.UserID = userID
	return item, nil
}

func (f fakeRecipeBoxRepository) DeleteShoppingItem(_ context.Context, _, _ int64) error {
	return f.err
}

func cloneRecipes(in []models.Recipe) []models.Recipe {
	if in == nil {
		return nil
	}
	out := make([]models.Recipe, len(in))
	copy(out, in)
	return out
}

func cloneMealPlans(in []models.MealPlan) []models.MealPlan {
	if in == nil {
		return nil
	}
	out := make([]models.MealPlan, len(in))
	copy(out, in)
	return out
}

func cloneShoppingItems(in []models.ShoppingItem) []models.ShoppingItem {
	if in == nil {
		return nil
	}
	out := make([]models.ShoppingItem, len(in))
	copy(out, in)
	return out
}

var _ repository.RecipeBoxRepository = fakeRecipeBoxRepository{}
