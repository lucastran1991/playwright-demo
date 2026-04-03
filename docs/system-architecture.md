# System Architecture

## High-Level Overview

```
┌─────────────────────────────────────────────────────┐
│                  Next.js Frontend (3000)             │
│  ┌──────────────────────────────────────────────┐   │
│  │  Pages: (auth), (dashboard), API Routes      │   │
│  │  Providers: Session, Query, Theme            │   │
│  │  Components: Forms, UI, Dashboard            │   │
│  └──────────────────────────────────────────────┘   │
└────────────────────────┬────────────────────────────┘
                         │ HTTP REST + JWT
                         │
          ┌──────────────────────────────────┐
          │  Go Gin Backend (8080)           │
          │  ┌──────────────────────────┐    │
          │  │ Router + Middleware      │    │
          │  │ Handlers → Services      │    │
          │  │ Repositories + Models    │    │
          │  └──────────────────────────┘    │
          └────────────┬─────────────────────┘
                       │
                   ┌───▼────┐
                   │ Postgres│
                   │   DB    │
                   └─────────┘
```

## Backend Architecture (Clean Layers)

### Entry Point
- **`cmd/server/main.go`** - Initializes dependencies, starts Gin server

### Configuration Layer
- **`internal/config/config.go`** - Loads environment variables, validates config
- Environment variables: `DB_*`, `JWT_SECRET`, `SERVER_PORT`

### Database Layer
- **`internal/database/database.go`** - PostgreSQL connection pooling, migrations
- **`internal/model/user.go`** - GORM model definition

### Repository Layer (Data Access)
- **`internal/repository/user_repository.go`** - CRUD operations on users
- Methods: `Create()`, `FindByID()`, `FindByEmail()`
- Uses GORM ORM

### Service Layer (Business Logic)
- **`internal/service/auth_service.go`** - Authentication operations
- Methods: `Register()`, `Login()`, `RefreshToken()`, `GetUser()`
- Handles password hashing, token generation, validation

### Handler Layer (HTTP)
- **`internal/handler/auth_handler.go`** - HTTP request/response handling
- Maps HTTP requests to service calls
- Input validation, error responses

### Router & Middleware
- **`internal/router/router.go`** - Route definitions, middleware setup
- **`internal/middleware/auth_middleware.go`** - JWT validation, extracts userID
- Middleware chain: CORS → Auth (protected routes) → Handler

### Token Management
- **`pkg/token/token.go`** - JWT generation/validation
- Access token: 15-minute expiry
- Refresh token: 7-day expiry
- Uses `golang-jwt/jwt/v5`

### Response Format
- **`pkg/response/response.go`** - Standard JSON response helpers
- Success: `{success: true, data: ...}`
- Error: `{success: false, error: ...}`

## Frontend Architecture (Next.js 15 App Router)

### Root Layout
- **`src/app/layout.tsx`** - HTML document setup, provider hierarchy
- Loads fonts (DM Sans, JetBrains Mono)
- Wraps content with: SessionProvider → ThemeProvider → QueryProvider

### Grouped Routes

#### Auth Group: `(auth)`
- **`(auth)/layout.tsx`** - Auth pages layout (centered card)
- **`(auth)/login/page.tsx`** - Login form with NextAuth signIn
- **`(auth)/register/page.tsx`** - Registration form with API call

#### Dashboard Group: `(dashboard)`
- **`(dashboard)/layout.tsx`** - Protected layout with sidebar + topbar
  - Sidebar: Navigation menu (responsive, collapsible)
  - Topbar: Breadcrumb, theme toggle, user menu
  - Main content area: Scrollable, padded
- **`(dashboard)/page.tsx`** - Dashboard home page

### API Routes
- **`src/app/api/auth/[...nextauth]/route.ts`** - NextAuth.js route handler
  - Exports `{ GET, POST }` from `lib/auth`

### Authentication Flow

1. **Registration**
   - User fills form → validate with Zod
   - POST `/api/auth/register` (Go backend)
   - Backend returns: `{user, access_token, refresh_token}`
   - NextAuth stores tokens in encrypted cookie

2. **Login**
   - User enters credentials → NextAuth signIn
   - NextAuth calls CredentialsProvider → POST `/api/auth/login` (Go backend)
   - Backend validates, returns tokens
   - NextAuth stores session in JWT cookie

3. **Protected Pages**
   - Middleware checks session before allowing access
   - Redirects to `/login` if no session
   - Frontend passes Bearer token in API requests

4. **Token Refresh**
   - NextAuth JWT callback refreshes tokens automatically
   - Frontend can call POST `/api/auth/refresh` if needed

### Providers & Context

**`src/providers/session-provider.tsx`**
- Wraps app with NextAuth SessionProvider
- Makes `useSession()` available

**`src/providers/theme-provider.tsx`**
- Wraps app with next-themes ThemeProvider
- Dark/light mode toggle support
- CSS variable-based theming

**`src/providers/query-provider.tsx`**
- TanStack Query QueryClientProvider
- Manages server state, caching, fetching
- Used by API hooks

### Custom Hooks

**`src/hooks/use-auth.ts`**
- Wrapper around `useSession()`
- Returns session & auth status

**`src/hooks/use-api.ts`**
- Wrapper around `useQuery`/`useMutation`
- Handles Bearer token injection
- Manages loading/error states

**`src/hooks/use-mobile.ts`**
- Detects mobile viewport
- Used for responsive UI (sidebar behavior)

### Components

**Auth Components** (`src/components/auth/`)
- `login-form.tsx` - Form with React Hook Form + Zod
- `register-form.tsx` - Registration form

**Dashboard Components** (`src/components/dashboard/`)
- `sidebar-nav.tsx` - Sidebar navigation menu
- `topbar.tsx` - Top bar with controls
- `user-menu.tsx` - User profile dropdown
- `theme-toggle.tsx` - Dark/light mode switch

**UI Components** (`src/components/ui/`)
- Shadcn/ui library: Button, Input, Card, Dialog, etc.

### Middleware
- **`src/middleware.ts`** - NextAuth middleware
- Protects `/dashboard/*` routes
- Redirects unauthenticated users to login

## API Contracts

### Authentication Endpoints

```
POST /api/auth/register
Body: {name, email, password}
Response: {user, access_token, refresh_token}

POST /api/auth/login
Body: {email, password}
Response: {user, access_token, refresh_token}

POST /api/auth/refresh
Body: {refresh_token}
Response: {access_token, refresh_token}

GET /api/auth/me
Headers: Authorization: Bearer {token}
Response: {user}
```

### Response Format

**Success (200-201)**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "user": {
    "id": 1,
    "name": "John",
    "email": "john@example.com"
  }
}
```

**Error (400-500)**
```json
{
  "error": "Invalid email or password"
}
```

## Data Flow Examples

### Login Flow
1. User submits credentials in NextAuth form
2. NextAuth calls CredentialsProvider.authorize()
3. authorize() POSTs to Go `/api/auth/login`
4. Go handler validates, returns user + tokens
5. NextAuth stores session in JWT cookie
6. Frontend redirects to dashboard
7. Subsequent API calls include Bearer token

### Protected Request Flow
1. Frontend component calls API via `use-api` hook
2. Hook injects Bearer token in Authorization header
3. Go middleware extracts & validates JWT
4. If valid, handler proceeds; if invalid, returns 401
5. Frontend refreshes token if needed, retries request

## Cross-Cutting Concerns

### CORS
- Backend configured for frontend origin (localhost:3000)
- Credentials allowed (cookies)

### Error Handling
- Backend: service errors mapped to HTTP status codes
- Frontend: useQuery/useMutation handle errors, display to user

### Logging
- Backend: simple log.Printf() for startup events
- Frontend: console logs for dev debugging

## Deployment Considerations

- Backend & frontend are separate services
- Frontend env: `NEXT_PUBLIC_API_URL` points to backend
- Backend env: `JWT_SECRET` must be strong, random
- Database: Connection pooling configured for production
- Cookies: Secure flag set in production HTTPS environments
