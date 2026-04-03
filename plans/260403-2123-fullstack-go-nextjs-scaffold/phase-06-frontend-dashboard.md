# Phase 6: Frontend Dashboard

## Overview

- **Priority:** P1
- **Status:** completed
- **Description:** Dashboard layout (sidebar + topbar), protected routes, TanStack Query integration, theme toggle.

## Key Insights

- Sidebar: 256px desktop, collapsible to 64px icons, mobile sheet overlay (design-guidelines.md)
- Topbar: 64px height, breadcrumb, theme toggle, user avatar menu
- Use Shadcn Sidebar component (`pnpm dlx shadcn@latest add sidebar-07`)
- TanStack Query hooks for authenticated API calls (pass accessToken from session)
- Server components by default, "use client" only for interactive parts

## Related Code Files

**Create:**
- `/frontend/src/app/(dashboard)/layout.tsx` -- dashboard shell
- `/frontend/src/app/(dashboard)/page.tsx` -- dashboard home
- `/frontend/src/components/dashboard/sidebar-nav.tsx` -- sidebar navigation
- `/frontend/src/components/dashboard/topbar.tsx` -- top bar with user menu
- `/frontend/src/components/dashboard/user-menu.tsx` -- avatar dropdown
- `/frontend/src/components/dashboard/theme-toggle.tsx` -- dark/light toggle
- `/frontend/src/hooks/use-auth.ts` -- useSession wrapper + auth helpers
- `/frontend/src/hooks/use-api.ts` -- authenticated TanStack Query hooks
- `/frontend/src/lib/api-client.ts` -- update with auth header support

**Shadcn components to add:**
```bash
pnpm dlx shadcn@latest add avatar dropdown-menu separator sheet sidebar breadcrumb
```

## Implementation Steps

### 1. Add Shadcn components

```bash
pnpm dlx shadcn@latest add avatar dropdown-menu separator sheet sidebar breadcrumb tooltip
```

### 2. Create dashboard layout (`app/(dashboard)/layout.tsx`)

```tsx
import { SidebarNav } from "@/components/dashboard/sidebar-nav"
import { Topbar } from "@/components/dashboard/topbar"
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar"

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider>
      <SidebarNav />
      <SidebarInset>
        <Topbar />
        <main className="flex-1 overflow-auto p-6">{children}</main>
      </SidebarInset>
    </SidebarProvider>
  )
}
```

### 3. Create sidebar navigation (`components/dashboard/sidebar-nav.tsx`)

- "use client"
- Use Shadcn Sidebar, SidebarContent, SidebarGroup, SidebarMenuItem
- Nav items: Dashboard (Home icon), Settings (placeholder)
- Active state with `sidebar-accent` bg color
- Collapsible: icon-only at 64px width
- App name/logo in SidebarHeader
- Responsive: sheet overlay on mobile via SidebarTrigger

```tsx
const navItems = [
  { title: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { title: "Settings", href: "/dashboard/settings", icon: Settings },
]
```

### 4. Create topbar (`components/dashboard/topbar.tsx`)

- Height: 64px (`h-16`)
- Left: SidebarTrigger + Breadcrumb
- Right: ThemeToggle + UserMenu
- Use Shadcn Breadcrumb component
- Sticky top, border-bottom

```tsx
export function Topbar() {
  return (
    <header className="sticky top-0 z-10 flex h-16 items-center justify-between border-b bg-background px-4">
      <div className="flex items-center gap-2">
        <SidebarTrigger />
        <Separator orientation="vertical" className="h-6" />
        <Breadcrumb>...</Breadcrumb>
      </div>
      <div className="flex items-center gap-2">
        <ThemeToggle />
        <UserMenu />
      </div>
    </header>
  )
}
```

### 5. Create theme toggle (`components/dashboard/theme-toggle.tsx`)

- "use client"
- `useTheme()` from next-themes
- Shadcn DropdownMenu or simple button toggle
- Sun/Moon icons

### 6. Create user menu (`components/dashboard/user-menu.tsx`)

- "use client"
- `useSession()` from next-auth/react
- Shadcn DropdownMenu with Avatar trigger
- Items: user email (disabled), separator, Sign Out
- Sign out calls `signOut({ callbackUrl: "/login" })`

### 7. Create auth hook (`hooks/use-auth.ts`)

```typescript
"use client"
import { useSession } from "next-auth/react"

export function useAuth() {
  const { data: session, status } = useSession()
  return {
    user: session?.user,
    accessToken: session?.accessToken,
    isAuthenticated: status === "authenticated",
    isLoading: status === "loading",
  }
}
```

### 8. Create API hooks (`hooks/use-api.ts`)

```typescript
import { useQuery } from "@tanstack/react-query"
import { useAuth } from "./use-auth"
import { apiFetch } from "@/lib/api-client"

export function useCurrentUser() {
  const { accessToken } = useAuth()
  return useQuery({
    queryKey: ["user", "me"],
    queryFn: () => apiFetch("/api/auth/me", {
      headers: { Authorization: `Bearer ${accessToken}` },
    }),
    enabled: !!accessToken,
  })
}
```

### 9. Update api-client.ts for auth support

Add optional `headers` merge to apiFetch so callers can pass Authorization header.

### 10. Create dashboard home page (`app/(dashboard)/page.tsx`)

- Welcome message with user name from useCurrentUser()
- Simple card-based layout showing placeholder stats
- Demonstrates TanStack Query integration works

```tsx
"use client"
import { useCurrentUser } from "@/hooks/use-api"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export default function DashboardPage() {
  const { data: user, isLoading } = useCurrentUser()
  if (isLoading) return <div>Loading...</div>
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Welcome, {user?.name}</h1>
      <div className="grid gap-6 md:grid-cols-3">
        <Card><CardHeader><CardTitle>Card 1</CardTitle></CardHeader>
          <CardContent>Placeholder</CardContent></Card>
        {/* more cards */}
      </div>
    </div>
  )
}
```

## Todo List

- [x] Add Shadcn sidebar + utility components
- [x] Create dashboard layout with sidebar + topbar
- [x] Implement sidebar navigation with nav items
- [x] Implement topbar with breadcrumb, theme toggle, user menu
- [x] Create theme toggle component
- [x] Create user menu with sign out
- [x] Create useAuth hook
- [x] Create TanStack Query hooks for authenticated API calls
- [x] Update api-client with auth header support
- [x] Create dashboard home page
- [x] Verify sidebar collapse/expand works
- [x] Verify mobile responsive (sheet overlay)
- [x] Verify theme toggle persists

## Success Criteria

- Dashboard renders at `/dashboard` with sidebar + topbar
- Sidebar shows nav items with active state highlighting
- Sidebar collapses to icon-only mode
- Mobile: sidebar opens as sheet overlay
- Theme toggle switches dark/light mode
- User menu shows email and sign-out option
- Sign out redirects to login page
- TanStack Query fetches user data with auth token
- Page is only accessible when authenticated
