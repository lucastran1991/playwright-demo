package testutil

import (
	"fmt"
	"os"
	"time"

	"github.com/user/app/internal/database"
	"github.com/user/app/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTestDBForMain connects to a test PostgreSQL database and runs migrations.
// Returns nil DB if PostgreSQL is unavailable (caller should skip or exit gracefully).
// Designed for TestMain where *testing.T is not available.
func SetupTestDBForMain() (*gorm.DB, func()) {
	dbName := getEnvDefault("TEST_DB_NAME", "app_test")
	cfg := &config.Config{
		DBHost:     getEnvDefault("DB_HOST", "localhost"),
		DBPort:     getEnvDefault("DB_PORT", "5432"),
		DBUser:     getEnvDefault("DB_USER", "postgres"),
		DBPassword: getEnvDefault("DB_PASSWORD", "postgres"),
		DBName:     dbName,
		DBSSLMode:  getEnvDefault("DB_SSLMODE", "disable"),
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, func() {}
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Minute)

	if err := database.Migrate(db); err != nil {
		return nil, func() {}
	}

	cleanup := func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}
	return db, cleanup
}

// TruncateAndSeedForMain atomically truncates all tables and seeds fixture data.
// Uses a PostgreSQL advisory lock so parallel test packages serialize their
// setup against the same database. Panics on failure (TestMain cannot t.Fatal).
func TruncateAndSeedForMain(db *gorm.DB) {
	const lockID = 12345
	db.Exec("SELECT pg_advisory_lock(?)", lockID)

	tables := []string{
		"blueprint_edges", "blueprint_node_memberships", "blueprint_nodes",
		"blueprint_types", "capacity_node_types", "dependency_rules", "impact_rules",
	}
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			db.Exec("SELECT pg_advisory_unlock(?)", lockID)
			panic(fmt.Sprintf("truncate %s: %v", table, err))
		}
	}

	seedTraceFixturesPanic(db)
	db.Exec("SELECT pg_advisory_unlock(?)", lockID)
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
