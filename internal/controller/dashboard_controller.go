package controller

import (
	"net/http"

	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/service"
	"recipebox-backend-go/internal/utils"
)

type DashboardController struct {
	service *service.DashboardService
}

func NewDashboardController(service *service.DashboardService) *DashboardController {
	return &DashboardController{service: service}
}

// GetDashboard godoc
// @Summary Get dashboard overview
// @Description Return recipes, meal plans, shopping items and summary in one payload.
// @Tags Dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.DashboardEnvelope
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/dashboard [get]
func (h *DashboardController) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.service.GetDashboard(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to load dashboard")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// GetRecipes godoc
// @Summary List recipe cards
// @Description Return recipes list used by dashboard/recipes page.
// @Tags Recipes
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.RecipesEnvelope
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/recipes [get]
func (h *DashboardController) GetRecipes(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.service.ListRecipes(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to load recipes")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// GetMealPlans godoc
// @Summary List meal plans
// @Description Return meal plans list used by dashboard/meal-plan page.
// @Tags Meal Plans
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MealPlansEnvelope
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/meal-plans [get]
func (h *DashboardController) GetMealPlans(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.service.ListMealPlans(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to load meal plans")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// GetShoppingItems godoc
// @Summary List shopping items
// @Description Return shopping items list used by dashboard/shopping page.
// @Tags Shopping Items
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ShoppingItemsEnvelope
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/shopping-items [get]
func (h *DashboardController) GetShoppingItems(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.service.ListShoppingItems(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to load shopping items")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

func (h *DashboardController) userIDFromRequest(r *http.Request) (int64, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	return userID, ok
}
