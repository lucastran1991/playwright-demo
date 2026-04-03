# Phase 2: Backend Database Models

## Overview

- **Priority:** P1
- **Status:** completed
- **Description:** GORM setup, User model, DB connection, AutoMigrate.

## Key Insights

- AutoMigrate for dev -- won't delete columns, safe for iteration
- Use `gorm.Model` base (ID, CreatedAt, UpdatedAt, DeletedAt) or custom fields
- Connection pooling via `SetMaxOpenConns`, `SetMaxIdleConns`
- Per tech-stack: singular `model/` package name

## Related Code Files

**Create:**
- `/backend/internal/model/user.go`
- `/backend/internal/database/database.go`

**Modify:**
- `/backend/cmd/server/main.go` -- add DB init call
- `/backend/internal/config/config.go` -- DSN builder helper

## Implementation Steps

1. **Create database.go** -- connection + AutoMigrate
   ```go
   package database

   func Connect(cfg *config.Config) (*gorm.DB, error) {
       dsn := fmt.Sprintf(
           "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
           cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
       )
       db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
       if err != nil {
           return nil, err
       }
       sqlDB, _ := db.DB()
       sqlDB.SetMaxOpenConns(25)
       sqlDB.SetMaxIdleConns(5)
       sqlDB.SetConnMaxLifetime(5 * time.Minute)
       return db, nil
   }

   func Migrate(db *gorm.DB) error {
       return db.AutoMigrate(&model.User{})
   }
   ```

2. **Create user.go** -- User model
   ```go
   package model

   type User struct {
       ID        uint           `gorm:"primaryKey" json:"id"`
       Name      string         `gorm:"size:100;not null" json:"name"`
       Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
       Password  string         `gorm:"not null" json:"-"`
       CreatedAt time.Time      `json:"created_at"`
       UpdatedAt time.Time      `json:"updated_at"`
   }
   ```
   - Password excluded from JSON via `json:"-"`
   - Email has unique index

3. **Update main.go** -- add DB connection and migration
   ```go
   db, err := database.Connect(cfg)
   if err != nil { log.Fatal("DB connection failed:", err) }
   database.Migrate(db)
   ```

4. **Install GORM dependencies**
   ```bash
   go get gorm.io/gorm
   go get gorm.io/driver/postgres
   ```

## Todo List

- [x] Create database connection module
- [x] Create User model
- [x] Wire DB connection in main.go
- [x] Run AutoMigrate on startup
- [x] Verify table created in PostgreSQL

## Success Criteria

- App connects to PostgreSQL on startup
- `users` table created with correct columns
- Connection pooling configured
- Build compiles cleanly
