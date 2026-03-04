package dto

import "time"

type DashboardRecipe struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	PrepTime  int       `json:"prepTime"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DashboardMealPlan struct {
	ID          int64    `json:"id"`
	Day         string   `json:"day"`
	MealName    string   `json:"mealName"`
	Servings    int      `json:"servings"`
	Ingredients []string `json:"ingredients"`
}

type DashboardShoppingItem struct {
	ID       int64  `json:"id"`
	MenuName string `json:"menuName"`
	Name     string `json:"name"`
	Qty      string `json:"qty"`
	Checked  bool   `json:"checked"`
}

type DashboardSummary struct {
	RecipeCount              int `json:"recipeCount"`
	UpcomingMealPlanCount    int `json:"upcomingMealPlanCount"`
	PendingShoppingItemCount int `json:"pendingShoppingItemCount"`
}

type DashboardResponse struct {
	Summary       DashboardSummary        `json:"summary"`
	Recipes       []DashboardRecipe       `json:"recipes"`
	MealPlans     []DashboardMealPlan     `json:"mealPlans"`
	ShoppingItems []DashboardShoppingItem `json:"shoppingItems"`
}