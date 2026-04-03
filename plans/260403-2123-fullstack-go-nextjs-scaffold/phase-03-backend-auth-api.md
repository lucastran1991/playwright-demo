# Phase 3: Backend Auth API

## Overview

- **Priority:** P1
- **Status:** completed
- **Description:** JWT utilities, auth handlers (register/login/refresh/me), auth middleware, route setup.

## Key Insights

- Use `golang-jwt/jwt/v5` directly (not appleboy/gin-jwt)
- Access token: 15min, Refresh token: 7 days (both JWT for simplicity)
- bcrypt for password hashing (`golang.org/x/crypto/bcrypt`)
- go-playground/validator for request validation
- Clean layers: handler -> service -> repository

## Related Code Files

**Create:**
- `/backend/pkg/token/token.go` -- JWT generate/validate
- `/backend/internal/repository/user_repository.go`
- `/backend/internal/service/auth_service.go`
- `/backend/internal/handler/auth_handler.go`
- `/backend/internal/middleware/auth_middleware.go`

**Modify:**
- `/backend/internal/router/router.go` -- add auth routes
- `/backend/cmd/server/main.go` -- wire dependencies

## Implementation Steps

### 1. JWT Token Utilities (`pkg/token/token.go`)

```go
package token

type Claims struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func GenerateAccessToken(userID uint, email, secret string) (string, error)
// Returns signed JWT, expiry 15min

func GenerateRefreshToken(userID uint, email, secret string) (string, error)
// Returns signed JWT, expiry 7 days

func ValidateToken(tokenString, secret string) (*Claims, error)
// Parse + validate, return claims or error
```

### 2. User Repository (`internal/repository/user_repository.go`)

```go
type UserRepository struct { db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository
func (r *UserRepository) Create(user *model.User) error
func (r *UserRepository) FindByEmail(email string) (*model.User, error)
func (r *UserRepository) FindByID(id uint) (*model.User, error)
```

### 3. Auth Service (`internal/service/auth_service.go`)

```go
type AuthService struct {
    userRepo  *repository.UserRepository
    jwtSecret string
}

func NewAuthService(repo *repository.UserRepository, secret string) *AuthService

func (s *AuthService) Register(name, email, password string) (*model.User, error)
// Validate input, hash password with bcrypt, create user, return user

func (s *AuthService) Login(email, password string) (accessToken, refreshToken string, err error)
// Find user by email, compare bcrypt hash, generate tokens

func (s *AuthService) RefreshToken(refreshTokenStr string) (accessToken, refreshToken string, err error)
// Validate refresh token, generate new pair

func (s *AuthService) GetUser(id uint) (*model.User, error)
// Return user by ID
```

### 4. Auth Handler (`internal/handler/auth_handler.go`)

Request/response structs with validation tags:

```go
type RegisterRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
    AccessToken  string      `json:"access_token"`
    RefreshToken string      `json:"refresh_token"`
    User         model.User  `json:"user"`
}
```

Handler methods:
```go
type AuthHandler struct { authService *service.AuthService }

func (h *AuthHandler) Register(c *gin.Context)
// POST /api/auth/register -- bind JSON, call service, return 201 + tokens

func (h *AuthHandler) Login(c *gin.Context)
// POST /api/auth/login -- bind JSON, call service, return 200 + tokens

func (h *AuthHandler) RefreshToken(c *gin.Context)
// POST /api/auth/refresh -- bind JSON, call service, return 200 + new tokens

func (h *AuthHandler) Me(c *gin.Context)
// GET /api/auth/me -- extract user ID from context (set by middleware), return user
```

### 5. Auth Middleware (`internal/middleware/auth_middleware.go`)

```go
func AuthRequired(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract "Bearer <token>" from Authorization header
        // 2. Validate token with token.ValidateToken()
        // 3. Set userID and email in gin.Context
        // 4. c.Next() or abort with 401
    }
}
```

### 6. Route Setup (update `internal/router/router.go`)

```go
func Setup(authHandler *handler.AuthHandler, jwtSecret string) *gin.Engine {
    r := gin.Default()
    r.Use(corsMiddleware())

    r.GET("/health", healthCheck)

    auth := r.Group("/api/auth")
    {
        auth.POST("/register", authHandler.Register)
        auth.POST("/login", authHandler.Login)
        auth.POST("/refresh", authHandler.RefreshToken)
    }

    protected := r.Group("/api")
    protected.Use(middleware.AuthRequired(jwtSecret))
    {
        protected.GET("/auth/me", authHandler.Me)
    }

    return r
}
```

### 7. Wire in main.go

```go
db, _ := database.Connect(cfg)
database.Migrate(db)
userRepo := repository.NewUserRepository(db)
authService := service.NewAuthService(userRepo, cfg.JWTSecret)
authHandler := handler.NewAuthHandler(authService)
r := router.Setup(authHandler, cfg.JWTSecret)
r.Run(":" + cfg.ServerPort)
```

### 8. Install dependencies

```bash
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

## Todo List

- [x] Implement JWT token utilities
- [x] Implement user repository
- [x] Implement auth service (register, login, refresh)
- [x] Implement auth handler with request validation
- [x] Implement auth middleware
- [x] Update router with auth routes
- [x] Wire dependencies in main.go
- [x] Test: register, login, refresh, me endpoints manually

## Success Criteria

- `POST /api/auth/register` creates user, returns tokens
- `POST /api/auth/login` validates credentials, returns tokens
- `POST /api/auth/refresh` issues new token pair
- `GET /api/auth/me` returns user (requires valid access token)
- Invalid/missing token returns 401
- Duplicate email returns 409
- Invalid input returns 400 with field errors
