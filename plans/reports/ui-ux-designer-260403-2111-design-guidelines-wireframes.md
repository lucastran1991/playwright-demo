# Design Guidelines & Wireframes Report

## Summary

Created complete design system and wireframes for Go + Next.js 15 fullstack scaffold with JWT auth.

## Deliverables

### 1. Design Guidelines (`/docs/design-guidelines.md`)
- **Already existed** with comprehensive 150-line spec
- Color palette: deep teal primary + warm amber accent, oklch color space, light/dark mode
- Typography: DM Sans (body) + JetBrains Mono (code), Vietnamese support
- Spacing: 4px base unit, Tailwind default scale
- Border radius: `rounded-xl` cards, `rounded-lg` buttons/inputs, `rounded-full` badges
- Component patterns for auth pages, dashboard, forms, buttons
- Shadcn CSS variable mapping for both themes

### 2. Wireframes (`/docs/wireframes/`)

| File | Status | Description |
|------|--------|-------------|
| `login.html` | Already existed | Centered card, email/password, social login, dark mode toggle |
| `register.html` | Already existed | Same layout + name, confirm password, terms checkbox |
| `dashboard.html` | **Created** | Full dashboard with sidebar, topbar, stat cards, chart placeholder, activity feed, users table |

### Dashboard Wireframe Details
- **Sidebar**: 256px desktop, collapsible to 64px (icon-only), mobile sheet overlay with backdrop
- **Top bar**: 64px height, hamburger (mobile), collapse toggle (desktop), breadcrumb, theme/notification/avatar
- **Content**: 4-col stat cards grid, 2/3+1/3 chart+activity layout, responsive data table
- **Responsive**: mobile-first, grid adapts at sm/md/lg breakpoints
- **Dark mode**: functional toggle, all elements styled for both themes
- **Interactive**: JS for sidebar collapse, mobile menu open/close, dark mode toggle

## Design Decisions
- Chose teal+amber palette for personality (vs default Shadcn zinc) -- teal conveys trust/professionalism, amber adds warmth
- DM Sans over Inter/Poppins -- geometric, modern, excellent Vietnamese support, not overused
- Sidebar pattern matches Shadcn sidebar component API for easy implementation
- All wireframes use Tailwind CDN for standalone browser preview

## Files
- `/Users/mac/studio/playwright-demo/docs/design-guidelines.md`
- `/Users/mac/studio/playwright-demo/docs/wireframes/login.html`
- `/Users/mac/studio/playwright-demo/docs/wireframes/register.html`
- `/Users/mac/studio/playwright-demo/docs/wireframes/dashboard.html`
