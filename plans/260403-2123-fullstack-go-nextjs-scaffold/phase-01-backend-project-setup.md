# Phase 1: Backend Project Setup

## Overview

- **Priority:** P1
- **Status:** completed
- **Description:** Initialize Go module, create directory structure, config loading, and entry point.

## Key Insights

- `internal/` prevents external import -- enforces encapsulation (Go convention)
- Use `godotenv` for .env loading + struct-based config with validation
- Singular package names per tech-stack.md: `handler/`, `model/`, `service/`, `repository/`

## Related Code Files

**Create:**
- `/backend/go.mod`
- `/backend/cmd/server/main.go`
- `/backend/internal/config/config.go`
- `/backend/internal/router/router.go`
- `/backend/pkg/response/response.go`
- `/backend/.env.example`
- `/backend/.env`
- `/backend/.gitignore`

## Implementation Steps

1. **Init Go module**
   ```bash
   cd /backend && go mod init github.com/user/app
   ```

2. **Create directory structure**
   ```
   backend/
   ├── cmd/server/main.go
   ├── internal/
   │   ├── config/
   │   ├── handler/
   │   ├── middleware/
   │   ├── model/
   │   ├── repository/
   │   ├── service/
   │   └── router/
   └── pkg/
       ├── response/
       └── token/
   ```

3. **Create config.go** -- load env vars into typed struct
   ```go
   type Config struct {
       DBHost     string
       DBPort     string
       DBUser     string
       DBPassword string
       DBName     string
       DBSSLMode  string
       JWTSecret  string
       ServerPort string
   }
   ```
   - Use `godotenv.Load()` then `os.Getenv()` with defaults
   - Validate required fields, return error if missing

4. **Create response.go** -- standard JSON response helpers
   ```go
   func Success(c *gin.Context, status int, data interface{})
   func Error(c *gin.Context, status int, message string)
   ```

5. **Create router.go** -- Gin engine setup with CORS middleware
   - `gin.Default()` with recovery and logger
   - CORS config: allow localhost:3000, credentials, common headers
   - Health check: `GET /health`
   - Return `*gin.Engine`

6. **Create main.go** -- wire config, DB (placeholder), router, start server
   ```go
   func main() {
       cfg := config.Load()
       r := router.Setup()
       r.Run(":" + cfg.ServerPort)
   }
   ```

7. **Create .env.example**
   ```
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=app_dev
   DB_SSLMODE=disable
   JWT_SECRET=your-secret-key-min-32-chars
   SERVER_PORT=8080
   ```

8. **Create .gitignore** -- include `.env`, binaries, vendor/

9. **Install dependencies**
   ```bash
   go get github.com/gin-gonic/gin
   go get github.com/gin-contrib/cors
   go get github.com/joho/godotenv
   ```

## Todo List

- [x] Init go module
- [x] Create directory structure
- [x] Implement config loading
- [x] Implement response helpers
- [x] Setup router with CORS & health check
- [x] Create main.go entry point
- [x] Create .env.example and .gitignore
- [x] Verify `go build ./...` succeeds

## Success Criteria

- `go build ./cmd/server` compiles without errors
- `go run ./cmd/server` starts server on :8080
- `GET /health` returns 200
- CORS headers present for localhost:3000
