# Project Overview & PDR

## Product Overview

Fullstack web application scaffold combining Go (backend) and Next.js 15 (frontend) with production-ready authentication, dashboard UI, and dark mode support.

## Target Users
- Developers bootstrapping fullstack projects
- Teams needing JWT + NextAuth integration patterns
- Projects requiring clean architecture separation

## Core Features

### Authentication System
- User registration (name, email, password)
- Login with email/password credentials
- JWT token refresh flow
- Protected endpoints with Bearer token validation
- Session management via NextAuth.js

### Dashboard
- Protected layout with sidebar navigation
- Responsive design (mobile/tablet/desktop)
- User profile menu in top bar
- Dark/light mode toggle

### Design System
- Shadcn/ui component library
- Tailwind CSS v4 with OKLCH color space
- Custom color palette (teal primary, amber accent)
- DM Sans + JetBrains Mono typography
- Consistent spacing & layout rules

## Functional Requirements

| ID | Requirement | Priority | Status |
|---|---|---|---|
| FR1 | User registration with validation | High | Complete |
| FR2 | Login/logout with JWT tokens | High | Complete |
| FR3 | Token refresh mechanism | High | Complete |
| FR4 | Get current user endpoint | High | Complete |
| FR5 | Protected dashboard layout | High | Complete |
| FR6 | Dark/light mode toggle | Medium | Complete |
| FR7 | Responsive sidebar navigation | Medium | Complete |

## Non-Functional Requirements

| ID | Requirement | Metric | Target |
|---|---|---|---|
| NFR1 | Password hashing | Algorithm | bcrypt (cost 10) |
| NFR2 | Auth token expiry | Duration | Access: 15m, Refresh: 7d |
| NFR3 | Database pool | Connections | 25 max, 5 idle |
| NFR4 | CORS policy | Headers | Frontend origin allowed |
| NFR5 | API response format | Structure | Consistent JSON + status codes |

## Technical Stack

### Backend
- **Language:** Go 1.22+
- **Web Framework:** Gin v1.10+
- **ORM:** GORM v1.25+
- **Database:** PostgreSQL 15+
- **Auth:** JWT (golang-jwt/jwt/v5)
- **Password:** bcrypt (golang.org/x/crypto)

### Frontend
- **Framework:** Next.js 15 (App Router)
- **Language:** TypeScript 5+
- **UI Library:** Shadcn/ui
- **Styling:** Tailwind CSS v4
- **Data Fetching:** TanStack Query v5
- **Forms:** React Hook Form v7
- **Validation:** Zod v3
- **Auth:** NextAuth.js v5

## Architecture Highlights

- **Backend:** Clean architecture (handler → service → repository)
- **Frontend:** Server components by default, "use client" only for interactivity
- **Auth Flow:** NextAuth CredentialsProvider → Go backend API → JWT in encrypted cookie
- **API:** REST JSON, Bearer token in Authorization header

## Success Criteria

- All auth endpoints functional & tested
- Frontend routes protected appropriately
- Dark mode toggle works across all pages
- Dashboard responsive on mobile/tablet/desktop
- Code follows project standards & conventions
- Documentation complete and accurate

## Constraints & Dependencies

- PostgreSQL 15+ required (not SQLite)
- Node.js 20+ for frontend
- Go 1.22+ for backend
- pnpm for package management
- Environment variables must be set before startup

## Open Questions

None at this time.
