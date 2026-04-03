# Phase 5: Frontend Auth Pages

## Overview

- **Priority:** P1
- **Status:** completed
- **Description:** NextAuth.js v5 config, login/register pages with React Hook Form + Zod, middleware for route protection.

## Key Insights

- NextAuth v5 exports `auth`, `handlers`, `signIn`, `signOut` from single config
- CredentialsProvider calls Go backend `/api/auth/login`
- Store backend JWT in NextAuth session via callbacks
- Middleware protects `/dashboard/*` routes
- Auth pages: centered card, max-width 420px per design-guidelines.md
- Labels above inputs, `space-y-4` field spacing, full-width CTA

## Related Code Files

**Create:**
- `/frontend/src/lib/auth.ts` -- NextAuth v5 config
- `/frontend/src/lib/schemas/auth-schema.ts` -- Zod schemas
- `/frontend/src/app/api/auth/[...nextauth]/route.ts` -- NextAuth route handler
- `/frontend/src/app/(auth)/layout.tsx` -- centered auth layout
- `/frontend/src/app/(auth)/login/page.tsx` -- login page
- `/frontend/src/app/(auth)/register/page.tsx` -- register page
- `/frontend/src/components/auth/login-form.tsx` -- login form component
- `/frontend/src/components/auth/register-form.tsx` -- register form component
- `/frontend/src/middleware.ts` -- route protection

**Modify:**
- `/frontend/src/app/layout.tsx` -- add SessionProvider
- `/frontend/src/types/index.ts` -- add auth types

## Implementation Steps

### 1. Create Zod schemas (`lib/schemas/auth-schema.ts`)

```typescript
import { z } from "zod"

export const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
})

export const registerSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters").max(100),
  email: z.string().email("Invalid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

export type LoginFormData = z.infer<typeof loginSchema>
export type RegisterFormData = z.infer<typeof registerSchema>
```

### 2. Configure NextAuth v5 (`lib/auth.ts`)

```typescript
import NextAuth from "next-auth"
import CredentialsProvider from "next-auth/providers/credentials"

export const { auth, handlers, signIn, signOut } = NextAuth({
  providers: [
    CredentialsProvider({
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" },
      },
      async authorize(credentials) {
        const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            email: credentials.email,
            password: credentials.password,
          }),
        })
        if (!res.ok) return null
        const data = await res.json()
        // Return user + tokens for session storage
        return {
          id: String(data.user.id),
          email: data.user.email,
          name: data.user.name,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
        }
      },
    }),
  ],
  session: { strategy: "jwt" },
  callbacks: {
    jwt: async ({ token, user }) => {
      if (user) {
        token.id = user.id
        token.accessToken = user.accessToken
        token.refreshToken = user.refreshToken
      }
      return token
    },
    session: async ({ session, token }) => {
      if (session.user) {
        session.user.id = token.id as string
        session.accessToken = token.accessToken as string
      }
      return session
    },
  },
  pages: {
    signIn: "/login",
    error: "/login",
  },
})
```

### 3. Create route handler (`app/api/auth/[...nextauth]/route.ts`)

```typescript
import { handlers } from "@/lib/auth"
export const { GET, POST } = handlers
```

### 4. Extend types (`types/index.ts`)

```typescript
// Extend NextAuth types
declare module "next-auth" {
  interface Session {
    accessToken?: string
    user: { id: string; name: string; email: string }
  }
  interface User {
    accessToken?: string
    refreshToken?: string
  }
}
declare module "next-auth/jwt" {
  interface JWT {
    id?: string
    accessToken?: string
    refreshToken?: string
  }
}
```

### 5. Create auth layout (`app/(auth)/layout.tsx`)

- Centered flex container, min-h-screen
- Card wrapper, max-w-[420px]
- App name/logo at top

```tsx
export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <div className="w-full max-w-[420px]">{children}</div>
    </div>
  )
}
```

### 6. Create login form (`components/auth/login-form.tsx`)

- "use client" directive
- useForm with zodResolver(loginSchema)
- Shadcn Form, FormField, FormItem, FormLabel, FormControl, FormMessage
- Call signIn("credentials", { email, password, redirect: false })
- Handle error: set root form error
- On success: redirect to /dashboard via useRouter
- Full-width submit button
- Link to /register below form

### 7. Create login page (`app/(auth)/login/page.tsx`)

```tsx
import { LoginForm } from "@/components/auth/login-form"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export default function LoginPage() {
  return (
    <Card className="shadow-sm">
      <CardHeader><CardTitle className="text-center text-2xl">Sign In</CardTitle></CardHeader>
      <CardContent><LoginForm /></CardContent>
    </Card>
  )
}
```

### 8. Create register form (`components/auth/register-form.tsx`)

- Same pattern as login form but with registerSchema
- Fields: name, email, password, confirmPassword
- Call backend `/api/auth/register` directly via fetch
- On success: auto-sign in via signIn("credentials")
- Link to /login below form

### 9. Create register page (`app/(auth)/register/page.tsx`)

Same Card wrapper as login, with RegisterForm component.

### 10. Create middleware (`middleware.ts`)

```typescript
import { auth } from "@/lib/auth"
import { NextResponse } from "next/server"

export default auth((req) => {
  if (!req.auth && req.nextUrl.pathname.startsWith("/dashboard")) {
    return NextResponse.redirect(new URL("/login", req.url))
  }
})

export const config = {
  matcher: ["/dashboard/:path*"],
}
```

### 11. Add SessionProvider to root layout

Wrap children with `<SessionProvider>` from `next-auth/react` (client component wrapper needed).

## Todo List

- [x] Create Zod auth schemas
- [x] Configure NextAuth v5 with CredentialsProvider
- [x] Create NextAuth route handler
- [x] Extend NextAuth types for custom session fields
- [x] Create centered auth layout
- [x] Create login form with React Hook Form
- [x] Create login page
- [x] Create register form
- [x] Create register page
- [x] Create middleware for route protection
- [x] Add SessionProvider to root layout
- [x] Verify login/register flow against backend

## Success Criteria

- Login page renders at `/login` with form validation
- Register page renders at `/register` with password confirmation
- Successful login redirects to `/dashboard`
- Invalid credentials show error message
- Unauthenticated access to `/dashboard` redirects to `/login`
- Session persists across page refreshes
- Backend JWT stored in NextAuth session
