package main

import (
	"log"

	"github.com/user/app/internal/config"
	"github.com/user/app/internal/database"
	"github.com/user/app/internal/handler"
	"github.com/user/app/internal/repository"
	"github.com/user/app/internal/router"
	"github.com/user/app/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}
	log.Println("Database connected and migrated")

	// Wire dependencies
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService)

	r := router.Setup(authHandler, cfg.JWTSecret)

	log.Printf("Server starting on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
