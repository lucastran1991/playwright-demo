# Codebase Summary

## Repository Overview

Fullstack application scaffold with Go Gin backend and Next.js 15 frontend. Clean architecture separation with JWT authentication, PostgreSQL persistence, and production-ready UI components.

**Generated:** 2026-04-03  
**Last Updated:** From repomix scan  
**Total Files:** 107 files  
**Primary Languages:** Go, TypeScript/TSX

## Directory Structure

```
playground-demo/
├── backend/                    # Go application
│   ├── cmd/server/
│   │   └── main.go            # Entry point, dependency wiring
│   ├── internal/              # Private packages
│   │   ├── config/
│   │   │   └── config.go      # Env var loading, validation
│   │   ├── database/
│   │   │   └── database.go    # PostgreSQL connection, migrations
│   │   ├── handler/
│   │   │   └── auth_handler.go# HTTP handlers (register, login, me)
│   │   ├── middleware/
│   │   │   └── auth_middleware.go # JWT validation
│   │   ├── model/
│   │   │   └── user.go        # GORM User model
│   │   ├── repository/
│   │   │   └── user_repository.go # Data access (CRUD)
│   │   ├── router/
│   │   │   └── router.go      # Route definitions, CORS setup
│   │   └── service/
│   │       └── auth_service.go # Business logic (register, login, refresh)
│   ├── pkg/                   # Reusable packages
│   │   ├── response/
│   │   │   └── response.go    # Standard JSON response helpers
│   │   └── token/
│   │       └── token.go       # JWT generation/validation
│   ├── .env.example           # Example environment config
│   ├── .gitignore
│   ├── go.mod                 # Go module definition
│   └── go.sum                 # Dependency checksums
│
├── frontend/                  # Next.js application
│   ├── src/
│   │   ├── app/              # App Router pages & layouts
│   │   │   ├── (auth)/       # Auth pages (grouped route)
│   │   │   │   ├── login/
│   │   │   │   │   └── page.tsx
│   │   │   │   ├── register/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── layout.tsx# Centered card layout
│   │   │   ├── (dashboard)/  # Protected pages (grouped route)
│   │   │   │   ├── layout.tsx# Sidebar + topbar layout
│   │   │   │   └── page.tsx  # Dashboard home
│   │   │   ├── api/
│   │   │   │   └── auth/[...nextauth]/
│   │   │   │       └── route.ts # NextAuth route handler
│   │   │   ├── globals.css   # Tailwind styles, theme variables
│   │   │   ├── layout.tsx    # Root layout (providers)
│   │   │   └── page.tsx      # Public landing page
│   │   ├── components/       # Reusable React components
│   │   │   ├── auth/
│   │   │   │   ├── login-form.tsx
│   │   │   │   └── register-form.tsx
│   │   │   ├── dashboard/
│   │   │   │   ├── sidebar-nav.tsx
│   │   │   │   ├── theme-toggle.tsx
│   │   │   │   ├── topbar.tsx
│   │   │   │   └── user-menu.tsx
│   │   │   └── ui/          # Shadcn/ui components
│   │   │       ├── avatar.tsx
│   │   │       ├── button.tsx
│   │   │       ├── card.tsx
│   │   │       ├── dropdown-menu.tsx
│   │   │       ├── input.tsx
│   │   │       ├── label.tsx
│   │   │       ├── separator.tsx
│   │   │       ├── sheet.tsx
│   │   │       ├── sidebar.tsx
│   │   │       ├── skeleton.tsx
│   │   │       └── tooltip.tsx
│   │   ├── hooks/           # Custom React hooks
│   │   │   ├── use-api.ts   # TanStack Query wrapper
│   │   │   ├── use-auth.ts  # NextAuth session wrapper
│   │   │   └── use-mobile.ts # Mobile viewport detection
│   │   ├── lib/             # Utilities & configuration
│   │   │   └── auth.ts      # NextAuth config (CredentialsProvider)
│   │   ├── providers/       # Context providers
│   │   │   ├── query-provider.tsx  # TanStack Query setup
│   │   │   ├── session-provider.tsx # NextAuth SessionProvider
│   │   │   └── theme-provider.tsx   # next-themes setup
│   │   ├── types/           # TypeScript definitions
│   │   │   ├── index.ts     # App types
│   │   │   └── next-auth.d.ts # NextAuth session type extension
│   │   ├── middleware.ts    # NextAuth middleware (protected routes)
│   │   └── favicon.ico
│   ├── public/              # Static assets
│   │   ├── file.svg
│   │   ├── globe.svg
│   │   ├── next.svg
│   │   ├── vercel.svg
│   │   └── window.svg
│   ├── .env.example
│   ├── .env.local
│   ├── .gitignore
│   ├── components.json      # Shadcn/ui config
│   ├── eslint.config.mjs
│   ├── next.config.ts
│   ├── package.json         # Dependencies
│   ├── pnpm-lock.yaml
│   ├── postcss.config.mjs
│   ├── tsconfig.json
│   ├── README.md
│   └── AGENTS.md
│
├── docs/                    # Documentation
│   ├── tech-stack.md        # Technology versions & stack
│   ├── design-guidelines.md # UI/UX guidelines (colors, typography)
│   ├── project-overview-pdr.md # Product requirements (this repo)
│   ├── system-architecture.md # Architecture & data flow (this repo)
│   ├── code-standards.md    # Code conventions (this repo)
│   ├── codebase-summary.md  # This file
│   └── wireframes/          # UI mockups
│       ├── dashboard.html
│       ├── login.html
│       └── register.html
│
├── plans/                   # Planning & research documents
│   ├── 260403-2123-fullstack-go-nextjs-scaffold/
│   │   ├── plan.md          # Overview of project phases
│   │   ├── phase-01-backend-project-setup.md
│   │   ├── phase-02-backend-database-models.md
│   │   ├── phase-03-backend-auth-api.md
│   │   ├── phase-04-frontend-project-setup.md
│   │   ├── phase-05-frontend-auth-pages.md
│   │   ├── phase-06-frontend-dashboard.md
│   │   └── reports/        # Detailed research & reviews
│   │       └── fullstack-developer-260403-2153-nextjs-frontend-setup.md
│   └── reports/            # Project-level reports
│       ├── researcher-260403-1859-go-gin-gorm-jwt-patterns.md
│       ├── researcher-260403-1859-nextjs15-nextauth-shadcn-patterns.md
│       └── ui-ux-designer-260403-2111-design-guidelines-wireframes.md
│
├── .gitignore              # Global git ignore
├── README.md               # Project intro & setup
└── repomix-output.xml      # Codebase compaction (this scan)
```

## Key Files & Responsibilities

### Backend Entry Points

**`backend/cmd/server/main.go`** (42 lines)
- Loads config from environment
- Connects to PostgreSQL
- Runs database migrations
- Wires dependencies (repository → service → handler)
- Starts Gin server on port 8080

### Backend Core Packages

**`internal/config/config.go`** (62 lines)
- Reads environment variables with fallbacks
- Validates required config (DB credentials, JWT secret)
- Returns `Config` struct
- Provides `DSN()` for PostgreSQL connection string

**`internal/database/database.go`** (36 lines)
- Opens PostgreSQL connection via GORM
- Configures connection pool (25 max, 5 idle)
- Runs `AutoMigrate(&model.User{})`
- Returns `*gorm.DB` instance

**`internal/model/user.go`** (Not shown, but referenced)
- GORM User model with fields: ID, Name, Email, Password, timestamps

**`internal/service/auth_service.go`** (123 lines)
- `Register(name, email, password)` - Creates user, hashes password, generates tokens
- `Login(email, password)` - Validates credentials, generates tokens
- `RefreshToken(refreshToken)` - Validates refresh token, issues new pair
- `GetUser(id)` - Fetches user by ID
- Handles error cases (user exists, invalid credentials, not found)

**`internal/repository/user_repository.go`** (Not shown, but referenced)
- `Create(user)` - INSERT user
- `FindByID(id)` - SELECT by ID
- `FindByEmail(email)` - SELECT by email

**`internal/handler/auth_handler.go`** (130 lines)
- `Register(c *gin.Context)` - POST /api/auth/register
- `Login(c *gin.Context)` - POST /api/auth/login
- `RefreshToken(c *gin.Context)` - POST /api/auth/refresh
- `Me(c *gin.Context)` - GET /api/auth/me (protected)
- Maps errors to HTTP status codes

**`internal/router/router.go`** (Not shown, but referenced)
- Gin engine setup with CORS middleware
- Routes: `/health`, `/api/auth/register`, `/api/auth/login`, `/api/auth/refresh`, `/api/auth/me`
- Applies auth middleware to protected routes

**`internal/middleware/auth_middleware.go`** (Not shown, but referenced)
- Validates JWT in Authorization header
- Extracts userID, stores in gin.Context
- Returns 401 if invalid/missing

**`pkg/token/token.go`** (Not shown, but referenced)
- `GenerateAccessToken(userID, email, secret)` - 15-minute JWT
- `GenerateRefreshToken(userID, email, secret)` - 7-day JWT
- `ValidateToken(token, secret)` - Verifies signature, returns claims

**`pkg/response/response.go`** (Not shown, but referenced)
- `Success(c, status, data)` - JSON response with data
- `Error(c, status, message)` - JSON error response

### Frontend Entry Points

**`src/app/layout.tsx`** (44 lines)
- Imports and configures Google fonts (DM Sans, JetBrains Mono)
- Wraps app with SessionProvider → ThemeProvider → QueryProvider
- Sets up HTML document structure with Tailwind classes

**`src/app/page.tsx`** (Not shown, but referenced)
- Public landing page (visible without auth)
- Likely has link to /login

**`src/middleware.ts`** (Not shown, but referenced)
- NextAuth middleware
- Protects `/dashboard/*` routes
- Redirects to `/login` if no session

### Auth Routes (`src/app/(auth)/`)

**`src/app/(auth)/layout.tsx`** (Not shown, but referenced)
- Centered card layout for login/register
- Card max-width: 420px
- Background fades/styling

**`src/app/(auth)/login/page.tsx`** (Not shown, but referenced)
- Renders LoginForm component
- Form submits via NextAuth signIn()

**`src/app/(auth)/register/page.tsx`** (Not shown, but referenced)
- Renders RegisterForm component
- Form calls /api/auth/register directly, then redirects to login

### Dashboard Routes (`src/app/(dashboard)/`)

**`src/app/(dashboard)/layout.tsx`** (Not shown, but referenced)
- Sidebar (responsive, icon-only on mobile)
- Topbar with breadcrumb, theme toggle, user menu
- Main content area (scrollable)

**`src/app/(dashboard)/page.tsx`** (Not shown, but referenced)
- Dashboard home page (protected)
- Displays user info

### API Routes

**`src/app/api/auth/[...nextauth]/route.ts`** (2 lines)
- Exports `{ GET, POST }` from `lib/auth`
- NextAuth route handler

**`src/lib/auth.ts`** (48 lines)
- NextAuth configuration
- CredentialsProvider: calls Go `/api/auth/login`
- JWT session strategy
- Callbacks: jwt() stores tokens, session() attaches to session object
- Pages: signIn at `/login`, error redirect to `/login`

### Authentication Components

**`src/components/auth/login-form.tsx`** (Not shown, but referenced)
- React Hook Form + Zod validation
- Email & password fields
- Calls NextAuth signIn() on submit
- Error & loading states

**`src/components/auth/register-form.tsx`** (Not shown, but referenced)
- React Hook Form + Zod validation
- Name, email, password fields
- POSTs to `/api/auth/register`
- Redirects to login on success

### Dashboard Components

**`src/components/dashboard/sidebar-nav.tsx`** (Not shown, but referenced)
- Navigation menu items
- Active state highlighting
- Icon + label layout

**`src/components/dashboard/topbar.tsx`** (Not shown, but referenced)
- Breadcrumb navigation
- Theme toggle button
- User menu dropdown

**`src/components/dashboard/user-menu.tsx`** (Not shown, but referenced)
- Dropdown with user name
- Logout button

**`src/components/dashboard/theme-toggle.tsx`** (Not shown, but referenced)
- Sun/moon icon toggle
- Calls `setTheme("light" | "dark")`

### Custom Hooks

**`src/hooks/use-auth.ts`** (Not shown, but referenced)
- Returns `useSession()` data
- Safe wrapper for session access

**`src/hooks/use-api.ts`** (Not shown, but referenced)
- Wraps TanStack useQuery
- Injects Bearer token in headers

**`src/hooks/use-mobile.ts`** (Not shown, but referenced)
- Media query check for mobile
- Used in sidebar (show as overlay on mobile)

### Providers

**`src/providers/session-provider.tsx`** (Not shown, but referenced)
- NextAuth SessionProvider wrapper

**`src/providers/theme-provider.tsx`** (Not shown, but referenced)
- next-themes ThemeProvider
- dark/light mode toggle support

**`src/providers/query-provider.tsx`** (Not shown, but referenced)
- TanStack Query QueryClientProvider
- Query caching & state management

### UI Components (Shadcn/ui)

Located in `src/components/ui/` - Pre-built, styled components:
- `button.tsx` - Button with variants
- `input.tsx` - Text input field
- `card.tsx` - Card container
- `dropdown-menu.tsx` - Dropdown menu
- `avatar.tsx` - User avatar circle
- `sidebar.tsx` - Collapsible sidebar
- `sheet.tsx` - Mobile sheet/drawer
- `separator.tsx` - Divider line
- `label.tsx` - Form label
- `tooltip.tsx` - Tooltip popup
- `breadcrumb.tsx` - Breadcrumb navigation
- `skeleton.tsx` - Loading placeholder

### Configuration Files

**`tsconfig.json`**
- TypeScript strict mode
- Path alias: `@/*` → `./src/*`
- Target: ES2020, Module: ESNext

**`next.config.ts`**
- App Router enabled
- TypeScript support

**`components.json`**
- Shadcn/ui configuration
- Component library paths
- Theme settings

**`eslint.config.mjs`**
- ESLint configuration
- JavaScript/TypeScript rules

**`postcss.config.mjs`**
- Tailwind CSS v4 plugin

**`package.json`**
- Dependencies: next, react, nextauth, tanstack-query, react-hook-form, zod, shadcn/ui, tailwindcss, next-themes
- Dev dependencies: typescript, eslint, tailwindcss, postcss

**`go.mod`**
- Module: `github.com/user/app`
- Go 1.22
- Dependencies: gin, gorm, postgres driver, jwt, bcrypt, godotenv, validator

## Data Model

### User (Backend GORM Model)
```
id (uint, primary key)
name (string)
email (string, unique)
password (string, bcrypt hashed)
created_at (timestamp)
updated_at (timestamp)
```

### Session (Frontend NextAuth)
```
user {
  id (string)
  email (string)
  name (string)
}
accessToken (string, JWT)
```

## Authentication Flow Diagram

```
1. User registers/logs in on frontend
2. Frontend form submits to Go backend (/api/auth/register or /api/auth/login)
3. Go backend validates, hashes password (bcrypt), generates JWT pair
4. Frontend receives tokens, passes to NextAuth
5. NextAuth stores tokens in encrypted HttpOnly cookie
6. Subsequent requests: NextAuth injects Bearer token in Authorization header
7. Go middleware extracts & validates JWT
8. Protected routes check middleware result, allow or deny access
9. Frontend can refresh tokens via NextAuth callback or manual /api/auth/refresh
```

## External Dependencies

**Backend:**
- Gin (HTTP framework)
- GORM (ORM)
- PostgreSQL driver
- JWT (golang-jwt/jwt/v5)
- bcrypt (password hashing)
- godotenv (config loading)

**Frontend:**
- Next.js 15 (framework)
- React 19 (UI library)
- NextAuth.js v5 (authentication)
- TanStack Query v5 (data fetching)
- React Hook Form v7 (form state)
- Zod v3 (validation)
- Tailwind CSS v4 (styling)
- Shadcn/ui (component library)
- next-themes (dark mode)

## Development Tools

- Go 1.22+
- Node.js 20+
- pnpm (package manager)
- PostgreSQL 15+
- golangci-lint (linting)
- ESLint (JavaScript linting)

## Notes

- Backend & frontend are separate deployable services
- Clean architecture separation (handler → service → repository)
- NextAuth CredentialsProvider pattern for custom auth backend
- Server components default in Next.js, "use client" for interactivity
- Tailwind CSS variables for theme (dark/light mode)
- No external auth service (Auth0, Firebase) - custom Go backend
