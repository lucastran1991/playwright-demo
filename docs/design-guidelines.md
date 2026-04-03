# Design Guidelines

## Color Palette

Primary: deep teal. Accent: warm amber. Neutral: slate with blue undertone.
All values in oklch for perceptual uniformity.

### Light Mode (CSS Variables)
```css
--background: oklch(0.985 0.002 250);        /* near-white with cool tint */
--foreground: oklch(0.145 0.015 250);        /* deep slate */
--card: oklch(0.995 0.001 250);              /* white card */
--card-foreground: oklch(0.145 0.015 250);
--popover: oklch(0.995 0.001 250);
--popover-foreground: oklch(0.145 0.015 250);
--primary: oklch(0.45 0.12 195);             /* deep teal */
--primary-foreground: oklch(0.985 0.005 195);
--secondary: oklch(0.94 0.01 250);           /* light slate */
--secondary-foreground: oklch(0.25 0.015 250);
--muted: oklch(0.94 0.008 250);
--muted-foreground: oklch(0.55 0.015 250);
--accent: oklch(0.75 0.15 75);              /* warm amber */
--accent-foreground: oklch(0.25 0.05 75);
--destructive: oklch(0.55 0.2 25);          /* red */
--destructive-foreground: oklch(0.985 0.005 25);
--border: oklch(0.88 0.008 250);
--input: oklch(0.88 0.008 250);
--ring: oklch(0.45 0.12 195);               /* matches primary */
--sidebar-background: oklch(0.97 0.005 250);
--sidebar-foreground: oklch(0.25 0.015 250);
--sidebar-accent: oklch(0.92 0.015 195);
--sidebar-accent-foreground: oklch(0.20 0.02 195);
--sidebar-border: oklch(0.90 0.008 250);
```

### Dark Mode
```css
--background: oklch(0.13 0.015 250);
--foreground: oklch(0.93 0.005 250);
--card: oklch(0.17 0.015 250);
--card-foreground: oklch(0.93 0.005 250);
--popover: oklch(0.17 0.015 250);
--popover-foreground: oklch(0.93 0.005 250);
--primary: oklch(0.65 0.14 195);            /* lighter teal */
--primary-foreground: oklch(0.13 0.02 195);
--secondary: oklch(0.22 0.015 250);
--secondary-foreground: oklch(0.88 0.005 250);
--muted: oklch(0.22 0.012 250);
--muted-foreground: oklch(0.65 0.01 250);
--accent: oklch(0.72 0.14 75);
--accent-foreground: oklch(0.15 0.05 75);
--destructive: oklch(0.60 0.2 25);
--destructive-foreground: oklch(0.95 0.005 25);
--border: oklch(0.27 0.012 250);
--input: oklch(0.27 0.012 250);
--ring: oklch(0.65 0.14 195);
--sidebar-background: oklch(0.11 0.015 250);
--sidebar-foreground: oklch(0.88 0.005 250);
--sidebar-accent: oklch(0.20 0.02 195);
--sidebar-accent-foreground: oklch(0.88 0.01 195);
--sidebar-border: oklch(0.22 0.012 250);
```

## Typography

**Primary font:** `DM Sans` (Google Fonts) -- geometric, modern, supports Vietnamese
**Mono font:** `JetBrains Mono` (Google Fonts) -- for code/data

| Element | Size | Weight | Line Height |
|---------|------|--------|-------------|
| h1 | 2.25rem (36px) | 700 | 1.2 |
| h2 | 1.75rem (28px) | 600 | 1.3 |
| h3 | 1.25rem (20px) | 600 | 1.4 |
| body | 0.9375rem (15px) | 400 | 1.6 |
| small / caption | 0.8125rem (13px) | 400 | 1.5 |
| button | 0.875rem (14px) | 500 | 1 |

## Spacing & Layout

- Base unit: 4px. Use Tailwind's default scale (multiples of 4).
- Page max-width: 1280px (`max-w-7xl`), centered.
- Content padding: `px-4` mobile, `px-6` tablet, `px-8` desktop.
- Section gap: `gap-6` (24px) default.
- Card padding: `p-6` desktop, `p-4` mobile.

## Border Radius

| Element | Radius |
|---------|--------|
| Card / Dialog | `rounded-xl` (12px) |
| Button | `rounded-lg` (8px) |
| Input | `rounded-lg` (8px) |
| Badge / Chip | `rounded-full` |
| Avatar | `rounded-full` |

`--radius: 0.75rem` (Shadcn base)

## Shadows

- Card: `shadow-sm` light, `shadow-none` dark (rely on border).
- Dropdown/popover: `shadow-lg`.
- Keep shadows subtle; dark mode uses borders instead.

## Component Patterns

### Auth Pages (Login / Register)
- Centered card layout, max-width 420px
- Logo/app name at top of card
- Social login buttons (optional) above divider
- Form fields with labels above inputs
- Primary CTA button full-width
- Link to alternate auth page below card

### Dashboard
- Sidebar: 256px wide desktop, collapsible to icon-only (64px), hidden on mobile (sheet overlay)
- Top bar: 64px height, contains breadcrumb, theme toggle, user avatar menu
- Main content area: scrollable, padded
- Sidebar nav items: icon + label, active state with `sidebar-accent` bg

### Forms
- Use Shadcn `Form` with React Hook Form + Zod
- Labels above inputs (not floating)
- Error messages below input in `destructive` color
- Consistent field spacing: `space-y-4`

### Buttons
- Primary: filled with `primary` color
- Secondary: filled with `secondary` color
- Outline: border only, transparent bg
- Ghost: no border, no bg, text only
- Destructive: filled with `destructive` color
- Minimum height: 40px (touch target)

### Responsive Breakpoints
- `sm`: 640px -- phone landscape
- `md`: 768px -- tablet
- `lg`: 1024px -- desktop
- `xl`: 1280px -- wide desktop

### Dark Mode
- Toggle via `next-themes` provider
- Class-based strategy (`class` on `<html>`)
- All colors swap via CSS variables -- no conditional classes needed

### Animation
- Transitions: 150ms ease default
- Page transitions: subtle fade (optional)
- Respect `prefers-reduced-motion`
- Sidebar collapse: 200ms ease-in-out
