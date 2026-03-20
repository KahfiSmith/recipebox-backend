CREATE INDEX IF NOT EXISTS idx_recipes_user_updated_at ON recipes(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_meal_plans_user_id_desc ON meal_plans(user_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_shopping_items_user_menu_updated_id ON shopping_items(user_id, menu_name, updated_at DESC, id DESC);
