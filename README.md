# Fullstack Go + Next.js App

Fullstack application scaffold with Go (Gin/GORM/PostgreSQL/JWT) backend and Next.js 15 (Shadcn/ui, TanStack Query, NextAuth.js v5) frontend.

## Tech Stack

### Backend (`/backend`)
- **Go** with Gin web framework
- **GORM** ORM with PostgreSQL
- **JWT** authentication (golang-jwt/jwt/v5)
- **bcrypt** password hashing

### Frontend (`/frontend`)
- **Next.js 15** (App Router, TypeScript)
- **Shadcn/ui** + Tailwind CSS v4
- **TanStack Query v5** for data fetching
- **React Hook Form** + Zod validation
- **NextAuth.js v5** for session management
- **next-themes** for dark/light mode

## Prerequisites

- Go 1.22+
- Node.js 20+
- pnpm
- PostgreSQL 15+

## Getting Started

### 1. Database Setup

```bash
createdb app_dev
```

### 2. Backend

```bash
cd backend
cp .env.example .env
# Edit .env with your PostgreSQL credentials and JWT secret
go run ./cmd/server
```

Server starts on `http://localhost:8080`.

### 3. Frontend

```bash
cd frontend
cp .env.example .env.local
# Edit .env.local if needed
pnpm install
pnpm dev
```

App starts on `http://localhost:3000`.

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/api/auth/register` | No | Register user |
| POST | `/api/auth/login` | No | Login |
| POST | `/api/auth/refresh` | No | Refresh tokens |
| GET | `/api/auth/me` | Yes | Get current user |

## Project Structure

```
backend/
  cmd/server/         Entry point
  internal/
    config/           Environment config
    database/         GORM connection + migrations
    handler/          HTTP handlers
    middleware/       JWT auth middleware
    model/            Database models
    repository/       Data access layer
    service/          Business logic
    router/           Route definitions
  pkg/
    response/         Standard JSON responses
    token/            JWT utilities

frontend/src/
  app/
    (auth)/           Login, register pages
    (dashboard)/      Dashboard (protected)
    api/auth/         NextAuth route handler
  components/
    auth/             Auth form components
    dashboard/        Sidebar, topbar, menus
    ui/               Shadcn/ui components
  hooks/              useAuth, useApi
  lib/                Auth config, API client, schemas
  providers/          Theme, Query, Session providers
  types/              TypeScript types
```
