# Tech Stack

## Backend (`/backend`)

| Category | Technology | Version |
|----------|-----------|---------|
| Language | Go | 1.22+ |
| Web Framework | Gin | v1.10+ |
| ORM | GORM | v1.25+ |
| Database | PostgreSQL | 15+ |
| Authentication | JWT (golang-jwt/jwt/v5) | v5 |
| Config | godotenv | latest |
| Validation | go-playground/validator | v10 |
| Password Hashing | bcrypt (golang.org/x/crypto) | latest |

### Backend Structure
```
backend/
  cmd/server/         # Entry point
  internal/
    config/           # App configuration
    handler/          # HTTP handlers (controllers)
    middleware/       # Auth, CORS, logging middleware
    model/            # GORM models
    repository/       # Database operations
    service/          # Business logic
    router/           # Route definitions
  pkg/
    response/         # Standard API responses
    token/            # JWT utilities
```

## Frontend (`/frontend`)

| Category | Technology | Version |
|----------|-----------|---------|
| Framework | Next.js | 15 (App Router) |
| Language | TypeScript | 5+ |
| UI Components | Shadcn/ui | latest |
| Styling | Tailwind CSS | v4 |
| Data Fetching | TanStack Query | v5 |
| Forms | React Hook Form | v7 |
| Validation | Zod | v3 |
| Authentication | NextAuth.js (Auth.js) | v5 |

### Frontend Structure
```
frontend/
  src/
    app/              # Next.js App Router pages
    components/       # Shared UI components
    lib/              # Utilities, API client, auth config
    hooks/            # Custom React hooks
    types/            # TypeScript type definitions
    providers/        # Context providers (query, auth, theme)
```

## Communication
- REST API over HTTP/HTTPS
- JSON request/response format
- JWT Bearer token in Authorization header
- NextAuth.js handles token lifecycle on frontend, Go backend issues/validates tokens

## Development Requirements
- Go 1.22+
- Node.js 20+
- PostgreSQL 15+
- pnpm (frontend package manager)
