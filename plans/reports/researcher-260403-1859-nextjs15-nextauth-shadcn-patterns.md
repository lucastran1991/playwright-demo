---
title: Next.js 15 Frontend Stack Research Report
date: 2026-04-03
scope: Next.js 15, NextAuth.js v5, TanStack Query v5, Shadcn/ui, React Hook Form + Zod
---

# Next.js 15 + NextAuth.js v5 + TanStack Query Frontend Patterns

## 1. Project Structure (/frontend)

```
frontend/
├── app/
│   ├── (auth)/
│   │   ├── login/page.tsx
│   │   ├── register/page.tsx
│   │   └── layout.tsx
│   ├── (dashboard)/
│   │   ├── layout.tsx (dashboard shell)
│   │   ├── page.tsx
│   │   └── [feature]/page.tsx
│   ├── api/auth/[...nextauth]/route.ts
│   ├── layout.tsx (root layout)
│   └── middleware.ts
├── components/
│   ├── ui/ (shadcn components)
│   ├── auth/ (login, register, session components)
│   └── forms/ (reusable form components)
├── lib/
│   ├── auth.ts (NextAuth config)
│   ├── api-client.ts (TanStack Query + fetch setup)
│   ├── query-client.ts (QueryClient instance)
│   ├── schemas/ (Zod schemas)
│   └── utils.ts
├── hooks/
│   ├── use-auth.ts (useSession wrapper)
│   └── use-api.ts (custom query/mutation hooks)
├── types/
│   └── index.ts (shared types)
└── public/

Key patterns:
- Server Components by default, "use client" only for interactivity
- API routes in app/api for edge-compatible auth
- Grouped routes with (parentheses) for layout organization
- No complexity until needed (YAGNI)
```

## 2. NextAuth.js v5 (Auth.js) Configuration

**File:** `lib/auth.ts`

```typescript
import NextAuth from "next-auth"
import CredentialsProvider from "next-auth/providers/credentials"

export const { auth, handlers, signIn, signOut } = NextAuth({
  providers: [
    CredentialsProvider({
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" }
      },
      async authorize(credentials) {
        // Call Go backend: POST /api/auth/login
        const res = await fetch("http://backend:8080/api/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(credentials)
        })
        if (!res.ok) return null
        const user = await res.json()
        return user // { id, email, name, etc }
      }
    })
  ],
  session: { strategy: "jwt" },
  callbacks: {
    jwt: async ({ token, user }) => {
      if (user) token.id = user.id
      return token
    },
    session: async ({ session, token }) => {
      if (session.user) session.user.id = token.id
      return session
    }
  },
  pages: {
    signIn: "/login",
    error: "/login"
  }
})
```

**Route handler:** `app/api/auth/[...nextauth]/route.ts`
```typescript
export { GET, POST } from "@/lib/auth"
```

**Env vars required:**
- `AUTH_SECRET` (generate via `npx auth secret`)
- `AUTH_URL` (e.g., `http://localhost:3000` for dev)

**Key insight:** JWT by default, no database required. Auth happens between Next.js frontend ↔ Go backend. NextAuth encrypts the JWT in a secure httpOnly cookie.

## 3. TanStack Query v5 Setup

**File:** `lib/query-client.ts`
```typescript
import { QueryClient } from "@tanstack/react-query"

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 5 * 60 * 1000 },
    mutations: { retry: 1 }
  }
})
```

**File:** `app/layout.tsx` (root)
```typescript
"use client"
import { QueryClientProvider } from "@tanstack/react-query"
import { queryClient } from "@/lib/query-client"

export default function RootLayout({ children }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}
```

**API client pattern:** `lib/api-client.ts`
```typescript
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

async function apiFetch(path: string, options: RequestInit = {}) {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers
    },
    credentials: "include" // for auth cookies
  })
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.json()
}

export { apiFetch }
```

**Query/Mutation pattern:** `hooks/queries.ts`
```typescript
import { useQuery, useMutation } from "@tanstack/react-query"

export const userKeys = {
  all: ["user"],
  detail: (id) => [...userKeys.all, id]
}

export function useUser(id: string) {
  return useQuery({
    queryKey: userKeys.detail(id),
    queryFn: () => apiFetch(`/api/users/${id}`)
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data) => apiFetch("/api/users", { method: "PUT", body: JSON.stringify(data) }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: userKeys.all })
  })
}
```

**Key insight:** Centralize query keys. Separate queries/mutations from components. Handle invalidation strategically (don't over-invalidate).

## 4. Shadcn/ui + Tailwind CSS v4 Theming

**Setup:** `components.json`
```json
{
  "tailwind": {
    "cssVariables": true,
    "prefix": ""
  }
}
```

**File:** `app/globals.css`
```css
@import "tailwindcss";

@theme {
  --color-background: oklch(0.97 0.003 247);
  --color-foreground: oklch(0.11 0.004 247);
  --color-primary: oklch(0.55 0.2 274);
  --color-primary-foreground: oklch(1 0 0);
  --color-secondary: oklch(0.69 0.06 300);
  --color-muted: oklch(0.92 0.004 247);
  --color-muted-foreground: oklch(0.45 0.02 247);
  --color-destructive: oklch(0.62 0.22 25);
  --radius: 0.5rem;
}

@dark {
  --color-background: oklch(0.12 0.004 247);
  --color-foreground: oklch(0.98 0.002 247);
  --color-primary: oklch(0.70 0.15 274);
  --color-primary-foreground: oklch(0.18 0.05 274);
  --color-secondary: oklch(0.56 0.08 300);
  --color-muted: oklch(0.24 0.005 247);
  --color-muted-foreground: oklch(0.65 0.02 247);
  --color-destructive: oklch(0.72 0.22 25);
}
```

Components auto-use tokens: `bg-background`, `text-foreground`, `bg-primary text-primary-foreground`. No custom class rewrites needed per theme change.

## 5. React Hook Form + Zod Validation

**File:** `lib/schemas/user-schema.ts`
```typescript
import { z } from "zod"

export const loginSchema = z.object({
  email: z.string().email("Invalid email"),
  password: z.string().min(8, "Min 8 chars")
})

export type LoginFormData = z.infer<typeof loginSchema>
```

**Component:** `components/forms/login-form.tsx`
```typescript
"use client"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { loginSchema } from "@/lib/schemas/user-schema"

export function LoginForm() {
  const form = useForm({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" }
  })

  async function onSubmit(data) {
    const result = await signIn("credentials", data)
    if (!result?.ok) form.setError("root", { message: "Auth failed" })
  }

  return (
    <form onSubmit={form.handleSubmit(onSubmit)}>
      <FormField
        control={form.control}
        name="email"
        render={({ field }) => (
          <FormItem>
            <FormLabel>Email</FormLabel>
            <FormControl>
              <Input {...field} type="email" />
            </FormControl>
            <FormMessage />
          </FormItem>
        )}
      />
      <Button type="submit">Sign In</Button>
    </form>
  )
}
```

**Key insight:** Schema = source of truth for client & server. Reuse in server actions for validation parity.

## 6. Authentication Flow (Frontend → Go Backend)

1. User submits login form (React Hook Form + Zod validation)
2. NextAuth CredentialsProvider calls Go backend `/api/auth/login`
3. Backend returns JWT + user data
4. NextAuth encrypts JWT in secure cookie
5. useSession() hooks read session from client
6. Protected routes via middleware.ts middleware at root

**Middleware pattern:** `middleware.ts`
```typescript
import { auth } from "@/lib/auth"

export async function middleware(request) {
  const session = await auth()
  if (!session && request.nextUrl.pathname.startsWith("/dashboard")) {
    return Response.redirect(new URL("/login", request.url))
  }
}

export const config = {
  matcher: ["/dashboard/:path*"]
}
```

---

## Key Takeaways

| Topic | Pattern | Why |
|-------|---------|-----|
| Structure | App Router, grouped routes, server-by-default | Minimal JS sent to client |
| Auth | NextAuth v5 JWT + Credentials Provider | Stateless, scales with Go backend, secure cookies |
| Data | TanStack Query with centralized keys | Deduplication, cache control, mutation invalidation |
| UI | Shadcn/ui with CSS variables | Token-based, easy theming, semantic naming |
| Forms | React Hook Form + Zod (shared schema) | Type-safe, minimal re-renders, DRY validation |

## Unresolved Questions

- Error handling strategy for auth token refresh/expiry?
- SSR prefetching with TanStack Query v5 (using dehydrate/hydrate)?
- Dark mode detection and persistence logic?

---

**Sources:**
- [Next.js 15 Project Structure](https://nextjs.org/docs/app/getting-started/project-structure)
- [Auth.js v5 Migration Guide](https://authjs.dev/getting-started/migrating-to-v5)
- [TanStack Query v5 Next.js Example](https://tanstack.com/query/v5/docs/framework/react/examples/nextjs)
- [shadcn/ui Theming](https://ui.shadcn.com/docs/theming)
- [React Hook Form + Zod Integration](https://ui.shadcn.com/docs/forms/react-hook-form)
- [Next.js 15 Best Practices (2026)](https://dev.to/ottoaria/nextjs-app-router-in-2026-the-complete-guide-for-full-stack-developers-5bjl)
- [Type-Safe Forms with Zod & React Hook Form](https://www.abstractapi.com/guides/email-validation/type-safe-form-validation-in-next-js-15-with-zod-and-react-hook-form)
