package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"recipebox-backend-go/internal/models"
	"gorm.io/gorm"
)

type RecipeBoxGormRepository struct {
	db *gorm.DB
}

func NewRecipeBoxGormRepository(db *gorm.DB) *RecipeBoxGormRepository {
	return &RecipeBoxGormRepository{db: db}
}

func (r *RecipeBoxGormRepository) ListRecipes(ctx context.Context, userID int64) ([]models.Recipe, error) {
	return r.ListRecipesPage(ctx, userID, 0, 0)
}

func (r *RecipeBoxGormRepository) ListRecipesPage(ctx context.Context, userID int64, limit, offset int) ([]models.Recipe, error) {
	type recipeRow struct {
		ID        int64     `gorm:"column:id"`
		Name      string    `gorm:"column:name"`
		Category  string    `gorm:"column:category"`
		PrepTime  int       `gorm:"column:prep_time"`
		UpdatedAt time.Time `gorm:"column:updated_at"`
	}

	nameColumn := "name"
	if !r.hasColumn(&models.Recipe{}, "name") && r.hasColumn(&models.Recipe{}, "title") {
		nameColumn = "title"
	}

	orderColumn := "updated_at"
	if !r.hasColumn(&models.Recipe{}, "updated_at") && r.hasColumn(&models.Recipe{}, "created_at") {
		orderColumn = "created_at"
	}

	selectExpr := []string{
		"id",
		fmt.Sprintf("COALESCE(%s, '') AS name", nameColumn),
		"'' AS category",
		"0 AS prep_time",
		fmt.Sprintf("%s AS updated_at", orderColumn),
	}
	if r.hasColumn(&models.Recipe{}, "category") {
		selectExpr[2] = "COALESCE(category, '') AS category"
	}
	if r.hasColumn(&models.Recipe{}, "prep_time") {
		selectExpr[3] = "COALESCE(prep_time, 0) AS prep_time"
	}

	var rows []recipeRow
	query := r.db.WithContext(ctx).
		Table("recipes").
		Select(selectExpr).
		Where("user_id = ?", userID).
		Order(orderColumn + " DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list recipes: %w", err)
	}

	recipes := make([]models.Recipe, 0, len(rows))
	for _, row := range rows {
		recipes = append(recipes, models.Recipe{
			ID:        row.ID,
			UserID:    userID,
			Name:      row.Name,
			Category:  row.Category,
			PrepTime:  row.PrepTime,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return recipes, nil
}

func (r *RecipeBoxGormRepository) CreateRecipe(ctx context.Context, userID int64, recipe models.Recipe) (models.Recipe, error) {
	recipe.ID = 0
	recipe.UserID = userID
	if err := r.db.WithContext(ctx).Create(&recipe).Error; err != nil {
		return models.Recipe{}, fmt.Errorf("create recipe: %w", err)
	}
	return recipe, nil
}

func (r *RecipeBoxGormRepository) UpdateRecipe(ctx context.Context, userID, recipeID int64, recipe models.Recipe) (models.Recipe, error) {
	updates := map[string]any{
		"name":       recipe.Name,
		"category":   recipe.Category,
		"prep_time":  recipe.PrepTime,
		"updated_at": time.Now().UTC(),
	}
	result := r.db.WithContext(ctx).
		Model(&models.Recipe{}).
		Where("id = ? AND user_id = ?", recipeID, userID).
		Updates(updates)
	if result.Error != nil {
		return models.Recipe{}, fmt.Errorf("update recipe: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.Recipe{}, models.ErrNotFound
	}

	var out models.Recipe
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", recipeID, userID).Take(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Recipe{}, models.ErrNotFound
		}
		return models.Recipe{}, fmt.Errorf("fetch updated recipe: %w", err)
	}
	return out, nil
}

func (r *RecipeBoxGormRepository) DeleteRecipe(ctx context.Context, userID, recipeID int64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", recipeID, userID).Delete(&models.Recipe{})
	if result.Error != nil {
		return fmt.Errorf("delete recipe: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *RecipeBoxGormRepository) ListMealPlans(ctx context.Context, userID int64) ([]models.MealPlan, error) {
	return r.ListMealPlansPage(ctx, userID, 0, 0)
}

func (r *RecipeBoxGormRepository) ListMealPlansPage(ctx context.Context, userID int64, limit, offset int) ([]models.MealPlan, error) {
	if r.hasColumn(&models.MealPlan{}, "day") {
		var rows []models.MealPlan
		query := r.db.WithContext(ctx).
			Table("meal_plans").
			Select("id, COALESCE(day, '') AS day, COALESCE(meal_name, '') AS meal_name, COALESCE(servings, 0) AS servings, ingredients").
			Where("user_id = ?", userID).
			Order("id DESC")
		if limit > 0 {
			query = query.Limit(limit)
		}
		if offset > 0 {
			query = query.Offset(offset)
		}
		if err := query.Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("list meal plans: %w", err)
		}

		result := make([]models.MealPlan, 0, len(rows))
		for _, row := range rows {
			result = append(result, models.MealPlan{
				ID:          row.ID,
				UserID:      userID,
				Day:         row.Day,
				MealName:    row.MealName,
				Servings:    row.Servings,
				Ingredients: append([]string(nil), row.Ingredients...),
			})
		}
		return result, nil
	}

	type legacyMealPlanRow struct {
		ID          int64      `gorm:"column:id"`
		ScheduledAt *time.Time `gorm:"column:scheduled_at"`
		MealType    string     `gorm:"column:meal_type"`
		RecipeName  string     `gorm:"column:recipe_name"`
	}

	recipeNameColumn := "r.name"
	if !r.hasColumn(&models.Recipe{}, "name") && r.hasColumn(&models.Recipe{}, "title") {
		recipeNameColumn = "r.title"
	}

	var rows []legacyMealPlanRow
	query := r.db.WithContext(ctx).
		Table("meal_plans AS mp").
		Select([]string{
			"mp.id",
			"mp.scheduled_at",
			"COALESCE(mp.meal_type, '') AS meal_type",
			fmt.Sprintf("COALESCE(%s, '') AS recipe_name", recipeNameColumn),
		}).
		Joins("LEFT JOIN recipes AS r ON r.id = mp.recipe_id").
		Where("mp.user_id = ?", userID).
		Order("mp.scheduled_at DESC, mp.id DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list meal plans: %w", err)
	}

	result := make([]models.MealPlan, 0, len(rows))
	for _, row := range rows {
		day := row.MealType
		if row.ScheduledAt != nil {
			day = row.ScheduledAt.Weekday().String()
		}
		result = append(result, models.MealPlan{
			ID:          row.ID,
			UserID:      userID,
			Day:         day,
			MealName:    row.RecipeName,
			Servings:    0,
			Ingredients: []string{},
		})
	}
	return result, nil
}

func (r *RecipeBoxGormRepository) CreateMealPlan(ctx context.Context, userID int64, mealPlan models.MealPlan) (models.MealPlan, error) {
	mealPlan.ID = 0
	mealPlan.UserID = userID
	if err := r.db.WithContext(ctx).Create(&mealPlan).Error; err != nil {
		return models.MealPlan{}, fmt.Errorf("create meal plan: %w", err)
	}
	return mealPlan, nil
}

func (r *RecipeBoxGormRepository) UpdateMealPlan(ctx context.Context, userID, mealPlanID int64, mealPlan models.MealPlan) (models.MealPlan, error) {
	updates := map[string]any{
		"day":         mealPlan.Day,
		"meal_name":   mealPlan.MealName,
		"servings":    mealPlan.Servings,
		"ingredients": mealPlan.Ingredients,
		"updated_at":  time.Now().UTC(),
	}
	result := r.db.WithContext(ctx).
		Model(&models.MealPlan{}).
		Where("id = ? AND user_id = ?", mealPlanID, userID).
		Updates(updates)
	if result.Error != nil {
		return models.MealPlan{}, fmt.Errorf("update meal plan: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.MealPlan{}, models.ErrNotFound
	}

	var out models.MealPlan
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", mealPlanID, userID).Take(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.MealPlan{}, models.ErrNotFound
		}
		return models.MealPlan{}, fmt.Errorf("fetch updated meal plan: %w", err)
	}
	return out, nil
}

func (r *RecipeBoxGormRepository) DeleteMealPlan(ctx context.Context, userID, mealPlanID int64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", mealPlanID, userID).Delete(&models.MealPlan{})
	if result.Error != nil {
		return fmt.Errorf("delete meal plan: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *RecipeBoxGormRepository) ListShoppingItems(ctx context.Context, userID int64) ([]models.ShoppingItem, error) {
	return r.ListShoppingItemsPage(ctx, userID, 0, 0)
}

func (r *RecipeBoxGormRepository) ListShoppingItemsPage(ctx context.Context, userID int64, limit, offset int) ([]models.ShoppingItem, error) {
	type shoppingRow struct {
		ID       int64  `gorm:"column:id"`
		MenuName string `gorm:"column:menu_name"`
		Name     string `gorm:"column:name"`
		Qty      string `gorm:"column:qty"`
		Checked  bool   `gorm:"column:checked"`
	}

	qtyColumn := "si.qty"
	if !r.hasColumn(&models.ShoppingItem{}, "qty") && r.hasColumn(&models.ShoppingItem{}, "quantity") {
		qtyColumn = "si.quantity"
	}

	menuNameExpr := "COALESCE(si.menu_name, '') AS menu_name"
	joins := ""
	if !r.hasColumn(&models.ShoppingItem{}, "menu_name") && r.hasColumn(&models.ShoppingItem{}, "recipe_id") {
		recipeNameColumn := "r.name"
		if !r.hasColumn(&models.Recipe{}, "name") && r.hasColumn(&models.Recipe{}, "title") {
			recipeNameColumn = "r.title"
		}
		menuNameExpr = fmt.Sprintf("COALESCE(%s, '') AS menu_name", recipeNameColumn)
		joins = "LEFT JOIN recipes AS r ON r.id = si.recipe_id"
	}

	orderExpr := "menu_name ASC, si.id DESC"
	if r.hasColumn(&models.ShoppingItem{}, "updated_at") {
		orderExpr = "menu_name ASC, si.updated_at DESC, si.id DESC"
	}

	var rows []shoppingRow
	query := r.db.WithContext(ctx).
		Table("shopping_items AS si").
		Select([]string{
			"si.id",
			menuNameExpr,
			"COALESCE(si.name, '') AS name",
			fmt.Sprintf("COALESCE(%s, '') AS qty", qtyColumn),
			"si.checked",
		}).
		Where("si.user_id = ?", userID)
	if joins != "" {
		query = query.Joins(joins)
	}
	query = query.Order(orderExpr)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list shopping items: %w", err)
	}

	items := make([]models.ShoppingItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.ShoppingItem{
			ID:       row.ID,
			UserID:   userID,
			MenuName: row.MenuName,
			Name:     row.Name,
			Qty:      row.Qty,
			Checked:  row.Checked,
		})
	}
	return items, nil
}

func (r *RecipeBoxGormRepository) CreateShoppingItem(ctx context.Context, userID int64, item models.ShoppingItem) (models.ShoppingItem, error) {
	item.ID = 0
	item.UserID = userID
	if item.Checked {
		now := time.Now().UTC()
		item.CheckedAt = &now
	} else {
		item.CheckedAt = nil
	}
	if err := r.db.WithContext(ctx).Create(&item).Error; err != nil {
		return models.ShoppingItem{}, fmt.Errorf("create shopping item: %w", err)
	}
	return item, nil
}

func (r *RecipeBoxGormRepository) UpdateShoppingItem(ctx context.Context, userID, itemID int64, item models.ShoppingItem) (models.ShoppingItem, error) {
	updates := map[string]any{
		"menu_name":  item.MenuName,
		"name":       item.Name,
		"qty":        item.Qty,
		"checked":    item.Checked,
		"updated_at": time.Now().UTC(),
	}
	if item.Checked {
		now := time.Now().UTC()
		updates["checked_at"] = &now
	} else {
		updates["checked_at"] = nil
	}

	result := r.db.WithContext(ctx).
		Model(&models.ShoppingItem{}).
		Where("id = ? AND user_id = ?", itemID, userID).
		Updates(updates)
	if result.Error != nil {
		return models.ShoppingItem{}, fmt.Errorf("update shopping item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.ShoppingItem{}, models.ErrNotFound
	}

	var out models.ShoppingItem
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", itemID, userID).Take(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ShoppingItem{}, models.ErrNotFound
		}
		return models.ShoppingItem{}, fmt.Errorf("fetch updated shopping item: %w", err)
	}
	return out, nil
}

func (r *RecipeBoxGormRepository) DeleteShoppingItem(ctx context.Context, userID, itemID int64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", itemID, userID).Delete(&models.ShoppingItem{})
	if result.Error != nil {
		return fmt.Errorf("delete shopping item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *RecipeBoxGormRepository) hasColumn(model any, column string) bool {
	return r.db != nil && r.db.Migrator().HasColumn(model, column)
}
