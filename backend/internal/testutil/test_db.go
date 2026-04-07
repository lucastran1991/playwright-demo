package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/user/app/internal/database"
	"github.com/user/app/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTestDB connects to a test PostgreSQL database, runs migrations, and
// returns the DB handle plus a cleanup function. Skips the test if DB is unavailable.
func SetupTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	dbName := getEnvDefault("TEST_DB_NAME", "app_test")
	cfg := &config.Config{
		DBHost:    getEnvDefault("DB_HOST", "localhost"),
		DBPort:    getEnvDefault("DB_PORT", "5432"),
		DBUser:    getEnvDefault("DB_USER", "postgres"),
		DBPassword: getEnvDefault("DB_PASSWORD", "postgres"),
		DBName:    dbName,
		DBSSLMode: getEnvDefault("DB_SSLMODE", "disable"),
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		t.Skipf("PostgreSQL not available (db=%s): %v", dbName, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Minute)

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}
	return db, cleanup
}

// TruncateAll removes all data from tracer-related tables in FK-safe order.
func TruncateAll(db *gorm.DB) error {
	tables := []string{
		"blueprint_edges",
		"blueprint_node_memberships",
		"blueprint_nodes",
		"blueprint_types",
		"capacity_node_types",
		"dependency_rules",
		"impact_rules",
	}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			return fmt.Errorf("truncate %s: %w", table, err)
		}
	}
	return nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
