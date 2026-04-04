package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	JWTSecret  string
	ServerPort   string
	CORSOrigin   string
	BlueprintDir string
	ModelDir     string
}

// systemCfg represents the relevant parts of system.cfg.json.
type systemCfg struct {
	Backend  struct{ Port int `json:"port"` }  `json:"backend"`
	Frontend struct{ Port int `json:"port"` }  `json:"frontend"`
}

// loadSystemCfg reads system.cfg.json from known paths (relative to backend/ or project root).
func loadSystemCfg() *systemCfg {
	for _, path := range []string{"../system.cfg.json", "./system.cfg.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg systemCfg
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Printf("WARNING: failed to parse %s: %v", path, err)
			continue
		}
		log.Printf("Loaded ports from %s (backend=%d, frontend=%d)", path, cfg.Backend.Port, cfg.Frontend.Port)
		return &cfg
	}
	return nil
}

// Load reads .env file and returns a validated Config.
// Priority: env vars > .env file > system.cfg.json > hardcoded defaults.
func Load() (*Config, error) {
	godotenv.Load()

	// Derive defaults from system.cfg.json
	defaultPort := "8889"
	defaultCORS := "http://localhost:8089"
	if sysCfg := loadSystemCfg(); sysCfg != nil {
		if sysCfg.Backend.Port > 0 {
			defaultPort = fmt.Sprintf("%d", sysCfg.Backend.Port)
		}
		if sysCfg.Frontend.Port > 0 {
			defaultCORS = fmt.Sprintf("http://localhost:%d", sysCfg.Frontend.Port)
		}
	}

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", ""),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
		ServerPort:   getEnv("SERVER_PORT", defaultPort),
		CORSOrigin:   getEnv("CORS_ORIGIN", defaultCORS),
		BlueprintDir: getEnv("BLUEPRINT_DIR", "./blueprint/Node & Edge"),
		ModelDir:     getEnv("MODEL_DIR", "./blueprint"),
	}

	if cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("DB_USER, DB_PASSWORD, and DB_NAME are required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
