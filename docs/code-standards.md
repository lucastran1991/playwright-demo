# Code Standards & Conventions

## Go Backend Standards

### File Naming
- **Convention:** `snake_case.go`
- **Examples:**
  - `auth_handler.go` - Handler for auth endpoints
  - `user_repository.go` - Repository for user data access
  - `auth_service.go` - Service for auth logic
  - `auth_middleware.go` - Middleware for auth validation

### Package Organization
```
internal/          # Private packages (internal to app)
  config/          # Configuration loading & validation
  database/        # DB connection & migrations
  handler/         # HTTP handlers (controllers)
  middleware/      # HTTP middleware (auth, CORS)
  model/           # GORM data models
  repository/      # Data access layer
  router/          # Route definitions
  service/         # Business logic

pkg/               # Public/reusable packages
  response/        # Standard response helpers
  token/           # JWT utilities
```

### Code Style
- **Formatting:** `go fmt` (enforced)
- **Linting:** `golangci-lint` recommended
- **Error Handling:** Always check errors explicitly
  ```go
  if err != nil {
    log.Printf("error: %v", err)
    return nil, err
  }
  ```
- **Comments:** Export package/function docs (uppercase)
  ```go
  // AuthService handles authentication business logic.
  type AuthService struct { ... }

  // Register creates a new user and returns tokens.
  func (s *AuthService) Register(...) { ... }
  ```

### Architecture Patterns

**Clean Architecture (Handler → Service → Repository)**
```
Handler
  ↓
Service (business logic, validation)
  ↓
Repository (data access, GORM)
  ↓
Model (GORM struct, database schema)
```

**Dependency Injection**
- Wire dependencies in `main.go`
- Pass repositories to services
- Pass services to handlers
- Never import across layers upward

**Error Types**
- Define service-level errors as package variables
  ```go
  var (
    ErrUserExists = errors.New("user already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
  )
  ```
- Use `errors.Is()` for type checking in handlers

**Request/Response**
- Use structs with JSON tags for binding
  ```go
  type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
  }
  ```
- Use Gin response helpers: `c.JSON()`, `c.Error()`
- Validate input in handler, not in service

### Testing
- Unit test files: `*_test.go` in same package
- Test functions: `TestFunctionName(t *testing.T)`
- Table-driven tests for multiple cases

## Next.js Frontend Standards

### File Naming
- **Convention:** `kebab-case.tsx` / `kebab-case.ts`
- **Examples:**
  - `login-form.tsx` - Login form component
  - `use-auth.ts` - Custom hook for auth
  - `theme-provider.tsx` - Theme context provider

### Directory Structure
```
src/
  app/              # Next.js App Router pages & layouts
    (auth)/         # Grouped route for auth pages
      login/
        page.tsx
      register/
        page.tsx
      layout.tsx    # Shared auth layout
    (dashboard)/    # Grouped route for protected pages
      layout.tsx
      page.tsx
    api/
      auth/
        [...nextauth]/
          route.ts
    layout.tsx      # Root layout

  components/       # Reusable React components
    auth/           # Auth-specific components
    dashboard/      # Dashboard-specific components
    ui/             # Shadcn/ui or generic UI components

  hooks/            # Custom React hooks
    use-auth.ts
    use-api.ts
    use-mobile.ts

  lib/              # Utilities & config
    auth.ts         # NextAuth configuration
    api-client.ts   # API fetching utilities
    schemas.ts      # Zod validation schemas

  providers/        # Context providers
    session-provider.tsx
    theme-provider.tsx
    query-provider.tsx

  types/            # TypeScript type definitions
    index.ts        # App types
    next-auth.d.ts  # NextAuth types

  middleware.ts     # NextAuth middleware
```

### Component Patterns

**Server Components (Default)**
- Use for data fetching, static content
- Can access databases, secrets directly
- No "use client" directive

**Client Components**
- Add "use client" directive at top
- Use for interactivity: forms, event handlers, hooks
- Example:
  ```tsx
  "use client"
  
  import { useState } from "react"
  
  export function Counter() {
    const [count, setCount] = useState(0)
    return <button onClick={() => setCount(count + 1)}>{count}</button>
  }
  ```

**Props & TypeScript**
- Define prop types explicitly
  ```tsx
  interface LoginFormProps {
    onSubmit: (email: string, password: string) => Promise<void>
    isLoading?: boolean
  }
  
  export function LoginForm({ onSubmit, isLoading }: LoginFormProps) { ... }
  ```

### Forms with React Hook Form + Zod

```tsx
"use client"

import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"

const schema = z.object({
  email: z.string().email("Invalid email"),
  password: z.string().min(8, "Min 8 characters"),
})

type FormData = z.infer<typeof schema>

export function MyForm() {
  const form = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      {/* form fields */}
    </form>
  )
}
```

### Custom Hooks

**useAuth** - Wrapper for NextAuth session
```tsx
export function useAuth() {
  const { data: session } = useSession()
  return { session, isAuthenticated: !!session }
}
```

**useApi** - Wrapper for TanStack Query with Bearer token
```tsx
export function useApi<T>(url: string) {
  const { data: session } = useSession()
  
  return useQuery({
    queryKey: [url],
    queryFn: async () => {
      const res = await fetch(url, {
        headers: {
          Authorization: `Bearer ${session?.accessToken}`,
        },
      })
      return res.json() as T
    },
  })
}
```

### Styling

**Tailwind CSS v4**
- Use Tailwind utility classes
- Custom colors via CSS variables
- Dark mode: class-based (set on `<html>`)
- Responsive breakpoints: `sm:`, `md:`, `lg:`, `xl:`

**CSS Modules** (if needed)
- Filename: `component.module.css`
- Import: `import styles from "./component.module.css"`

### API Fetching

**Client-side (useQuery)**
```tsx
const { data, isLoading, error } = useQuery({
  queryKey: ["user"],
  queryFn: async () => {
    const res = await fetch("/api/user", {
      headers: { Authorization: `Bearer ${token}` },
    })
    return res.json()
  },
})
```

**Server-side (fetch in server component)**
```tsx
async function getData() {
  const res = await fetch("http://localhost:8080/api/data", {
    headers: { Authorization: `Bearer ${token}` },
    next: { revalidate: 60 }, // ISR
  })
  return res.json()
}

export default async function Page() {
  const data = await getData()
  return <div>{data}</div>
}
```

### Type Safety

**Shared Types** (`src/types/index.ts`)
```tsx
export interface User {
  id: string
  name: string
  email: string
}

export interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: User
}
```

**NextAuth Extension** (`src/types/next-auth.d.ts`)
```tsx
declare module "next-auth" {
  interface Session {
    user: {
      id: string
      email: string
      name: string
    }
    accessToken: string
  }
}
```

## Shared Conventions

### Error Handling

**Backend**
- Service layer throws domain errors
- Handler layer maps to HTTP status codes
- Return consistent error JSON

**Frontend**
- Catch errors in useQuery/useMutation
- Display user-friendly messages
- Log errors to console in dev

### Naming Conventions

| Item | Convention | Example |
|------|-----------|---------|
| Go functions | PascalCase (exported), camelCase (private) | `GetUser()`, `getUserByID()` |
| Go variables | camelCase | `userID`, `authService` |
| Go types | PascalCase | `AuthService`, `User` |
| TS/JS variables | camelCase | `userName`, `isLoading` |
| TS/JS functions | camelCase | `getUserData()` |
| TS/JS types/interfaces | PascalCase | `UserResponse`, `AuthContext` |
| File names (Go) | snake_case | `auth_handler.go` |
| File names (TS/JS) | kebab-case | `user-menu.tsx` |
| Environment vars | UPPER_SNAKE_CASE | `JWT_SECRET`, `DATABASE_URL` |
| CSS classes | kebab-case | `.auth-form`, `.sidebar-nav` |

### Comments

**Go**
- Exported symbols must have doc comments
- Inline comments for non-obvious logic
- No commented-out code (use git history)

**TypeScript/JavaScript**
- JSDoc for exported functions/types
- Inline comments for complex logic
- TODO comments with explanation
  ```tsx
  // TODO: Add pagination when list grows past 100 items
  ```

## Code Review Checklist

- [ ] Does code follow naming conventions?
- [ ] Are errors handled explicitly (Go) or caught (TS)?
- [ ] Are types defined for all function parameters?
- [ ] Are comments clear and necessary?
- [ ] Is code DRY (no duplication)?
- [ ] Does implementation match architecture patterns?
- [ ] Are security concerns addressed (auth, validation)?
- [ ] Are tests included (for Go) or stories (for TS)?

## Tools & Configuration

**Go**
- `go fmt` for formatting
- `go vet` for common errors
- `.golangci.yml` for linting config

**Frontend**
- `eslint.config.mjs` for JS/TS linting
- `prettier` for code formatting
- `tsconfig.json` for TypeScript config

**Both**
- `.gitignore` excludes build artifacts, env files
- Environment variables in `.env` (never committed)
