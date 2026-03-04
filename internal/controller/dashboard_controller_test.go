package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"recipebox-backend-go/internal/models"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/repository"
	"recipebox-backend-go/internal/service"
)

func TestDashboardRequiresAuthentication(t *testing.T) {
	t.Parallel()

	dashboardController := NewDashboardController(service.NewDashboardService(newDashboardTestRepository()))
	authService := service.NewAuthService(nil, strings.Repeat("a", 32), 15*time.Minute, 24*time.Hour, 10)
	handler := middleware.AuthJWT(authService)(http.HandlerFunc(dashboardController.GetDashboard))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestDashboardReturnsRecipesMealPlansAndShoppingItems(t *testing.T) {
	t.Parallel()

	secret := strings.Repeat("a", 32)
	dashboardService := service.NewDashboardService(newDashboardTestRepository())
	dashboardController := NewDashboardController(dashboardService)
	authService := service.NewAuthService(nil, secret, 15*time.Minute, 24*time.Hour, 10)
	handler := middleware.AuthJWT(authService)(http.HandlerFunc(dashboardController.GetDashboard))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+makeDashboardToken(t, secret, 42))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Data struct {
			Summary struct {
				RecipeCount              int `json:"recipeCount"`
				UpcomingMealPlanCount    int `json:"upcomingMealPlanCount"`
				PendingShoppingItemCount int `json:"pendingShoppingItemCount"`
			} `json:"summary"`
			Recipes       []map[string]any `json:"recipes"`
			MealPlans     []map[string]any `json:"mealPlans"`
			ShoppingItems []map[string]any `json:"shoppingItems"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Data.Summary.RecipeCount != 3 {
		t.Fatalf("expected 3 recipes, got %d", payload.Data.Summary.RecipeCount)
	}
	if payload.Data.Summary.UpcomingMealPlanCount != 2 {
		t.Fatalf("expected 2 meal plans, got %d", payload.Data.Summary.UpcomingMealPlanCount)
	}
	if payload.Data.Summary.PendingShoppingItemCount != 3 {
		t.Fatalf("expected 3 pending shopping items, got %d", payload.Data.Summary.PendingShoppingItemCount)
	}
	if len(payload.Data.Recipes) == 0 || len(payload.Data.MealPlans) == 0 || len(payload.Data.ShoppingItems) == 0 {
		t.Fatalf("expected dashboard lists to be populated")
	}
}

func TestDashboardMenuEndpointsReturnData(t *testing.T) {
	t.Parallel()

	secret := strings.Repeat("a", 32)
	dashboardService := service.NewDashboardService(newDashboardTestRepository())
	dashboardController := NewDashboardController(dashboardService)
	authService := service.NewAuthService(nil, secret, 15*time.Minute, 24*time.Hour, 10)
	token := "Bearer " + makeDashboardToken(t, secret, 42)

	tests := []struct {
		name string
		path string
		handler http.HandlerFunc
	}{
		{name: "recipes", path: "/api/v1/recipes", handler: dashboardController.GetRecipes},
		{name: "meal plans", path: "/api/v1/meal-plans", handler: dashboardController.GetMealPlans},
		{name: "shopping items", path: "/api/v1/shopping-items", handler: dashboardController.GetShoppingItems},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := middleware.AuthJWT(authService)(tc.handler)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", token)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", rec.Code)
			}

			var payload struct {
				Data []map[string]any `json:"data"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(payload.Data) == 0 {
				t.Fatalf("expected populated list response")
			}
		})
	}
}

func makeDashboardToken(t *testing.T, secret string, userID int64) string {
	t.Helper()

	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Issuer:    "recipebox-api",
		Subject:   fmt.Sprintf("%d", userID),
		Audience:  []string{"recipebox-client"},
		ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        "acc_dashboard-test-token-id",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func newDashboardTestRepository() repository.RecipeBoxRepository {
	now := time.Date(2026, 3, 3, 8, 0, 0, 0, time.UTC)
	return dashboardTestRepository{
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
	}
}

type dashboardTestRepository struct {
	recipes       []models.Recipe
	mealPlans     []models.MealPlan
	shoppingItems []models.ShoppingItem
}

func (r dashboardTestRepository) ListRecipes(_ context.Context, _ int64) ([]models.Recipe, error) {
	out := make([]models.Recipe, len(r.recipes))
	copy(out, r.recipes)
	return out, nil
}

func (r dashboardTestRepository) ListMealPlans(_ context.Context, _ int64) ([]models.MealPlan, error) {
	out := make([]models.MealPlan, len(r.mealPlans))
	copy(out, r.mealPlans)
	return out, nil
}

func (r dashboardTestRepository) ListShoppingItems(_ context.Context, _ int64) ([]models.ShoppingItem, error) {
	out := make([]models.ShoppingItem, len(r.shoppingItems))
	copy(out, r.shoppingItems)
	return out, nil
}
