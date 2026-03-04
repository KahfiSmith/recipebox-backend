package models

import "time"

type Recipe struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `json:"userId" gorm:"column:user_id;not null;index:idx_recipes_user_id"`
	Name      string    `json:"name" gorm:"column:name;type:text;not null"`
	Category  string    `json:"category" gorm:"column:category;type:text;not null;default:''"`
	PrepTime  int       `json:"prepTime" gorm:"column:prep_time;not null;default:15"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime;index:idx_recipes_updated_at,sort:desc"`
}

func (Recipe) TableName() string {
	return "recipes"
}

type MealPlan struct {
	ID          int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID      int64     `json:"userId" gorm:"column:user_id;not null;index:idx_meal_plans_user_id"`
	Day         string    `json:"day" gorm:"column:day;type:text;not null;default:'';index:idx_meal_plans_day"`
	MealName    string    `json:"mealName" gorm:"column:meal_name;type:text;not null;default:''"`
	Servings    int       `json:"servings" gorm:"column:servings;not null;default:2"`
	Ingredients []string  `json:"ingredients" gorm:"column:ingredients;serializer:json;type:jsonb;not null;default:'[]'"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (MealPlan) TableName() string {
	return "meal_plans"
}

type ShoppingItem struct {
	ID        int64      `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64      `json:"userId" gorm:"column:user_id;not null;index:idx_shopping_items_user_id,priority:1;index:idx_shopping_items_checked,priority:1"`
	MenuName  string     `json:"menuName" gorm:"column:menu_name;type:text;not null;default:'';index:idx_shopping_items_menu_name"`
	Name      string     `json:"name" gorm:"column:name;type:text;not null"`
	Qty       string     `json:"qty" gorm:"column:qty;type:text;not null;default:''"`
	Checked   bool       `json:"checked" gorm:"column:checked;not null;default:false;index:idx_shopping_items_checked,priority:2"`
	CheckedAt *time.Time `json:"checkedAt,omitempty" gorm:"column:checked_at"`
	CreatedAt time.Time  `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (ShoppingItem) TableName() string {
	return "shopping_items"
}
