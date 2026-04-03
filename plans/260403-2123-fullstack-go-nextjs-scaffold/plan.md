---
title: "Fullstack Go + Next.js Scaffold"
description: "Scaffold fullstack app with Go Gin backend and Next.js 15 frontend"
status: completed
priority: P1
effort: 6h
branch: main
tags: [fullstack, backend, frontend, auth, scaffold]
created: 2026-04-03
completed: 2026-04-03
---

# Fullstack Go + Next.js Scaffold

## Overview

Scaffold a fullstack application with Go (Gin/GORM/JWT) backend and Next.js 15 (Shadcn/ui, TanStack Query, NextAuth v5) frontend. Backend on :8080, frontend on :3000.

## Research Reports

- [Go Patterns](../reports/researcher-260403-1859-go-gin-gorm-jwt-patterns.md)
- [Next.js Patterns](../reports/researcher-260403-1859-nextjs15-nextauth-shadcn-patterns.md)
- [Tech Stack](../../docs/tech-stack.md)
- [Design Guidelines](../../docs/design-guidelines.md)

## Phases

| # | Phase | Status | Effort |
|---|-------|--------|--------|
| 1 | [Backend Project Setup](phase-01-backend-project-setup.md) | completed | 45min |
| 2 | [Backend Database Models](phase-02-backend-database-models.md) | completed | 30min |
| 3 | [Backend Auth API](phase-03-backend-auth-api.md) | completed | 1.5h |
| 4 | [Frontend Project Setup](phase-04-frontend-project-setup.md) | completed | 45min |
| 5 | [Frontend Auth Pages](phase-05-frontend-auth-pages.md) | completed | 1h |
| 6 | [Frontend Dashboard](phase-06-frontend-dashboard.md) | completed | 1.5h |

## Dependencies

```
Phase 1 → Phase 2 → Phase 3 (backend sequential)
Phase 4 → Phase 5 → Phase 6 (frontend sequential)
Backend & frontend tracks can run in parallel.
Phase 5 depends on Phase 3 (auth endpoints).
```

## Key Decisions

- Simple JWT with `golang-jwt/jwt/v5` (not appleboy/gin-jwt)
- AutoMigrate for dev (no migration tool)
- No Docker, CI/CD, or testing setup
- NextAuth v5 CredentialsProvider calls Go backend
- Design tokens from `docs/design-guidelines.md`
