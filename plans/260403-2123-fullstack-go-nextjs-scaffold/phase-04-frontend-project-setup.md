# Phase 4: Frontend Project Setup

## Overview

- **Priority:** P1
- **Status:** completed
- **Description:** Initialize Next.js 15 with TypeScript, Shadcn/ui, Tailwind CSS v4, providers, and base layout.

## Key Insights

- Next.js 15 App Router: server components by default
- Tailwind v4 uses `@import "tailwindcss"` + `@theme` block (no tailwind.config.js)
- Shadcn/ui CLI initializes components.json and base components
- Design tokens from `docs/design-guidelines.md` -- deep teal primary, warm amber accent
- `src/` directory per tech-stack.md
- `next-themes` for dark mode toggle (class strategy)

## Related Code Files

**Create (via CLI + manual):**
- `/frontend/package.json`
- `/frontend/src/app/layout.tsx` -- root layout with providers
- `/frontend/src/app/page.tsx` -- landing/redirect
- `/frontend/src/app/globals.css` -- Tailwind + design tokens
- `/frontend/src/lib/utils.ts` -- cn() utility (Shadcn)
- `/frontend/src/providers/query-provider.tsx`
- `/frontend/src/providers/theme-provider.tsx`
- `/frontend/src/lib/api-client.ts` -- fetch wrapper
- `/frontend/src/lib/query-client.ts` -- QueryClient config
- `/frontend/src/types/index.ts` -- shared types
- `/frontend/.env.local`
- `/frontend/.env.example`

## Implementation Steps

### 1. Create Next.js project

```bash
cd /Users/mac/studio/playwright-demo
pnpm create next-app@latest frontend \
  --typescript --tailwind --eslint --app --src-dir \
  --import-alias "@/*" --use-pnpm
```

### 2. Install dependencies

```bash
cd frontend
pnpm add @tanstack/react-query @tanstack/react-query-devtools
pnpm add next-auth@beta
pnpm add react-hook-form @hookform/resolvers zod
pnpm add next-themes
```

### 3. Initialize Shadcn/ui

```bash
pnpm dlx shadcn@latest init
# Select: New York style, CSS variables, src/ path aliases
```

Add base components:
```bash
pnpm dlx shadcn@latest add button input label card form
```

### 4. Configure globals.css with design tokens

Replace default with design-guidelines.md values:

```css
@import "tailwindcss";

@theme {
  --color-background: oklch(0.985 0.002 250);
  --color-foreground: oklch(0.145 0.015 250);
  --color-card: oklch(0.995 0.001 250);
  --color-card-foreground: oklch(0.145 0.015 250);
  --color-primary: oklch(0.45 0.12 195);
  --color-primary-foreground: oklch(0.985 0.005 195);
  --color-secondary: oklch(0.94 0.01 250);
  --color-secondary-foreground: oklch(0.25 0.015 250);
  --color-muted: oklch(0.94 0.008 250);
  --color-muted-foreground: oklch(0.55 0.015 250);
  --color-accent: oklch(0.75 0.15 75);
  --color-accent-foreground: oklch(0.25 0.05 75);
  --color-destructive: oklch(0.55 0.2 25);
  --color-destructive-foreground: oklch(0.985 0.005 25);
  --color-border: oklch(0.88 0.008 250);
  --color-input: oklch(0.88 0.008 250);
  --color-ring: oklch(0.45 0.12 195);
  --color-sidebar-background: oklch(0.97 0.005 250);
  --color-sidebar-foreground: oklch(0.25 0.015 250);
  --radius: 0.75rem;
  --font-sans: "DM Sans", ui-sans-serif, system-ui, sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, monospace;
}

.dark {
  --color-background: oklch(0.13 0.015 250);
  --color-foreground: oklch(0.93 0.005 250);
  /* ... all dark mode tokens from design-guidelines.md */
}
```

### 5. Create providers

**theme-provider.tsx:**
```tsx
"use client"
import { ThemeProvider as NextThemesProvider } from "next-themes"
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return <NextThemesProvider attribute="class" defaultTheme="system" enableSystem>
    {children}
  </NextThemesProvider>
}
```

**query-provider.tsx:**
```tsx
"use client"
import { QueryClientProvider } from "@tanstack/react-query"
import { queryClient } from "@/lib/query-client"
export function QueryProvider({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
}
```

### 6. Create root layout

```tsx
// src/app/layout.tsx
import { ThemeProvider } from "@/providers/theme-provider"
import { QueryProvider } from "@/providers/query-provider"
import { DM_Sans, JetBrains_Mono } from "next/font/google"
import "./globals.css"

const dmSans = DM_Sans({ subsets: ["latin", "vietnamese"], variable: "--font-sans" })
const jetbrainsMono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono" })

export default function RootLayout({ children }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={`${dmSans.variable} ${jetbrainsMono.variable} font-sans`}>
        <ThemeProvider>
          <QueryProvider>{children}</QueryProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
```

### 7. Create api-client.ts and query-client.ts

**api-client.ts:** fetch wrapper with base URL, JSON headers, error handling
**query-client.ts:** QueryClient with 5min staleTime default

### 8. Create types/index.ts

```typescript
export interface User {
  id: number
  name: string
  email: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: User
}
```

### 9. Create .env.example

```
NEXT_PUBLIC_API_URL=http://localhost:8080
AUTH_SECRET=generate-with-npx-auth-secret
AUTH_URL=http://localhost:3000
```

## Todo List

- [x] Create Next.js project with pnpm
- [x] Install all dependencies
- [x] Initialize Shadcn/ui + base components
- [x] Configure globals.css with design tokens
- [x] Create theme provider
- [x] Create query provider
- [x] Setup root layout with fonts and providers
- [x] Create API client and query client
- [x] Create shared types
- [x] Create .env files
- [x] Verify `pnpm dev` runs on :3000

## Success Criteria

- `pnpm dev` starts on localhost:3000 without errors
- Design tokens applied (teal primary visible)
- Dark/light mode toggle works
- No TypeScript errors
- Shadcn components render correctly
