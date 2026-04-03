---
title: Go Backend Best Practices - Gin, GORM, JWT
date: 2026-04-03
type: Research Report
---

# Go Backend Best Practices: Gin, GORM, JWT Patterns

## Project Structure (DDD/Clean Architecture)

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point only
├── internal/
│   ├── config/                     # Config loading (env, secrets)
│   ├── middleware/                 # Gin middleware (auth, logging, CORS)
│   ├── handlers/                   # HTTP handlers (request/response)
│   ├── models/                     # GORM models & migrations
│   ├── services/                   # Business logic layer
│   ├── repositories/               # Data access layer (GORM)
│   ├── auth/                       # JWT token generation/validation
│   └── errors/                     # Custom error types & handling
├── pkg/
│   ├── logger/                     # Reusable logging
│   └── utils/                      # Utility functions
├── migrations/                     # SQL migration files (optional)
├── tests/                          # Integration/e2e tests
├── go.mod & go.sum
├── Dockerfile
└── .env.example
```

**Key principle:** `internal/` is not importable by external packages. `pkg/` is reusable.

## Gin Middleware Patterns for JWT

**Approach: appleboy/gin-jwt** (production standard)

1. **Two-token system:**
   - **Access token** (JWT, short-lived: 15min) - stateless, embedded claims
   - **Refresh token** (opaque, long-lived: 7d) - server-side stored, rotated on use

2. **Middleware setup:**
   ```go
   authMiddleware, _ := jwt.New(&jwt.GinJWTMiddleware{
       Realm: "api",
       Key: []byte(os.Getenv("JWT_SECRET")),
       Timeout: 15 * time.Minute,
       RefreshTimeout: 7 * 24 * time.Hour,
       Authenticator: authenticateUser,      // Login validation
       Authorizer: authorizeUser,            // Role/permission check
       TokenLookup: "header:Authorization",  // Extract from Bearer token
       TokenHeadName: "Bearer",
   })
   
   router.POST("/login", authMiddleware.LoginHandler)
   router.POST("/refresh", authMiddleware.RefreshHandler)
   protected := router.Group("/api")
   protected.Use(authMiddleware.MiddlewareFunc())
   ```

3. **Refresh token storage:** Server-side (Redis or in-memory). Don't use JWT for refresh tokens—prevents vulnerabilities.

## GORM Models & Migrations

**For production:** Use versioned migrations (Goose, golang-migrate, or Gormigrate) + AutoMigrate safeguard.

```go
// models/user.go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex;not null"`
    Password  string    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// migrations/migrate.go
func RunMigrations(db *gorm.DB) error {
    // Option 1: Simple AutoMigrate for dev
    return db.AutoMigrate(&User{}, &Post{})
    
    // Option 2: Gormigrate for explicit control + rollback
    m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
        {
            ID: "202604031000",
            Migrate: func(tx *gorm.DB) error {
                return tx.AutoMigrate(&User{})
            },
            Rollback: func(tx *gorm.DB) error {
                return tx.Migrator().DropTable(&User{})
            },
        },
    })
    return m.Migrate()
}
```

**Key rule:** AutoMigrate won't delete unused columns (data safety), but use explicit migrations in production for rollback support & team CI/CD.

## Error Handling Pattern

Go philosophy: errors as values, explicit handling. Use custom types for domain errors.

```go
// internal/errors/errors.go
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUnauthorized = errors.New("unauthorized")
)

type AppError struct {
    Code    string
    Message string
    Status  int
    Err     error
}

// In handler:
user, err := repo.GetUser(id)
if err != nil {
    if errors.Is(err, ErrUserNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
    return
}
```

**Pattern:** Sentinel errors + error wrapping with `fmt.Errorf("%w", err)` for stack context.

## Environment Configuration

Use Go's `os` pkg + struct-based validation:

```go
// internal/config/config.go
type Config struct {
    DB struct {
        Host     string `env:"DB_HOST" default:"localhost"`
        Port     string `env:"DB_PORT" default:"5432"`
        User     string `env:"DB_USER" required:"true"`
        Password string `env:"DB_PASSWORD" required:"true"`
        Name     string `env:"DB_NAME" required:"true"`
    }
    JWT struct {
        Secret string `env:"JWT_SECRET" required:"true"`
    }
    Server struct {
        Port string `env:"SERVER_PORT" default:"8080"`
    }
}

// Load from .env via github.com/joho/godotenv
func Load() (*Config, error) {
    godotenv.Load()
    // Parse from env vars + validate
}
```

## Dependency Injection

Avoid globals. Wire dependencies in handlers:

```go
type Handler struct {
    userService *UserService
    authService *AuthService
}

func (h *Handler) Login(c *gin.Context) {
    // Use h.userService, h.authService
}

// In main.go:
db := initDB()
userRepo := repositories.NewUserRepository(db)
userService := services.NewUserService(userRepo)
handler := handlers.NewHandler(userService)
```

## Performance Best Practices

- **Connection pooling:** GORM auto-manages via `db.DB().SetMaxOpenConns(25)`.
- **Route grouping:** Group related routes to avoid duplicate middleware.
  ```go
  api := router.Group("/api")
  api.Use(authMiddleware.MiddlewareFunc())
  api.GET("/users", handler.ListUsers)
  ```
- **Caching:** Redis for refresh token revocation lists + session data.
- **Middleware order:** Auth → logging → CORS → rate limit.

## Testing Strategy

- **Unit tests:** Services & repositories with mocked dependencies.
- **Integration tests:** Hit real Postgres in Docker.
- **Middleware tests:** Validate JWT extraction & claim validation.

```go
// Example: Test JWT middleware
func TestAuthMiddleware(t *testing.T) {
    router := setupTestRouter()
    req := httptest.NewRequest("GET", "/api/users", nil)
    req.Header.Set("Authorization", "Bearer invalid_token")
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)
    assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
```

---

## Key Takeaways

1. **Structure:** Clean layers (handlers → services → repos) + dependency injection.
2. **JWT:** Use appleboy/gin-jwt; keep access tokens short-lived, refresh tokens server-side.
3. **Migrations:** AutoMigrate for dev, Gormigrate/golang-migrate for prod + rollback.
4. **Errors:** Custom error types, explicit handling, no exceptions.
5. **Config:** Environment variables, validated at startup, no secrets in code.
6. **Testing:** Unit + integration, real DB in tests.

---

## Sources

- [Go Gin Tutorial](https://go.dev/doc/tutorial/web-service-gin)
- [Gin Framework Best Practices](https://withcodeexample.com/chapter-10-real-world-projects-and-best-practices-with-gin/)
- [GORM Migration Docs](https://gorm.io/docs/migration.html)
- [Gormigrate Library](https://github.com/go-gormigrate/gormigrate)
- [Gin-JWT Middleware](https://github.com/appleboy/gin-jwt)
- [JWT Refresh Tokens in Go](https://ademawan.medium.com/jwt-authentication-in-go-implementing-refresh-tokens-for-secure-sessions-9a5baa9d2650)
- [Go Clean Architecture](https://dev.to/leapcell/clean-architecture-in-go-a-practical-guide-with-go-clean-arch-51h7)
- [Go Error Handling Patterns](https://dev.to/djamware_tutorial_eba1a61/error-handling-in-go-idiomatic-patterns-for-clean-code-4j54)
