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
- Environment variables: `DB_*`, `JWT_SECRET`, `SERVER_PORT`, `BLUEPRINT_DIR`, `MODEL_DIR`
  - `BLUEPRINT_DIR`: Path to blueprint CSV files (default: `./blueprint/Node & Edge`)
  - `MODEL_DIR`: Path to model metadata CSVs (default: `./blueprint`)

### Database Layer
- **`internal/database/database.go`** - PostgreSQL connection pooling, migrations
- **`internal/model/user.go`** - GORM model definition
- **`internal/model/capacity_node_type.go`** - Capacity domain metadata
- **`internal/model/dependency_rule.go`** - Upstream/local dependency rules
- **`internal/model/impact_rule.go`** - Downstream/load impact rules

### Repository Layer (Data Access)
- **`internal/repository/user_repository.go`** - CRUD operations on users
  - Methods: `Create()`, `FindByID()`, `FindByEmail()`
- **`internal/repository/tracer_repository.go`** - Dependency tracing with recursive CTEs
  - Methods: `FindUpstreamNodes()`, `FindDownstreamNodes()`, `FindLocalNodes()`
  - Queries: Capacity node types, dependency rules, impact rules
- Uses GORM ORM

### Service Layer (Business Logic)
- **`internal/service/auth_service.go`** - Authentication operations
  - Methods: `Register()`, `Login()`, `RefreshToken()`, `GetUser()`
  - Handles password hashing, token generation, validation
- **`internal/service/model_csv_parser.go`** - Parses model metadata CSVs
  - Methods: `ParseCapacityNodesCSV()`, `ParseDependenciesCSV()`, `ParseImpactsCSV()`
- **`internal/service/model_ingestion_service.go`** - Ingests model metadata into DB
  - Methods: `IngestAll()`
  - Transactions for idempotent upsert
- **`internal/service/dependency_tracer.go`** - Traces node dependencies and impacts
  - Methods: `TraceDependencies()`, `TraceImpacts()`
  - Groups results by topology and level

### Handler Layer (HTTP)
- **`internal/handler/auth_handler.go`** - HTTP request/response handling
  - Maps HTTP requests to service calls
  - Input validation, error responses
- **`internal/handler/tracer_handler.go`** - Model ingestion + dependency tracing HTTP handlers
  - Endpoints: model ingestion, capacity node listing, dependency/impact tracing

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

## Blueprint CSV Ingestion Feature

### Overview
Ingests hierarchical blueprint data from CSV files. Blueprint types (domains like Cooling, Electrical) contain nodes (entities) connected by edges (relationships). Supports recursive tree traversal.

### Models
- **BlueprintType** - Category/domain (e.g., Cooling, Electrical)
  - Fields: `id`, `name`, `slug`, `folder_name`, `created_at`, `updated_at`
- **BlueprintNode** - Entity within a type
  - Fields: `id`, `node_id`, `blueprint_type_id`, `created_at`, `updated_at`
  - `node_id`: Unique identifier per blueprint type
- **BlueprintNodeMembership** - Hierarchical parent-child relationship
  - Fields: `id`, `parent_node_id`, `child_node_id`, `created_at`
- **BlueprintEdge** - Connection between two nodes
  - Fields: `id`, `source_node_id`, `target_node_id`, `blueprint_type_id`, `created_at`

### CSV Format
Blueprint data stored in `./blueprint/Node & Edge/` directory:
```
{type}/
  ├── Nodes.csv        # node_id
  ├── Edges.csv        # source_node_id, target_node_id
  └── Hierarchy.csv    # parent_node_id, child_node_id
```

### API Endpoints

**Blueprint Ingestion (Protected)**
- `POST /api/blueprints/ingest` - Trigger full CSV ingestion, returns summary

**Blueprint Read (Public)**
- `GET /api/blueprints/types` - List all blueprint domains
- `GET /api/blueprints/nodes?type=slug&limit=20&offset=0` - List nodes with type filter
- `GET /api/blueprints/nodes/:nodeId` - Get single node + memberships
- `GET /api/blueprints/edges?type=slug&limit=20&offset=0` - List edges for type
- `GET /api/blueprints/tree/:typeSlug` - Recursive tree structure

**Model Ingestion (Protected)**
- `POST /api/models/ingest` - Ingest capacity nodes, dependency rules, impact rules

**Model & Trace Read (Public)**
- `GET /api/models/capacity-nodes` - List all capacity node types with constraints
- `GET /api/trace/dependencies/:nodeId?levels=2&include_local=true` - Upstream dependencies of node
- `GET /api/trace/impacts/:nodeId?levels=2&load_scope=` - Downstream impacts of node

### Ingestion Service
**`internal/service/blueprint_ingestion_service.go`**
- `IngestAll(dir string)` - Orchestrates parsing & persistence
- Reads all blueprint type folders
- Parses Nodes, Edges, Hierarchy CSVs
- Saves to database with transaction consistency
- Returns summary: types count, nodes count, edges count

### CSV Parser
**`internal/service/blueprint_csv_parser.go`**
- `ParseNodes(file, typeID)` - Extracts node_id rows
- `ParseEdges(file, typeID)` - Extracts source/target edges
- `ParseHierarchy(file)` - Extracts parent/child memberships
- Handles malformed CSV gracefully with validation

## Capacity Nodes, Dependency & Impact Rules Feature

### Overview
Ingests metadata that defines capacity domains, upstream/local dependencies, and downstream/load impacts across system topologies. Enables querying which nodes depend on or impact a given node.

### Models
- **CapacityNodeType** - Capacity domain classification
  - Fields: `id`, `node_type`, `topology`, `is_capacity_node`, `active_constraint`, `created_at`, `updated_at`
- **DependencyRule** - Type-level upstream/local dependencies
  - Fields: `id`, `node_type`, `dependency_node_type`, `relationship_type`, `topological_relationship`, `upstream_level`
- **ImpactRule** - Type-level downstream/load impacts
  - Fields: `id`, `node_type`, `impact_node_type`, `topological_relationship`, `downstream_level`

### CSV Format
Model data stored in `./blueprint/` directory:
```
├── Capacity Nodes.csv   # node_type, topology, is_capacity_node, active_constraint
├── Dependencies.csv     # node_type, dependency_node_type, relationship_type, topological_relationship, upstream_level
└── Impacts.csv          # node_type, impact_node_type, topological_relationship, downstream_level
```

### Ingestion Service
**`internal/service/model_ingestion_service.go`**
- `IngestAll(modelDir string)` - Orchestrates parsing & persistence
- Parses 3 model CSVs (Capacity Nodes, Dependencies, Impacts)
- Idempotent: re-running produces same state via ON CONFLICT upsert

### Tracing Service
**`internal/service/dependency_tracer.go`**
- `TraceDependencies(nodeID, maxLevels, includeLocal)` - Resolves upstream dependencies from rules + CTE queries
- `TraceImpacts(nodeID, maxLevels, loadScope)` - Resolves downstream impacts
- Groups results by topology and hop level
- Filters results to only nodes matching rule target types

### Tracing Repository
**`internal/repository/tracer_repository.go`**
- Recursive CTE queries against blueprint_edges topology
- `FindUpstreamNodes()` - Parent walk (follows edge.from_node_id)
- `FindDownstreamNodes()` - Child walk (follows edge.to_node_id)
- `FindLocalNodes()` - Direct edge neighbors
- Rule lookups: `GetDependencyRules()`, `GetImpactRules()`, `ListCapacityNodeTypes()`

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
