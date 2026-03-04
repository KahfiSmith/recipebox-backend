package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"recipebox-backend-go/internal/dto"
	"recipebox-backend-go/internal/middleware"
	"recipebox-backend-go/internal/models"
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

// CreateRecipe godoc
// @Summary Create recipe
// @Description Create a recipe menu item.
// @Tags Recipes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.UpsertRecipeRequest true "Recipe payload"
// @Success 201 {object} dto.RecipeEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/recipes [post]
func (h *DashboardController) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input dto.UpsertRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.CreateRecipe(r.Context(), userID, input)
	if err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]any{"data": resp})
}

// UpdateRecipe godoc
// @Summary Update recipe
// @Description Update a recipe menu item by ID.
// @Tags Recipes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Recipe ID"
// @Param payload body dto.UpsertRecipeRequest true "Recipe payload"
// @Success 200 {object} dto.RecipeEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/recipes/{id} [put]
func (h *DashboardController) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	recipeID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	var input dto.UpsertRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.UpdateRecipe(r.Context(), userID, recipeID, input)
	if err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// DeleteRecipe godoc
// @Summary Delete recipe
// @Description Delete a recipe menu item by ID.
// @Tags Recipes
// @Produce json
// @Security BearerAuth
// @Param id path int true "Recipe ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/recipes/{id} [delete]
func (h *DashboardController) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	recipeID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteRecipe(r.Context(), userID, recipeID); err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"message": "recipe deleted"})
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

// CreateMealPlan godoc
// @Summary Create meal plan
// @Description Create a meal plan menu item.
// @Tags Meal Plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.UpsertMealPlanRequest true "Meal plan payload"
// @Success 201 {object} dto.MealPlanEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/meal-plans [post]
func (h *DashboardController) CreateMealPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input dto.UpsertMealPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.CreateMealPlan(r.Context(), userID, input)
	if err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]any{"data": resp})
}

// UpdateMealPlan godoc
// @Summary Update meal plan
// @Description Update a meal plan menu item by ID.
// @Tags Meal Plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Meal Plan ID"
// @Param payload body dto.UpsertMealPlanRequest true "Meal plan payload"
// @Success 200 {object} dto.MealPlanEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/meal-plans/{id} [put]
func (h *DashboardController) UpdateMealPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mealPlanID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	var input dto.UpsertMealPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.UpdateMealPlan(r.Context(), userID, mealPlanID, input)
	if err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// DeleteMealPlan godoc
// @Summary Delete meal plan
// @Description Delete a meal plan menu item by ID.
// @Tags Meal Plans
// @Produce json
// @Security BearerAuth
// @Param id path int true "Meal Plan ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/meal-plans/{id} [delete]
func (h *DashboardController) DeleteMealPlan(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mealPlanID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteMealPlan(r.Context(), userID, mealPlanID); err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"message": "meal plan deleted"})
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

// CreateShoppingItem godoc
// @Summary Create shopping item
// @Description Create a shopping item menu entry.
// @Tags Shopping Items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body dto.UpsertShoppingItemRequest true "Shopping item payload"
// @Success 201 {object} dto.ShoppingItemEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/shopping-items [post]
func (h *DashboardController) CreateShoppingItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input dto.UpsertShoppingItemRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.CreateShoppingItem(r.Context(), userID, input)
	if err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]any{"data": resp})
}

// UpdateShoppingItem godoc
// @Summary Update shopping item
// @Description Update a shopping item menu entry by ID.
// @Tags Shopping Items
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Shopping Item ID"
// @Param payload body dto.UpsertShoppingItemRequest true "Shopping item payload"
// @Success 200 {object} dto.ShoppingItemEnvelope
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/shopping-items/{id} [put]
func (h *DashboardController) UpdateShoppingItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	var input dto.UpsertShoppingItemRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.UpdateShoppingItem(r.Context(), userID, itemID, input)
	if err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"data": resp})
}

// DeleteShoppingItem godoc
// @Summary Delete shopping item
// @Description Delete a shopping item menu entry by ID.
// @Tags Shopping Items
// @Produce json
// @Security BearerAuth
// @Param id path int true "Shopping Item ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/shopping-items/{id} [delete]
func (h *DashboardController) DeleteShoppingItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userIDFromRequest(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteShoppingItem(r.Context(), userID, itemID); err != nil {
		h.handleDashboardMutationError(w, err)
		return
	}

	utils.JSON(w, http.StatusOK, map[string]any{"message": "shopping item deleted"})
}

func (h *DashboardController) userIDFromRequest(r *http.Request) (int64, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	return userID, ok
}

func (h *DashboardController) handleDashboardMutationError(w http.ResponseWriter, err error) {
	switch {
	case models.IsValidationError(err):
		utils.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, models.ErrNotFound):
		utils.Error(w, http.StatusNotFound, "resource not found")
	default:
		utils.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func pathID(w http.ResponseWriter, r *http.Request, param string) (int64, bool) {
	raw := chi.URLParam(r, param)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		utils.Error(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
