package db

import (
	"context"
	"fmt"
	"time"

	"recipebox-backend-go/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenPostgres(ctx context.Context, dsn string) (*gorm.DB, error) {
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return database, nil
}

func AutoMigrate(ctx context.Context, database *gorm.DB) error {
	if err := database.WithContext(ctx).AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.EmailVerificationToken{},
		&models.PasswordResetToken{},
	); err != nil {
		return fmt.Errorf("auto migrate auth: %w", err)
	}

	if err := AutoMigrateRecipeBox(ctx, database); err != nil {
		return err
	}

	return nil
}

func AutoMigrateRecipeBox(ctx context.Context, database *gorm.DB) error {
	migrator := database.WithContext(ctx).Migrator()

	if migrator.HasTable(&models.Recipe{}) &&
		migrator.HasColumn(&models.Recipe{}, "title") &&
		!migrator.HasColumn(&models.Recipe{}, "name") {
		if err := migrator.RenameColumn(&models.Recipe{}, "title", "name"); err != nil {
			return fmt.Errorf("rename recipes.title to recipes.name: %w", err)
		}
	}

	if migrator.HasTable(&models.ShoppingItem{}) &&
		migrator.HasColumn(&models.ShoppingItem{}, "quantity") &&
		!migrator.HasColumn(&models.ShoppingItem{}, "qty") {
		if err := migrator.RenameColumn(&models.ShoppingItem{}, "quantity", "qty"); err != nil {
			return fmt.Errorf("rename shopping_items.quantity to shopping_items.qty: %w", err)
		}
	}

	if err := database.WithContext(ctx).AutoMigrate(
		&models.Recipe{},
		&models.MealPlan{},
		&models.ShoppingItem{},
	); err != nil {
		return fmt.Errorf("auto migrate recipebox: %w", err)
	}

	for _, column := range []string{"description", "instructions", "total_ingredients"} {
		if migrator.HasColumn(&models.Recipe{}, column) {
			if err := migrator.DropColumn(&models.Recipe{}, column); err != nil {
				return fmt.Errorf("drop recipes.%s: %w", column, err)
			}
		}
	}

	for _, column := range []string{"recipe_id", "scheduled_at", "meal_type", "notes"} {
		if migrator.HasColumn(&models.MealPlan{}, column) {
			if err := migrator.DropColumn(&models.MealPlan{}, column); err != nil {
				return fmt.Errorf("drop meal_plans.%s: %w", column, err)
			}
		}
	}

	for _, column := range []string{"recipe_id"} {
		if migrator.HasColumn(&models.ShoppingItem{}, column) {
			if err := migrator.DropColumn(&models.ShoppingItem{}, column); err != nil {
				return fmt.Errorf("drop shopping_items.%s: %w", column, err)
			}
		}
	}

	return nil
}
