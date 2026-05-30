# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`grbac` is a multi-tenant RBAC (Role-Based Access Control) permission management platform. It serves as a centralized permission center for multiple external systems.

## Architecture

- **Backend**: Go (Gin framework), MySQL (GORM), Redis
- **Frontend**: Vue 3 + TypeScript, Element Plus, Pinia, Vite
- **Infrastructure**: Docker Compose (MySQL 8, Redis 7, Go server, Nginx)

## Key Concepts

- **System**: Top-level tenant. Each registered application has its own roles, permissions, and menus.
- **User → Role → Permission/Menu**: Classic RBAC model. Roles are scoped to a system.
- **SystemMember**: User-system membership with role "admin" (manages config) or "member" (regular).
- **Super Admin**: Global bypass (`is_super_admin=1`), can manage all systems and users.
- **External API**: Machine-to-machine access via `X-System-Code` header authentication.

## Backend Structure

```
backend/internal/
  handler/     → HTTP handlers (Gin)
  service/     → Business logic
  repository/  → Data access (GORM)
  model/       → Data models
  middleware/  → Auth, permission, audit, rate-limit, CORS, external auth
  pkg/         → Shared utilities (jwt, crypto, response, errors)
```

## Frontend Structure

```
frontend/src/
  api/         → Axios API service modules
  stores/      → Pinia stores (auth, user, system)
  views/       → Page components
  layouts/     → Shell layout (sidebar + header)
  utils/       → Request interceptor, token helpers, formatters
  styles/      → Design tokens, Element Plus overrides
```

## Development

```bash
# Start all services
docker-compose up -d

# Backend (local)
cd backend && go run cmd/server/main.go

# Frontend (local)
cd frontend && npm run dev  # port 3000, proxies /api to :8080
```

## Default Credentials

- Username: `admin`
- Password: `admin123`
