package database

import (
	"fmt"
	"time"

	"github.com/user/app/internal/config"
	"github.com/user/app/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect establishes a connection to PostgreSQL and configures the pool.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// Migrate runs AutoMigrate for all models.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.User{})
}
