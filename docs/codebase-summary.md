# Codebase Summary

## Repository Overview

Fullstack application scaffold with Go Gin backend and Next.js 15 frontend. Clean architecture separation with JWT authentication, PostgreSQL persistence, and production-ready UI components.

**Generated:** 2026-04-03  
**Last Updated:** 2026-04-04 (Added capacity nodes, dependency/impact rules, tracer API)  
**Total Files:** 120+ files (with tracer feature)  
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

**`internal/model/blueprint_type.go`**
- BlueprintType model: ID, Name (unique), Slug (unique), FolderName, timestamps

**`internal/model/blueprint_node.go`**
- BlueprintNode model: ID, NodeID (unique per type), BlueprintTypeID (FK), timestamps

**`internal/model/blueprint_node_membership.go`**
- BlueprintNodeMembership model: ID, ParentNodeID (FK), ChildNodeID (FK), timestamps
- Represents hierarchical parent-child relationships

**`internal/model/blueprint_edge.go`**
- BlueprintEdge model: ID, SourceNodeID (FK), TargetNodeID (FK), BlueprintTypeID (FK), timestamps
- Represents directed edges between nodes

**`internal/model/capacity_node_type.go`**
- CapacityNodeType model: ID, NodeType, Topology, IsCapacityNode, ActiveConstraint, timestamps
- Metadata: which node types are capacity domains across topologies

**`internal/model/dependency_rule.go`**
- DependencyRule model: ID, NodeType, DependencyNodeType, RelationshipType, TopologicalRelationship, UpstreamLevel, timestamps
- Type-level upstream/local dependencies (composite unique key on node_type + dependency_node_type)

**`internal/model/impact_rule.go`**
- ImpactRule model: ID, NodeType, ImpactNodeType, TopologicalRelationship, DownstreamLevel, timestamps
- Type-level downstream/load impacts (composite unique key on node_type + impact_node_type)

**`internal/service/auth_service.go`** (123 lines)
- `Register(name, email, password)` - Creates user, hashes password, generates tokens
- `Login(email, password)` - Validates credentials, generates tokens
- `RefreshToken(refreshToken)` - Validates refresh token, issues new pair
- `GetUser(id)` - Fetches user by ID
- Handles error cases (user exists, invalid credentials, not found)

**`internal/service/blueprint_ingestion_service.go`**
- `IngestAll(dir string)` - Orchestrates CSV parsing for all blueprint types
- Reads blueprint directory, detects type folders
- Parses Nodes, Edges, Hierarchy CSVs per type
- Returns summary: types count, nodes count, edges count

**`internal/service/blueprint_csv_parser.go`**
- `ParseNodes(file, typeID)` - Extracts node_id rows from Nodes.csv
- `ParseEdges(file, typeID)` - Extracts source/target edges from Edges.csv
- `ParseHierarchy(file)` - Extracts parent/child memberships from Hierarchy.csv
- Handles CSV parsing with validation
- `ReadCSV(filePath)` - Exported helper for CSV file reading

**`internal/service/model_csv_parser.go`**
- `ParseCapacityNodesCSV(filePath)` - Parses Capacity Nodes.csv (24 rows)
- `ParseDependenciesCSV(filePath)` - Parses Dependencies.csv (147 rows)
- `ParseImpactsCSV(filePath)` - Parses Impacts.csv (118 rows)
- Handles nullable integers (upstream_level, downstream_level)
- Handles "True"/"False" string to bool conversion

**`internal/service/model_ingestion_service.go`**
- `IngestAll(modelDir)` - Orchestrates ingestion of 3 model CSVs
- Idempotent upsert: ON CONFLICT DO UPDATE per table
- Returns summary: capacity nodes, dependency rules, impact rules upserted

**`internal/service/dependency_tracer.go`**
- `TraceDependencies(nodeID, maxLevels, includeLocal)` - Resolves upstream dependencies
- `TraceImpacts(nodeID, maxLevels, loadScope)` - Resolves downstream impacts
- Groups results by topology and hop level
- Filters to only nodes matching rule target types

**`internal/repository/user_repository.go`** (Not shown, but referenced)
- `Create(user)` - INSERT user
- `FindByID(id)` - SELECT by ID
- `FindByEmail(email)` - SELECT by email

**`internal/repository/blueprint_repository.go`**
- `ListTypes()` - Retrieves all blueprint types (domains)
- `ListNodes(typeSlug, limit, offset)` - Retrieves nodes with pagination, optional type filter
- `GetNodeByNodeID(nodeId)` - Retrieves single node + its memberships
- `ListEdges(typeSlug, limit, offset)` - Retrieves edges for type with pagination
- `GetTree(typeSlug)` - Recursive tree traversal from root nodes
- `SaveTypes(types)` - Persists blueprint types
- `SaveNodes(nodes)` - Persists blueprint nodes
- `SaveEdges(edges)` - Persists blueprint edges
- `SaveMemberships(memberships)` - Persists node memberships

**`internal/repository/tracer_repository.go`**
- `FindUpstreamNodes(sourceDBID, typeSlug, maxLevel)` - Recursive CTE: parent walk
- `FindDownstreamNodes(sourceDBID, typeSlug, maxLevel)` - Recursive CTE: child walk
- `FindLocalNodes(sourceDBID, typeSlug)` - Direct edge neighbors
- `ListCapacityNodeTypes()` - All capacity node types
- `GetDependencyRules(nodeType)` - Rules for given node type
- `GetImpactRules(nodeType)` - Impact rules for given node type

**`internal/handler/auth_handler.go`** (130 lines)
- `Register(c *gin.Context)` - POST /api/auth/register
- `Login(c *gin.Context)` - POST /api/auth/login
- `RefreshToken(c *gin.Context)` - POST /api/auth/refresh
- `Me(c *gin.Context)` - GET /api/auth/me (protected)
- Maps errors to HTTP status codes

**`internal/handler/blueprint_handler.go`**
- `Ingest(c *gin.Context)` - POST /api/blueprints/ingest (protected) - Triggers CSV ingestion
- `ListTypes(c *gin.Context)` - GET /api/blueprints/types - Lists all blueprint domains
- `ListNodes(c *gin.Context)` - GET /api/blueprints/nodes - Lists nodes with type filter
- `GetNode(c *gin.Context)` - GET /api/blueprints/nodes/:nodeId - Single node + memberships
- `ListEdges(c *gin.Context)` - GET /api/blueprints/edges - Lists edges for type
- `GetTree(c *gin.Context)` - GET /api/blueprints/tree/:typeSlug - Recursive tree structure

**`internal/handler/tracer_handler.go`**
- `IngestModels(c *gin.Context)` - POST /api/models/ingest (protected) - Triggers model CSV ingestion
- `ListCapacityNodes(c *gin.Context)` - GET /api/models/capacity-nodes - Lists all capacity node types
- `TraceDependencies(c *gin.Context)` - GET /api/trace/dependencies/:nodeId - Upstream dependencies
- `TraceImpacts(c *gin.Context)` - GET /api/trace/impacts/:nodeId - Downstream impacts

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

### Blueprint Type
```
id (uint, primary key)
name (string, unique)
slug (string, unique)
folder_name (string)
created_at (timestamp)
updated_at (timestamp)
```

### Blueprint Node
```
id (uint, primary key)
node_id (string)
blueprint_type_id (uint, FK to blueprint_types)
created_at (timestamp)
updated_at (timestamp)
```

### Blueprint Node Membership
```
id (uint, primary key)
parent_node_id (uint, FK to blueprint_nodes)
child_node_id (uint, FK to blueprint_nodes)
created_at (timestamp)
```

### Blueprint Edge
```
id (uint, primary key)
source_node_id (uint, FK to blueprint_nodes)
target_node_id (uint, FK to blueprint_nodes)
blueprint_type_id (uint, FK to blueprint_types)
created_at (timestamp)
```

### Capacity Node Type
```
id (uint, primary key)
node_type (string, unique)
topology (string)
is_capacity_node (bool)
active_constraint (bool)
created_at (timestamp)
updated_at (timestamp)
```

### Dependency Rule
```
id (uint, primary key)
node_type (string, composite unique)
dependency_node_type (string, composite unique)
relationship_type (string)
topological_relationship (string) -- "Upstream" or "Local"
upstream_level (int, nullable)
created_at (timestamp)
updated_at (timestamp)
```

### Impact Rule
```
id (uint, primary key)
node_type (string, composite unique)
impact_node_type (string, composite unique)
topological_relationship (string) -- "Downstream" or "Load"
downstream_level (int, nullable)
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
- encoding/csv (CSV parsing)
- database/sql (raw SQL for recursive CTEs)

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
