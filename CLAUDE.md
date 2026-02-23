# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Ender is a full-stack SMS gateway platform that allows sending SMS messages using registered devices (Android phones or modems) as gateways. Features include quota management, webhooks, multi-user support, and API key authentication.

## Commands

### Development (Docker Compose - Recommended)
```bash
docker compose up -d              # Start app
docker compose logs -f app        # View logs
docker compose down -v            # Clean up with volumes
```

### Backend
```bash
cd backend
go run . serve --http=0.0.0.0:8090   # Run dev server
go build -o ender .                   # Build binary
go build ./...                        # Verify compilation
```

### Frontend
```bash
cd frontend
fnm use                       # Switch to Node 24 (or nvm use)
npm install
npm run dev                   # Dev server at localhost:5173
npm run build                 # TypeScript + Vite build
npm run lint                  # Biome check with auto-fix

# E2E tests (requires backend running)
npx playwright test
npx playwright test --ui      # Interactive UI mode
```

## Architecture

### Backend (`backend/`)
- **Framework**: PocketBase (Go) - provides auth, CRUD, admin dashboard, cron, migrations
- **Database**: SQLite (embedded)
- **Key directories**:
  - `handlers/` - Custom API routes (sms, plans, webhooks)
  - `services/` - Business logic (SMS orchestration, FCM, quota, subscriptions, webhooks)
  - `services/payment/` - Payment provider abstraction (QvaPay)
  - `middleware/` - API key auth, maintenance mode
  - `migrations/` - PocketBase collection definitions + seed data
  - `main.go` - PocketBase init, record hooks, cron jobs, route registration

### Frontend (`frontend/`)
- **Framework**: React 19 + TypeScript + Vite
- **Routing/State**: TanStack Router + TanStack Query
- **Styling**: Tailwind CSS + shadcn/ui components
- **Key directories**:
  - `src/routes/` - Pages using TanStack Router file-based routing
  - `src/components/` - React components
  - `src/hooks/` - Custom React hooks using PocketBase JS SDK
  - `src/lib/pocketbase.ts` - PocketBase client instance

### Services Integration
- **FCM**: Firebase Cloud Messaging for push notifications to devices (via goroutines)
- **Payments**: QvaPay for subscription billing
- **Email**: PocketBase built-in SMTP, Mailcatcher for local dev at localhost:1080

## Design System

The Ender design system is defined in the **ender-homepage** repo (`../ender-homepage/src/pages/design-system.astro`) and documented at `/design-system` on the homepage site. It is the **single source of truth** for colors, typography, components, and patterns.

- **Reference**: `../ender-homepage/src/styles/global.css` — all CSS custom properties (colors, fonts, neutrals, code syntax)
- **Dashboard mapping**: `frontend/src/index.css` maps the same palette to shadcn/ui semantic variables
- **Fonts**: Inter (sans/body), Libre Baskerville (serif/headings), JetBrains Mono (mono/code) — loaded via Google Fonts in `frontend/index.html`
- **Accent**: `#2dd4a8` (mint/teal) — used consistently across both projects
- **Neutrals**: Mint-tinted gray scale (50–950), not standard Tailwind grays

When changing visual styles (colors, fonts, spacing, component patterns), update the design system page in ender-homepage **first**, then propagate changes to the dashboard's `frontend/src/index.css`.

## Code Quality Standards

### Backend
- Go standard formatting (`gofmt`)
- All code must compile: `go build ./...`

### Frontend
- Biome for linting and formatting
- TypeScript strict mode
- PocketBase JS SDK for all API calls (no auto-generated client)
- All visual styles must follow the design system (see above)

## Development URLs
| Service | URL |
|---------|-----|
| App (API + Frontend) | http://localhost:8090 |
| PocketBase Admin | http://localhost:8090/_/ |
| Frontend (dev server) | http://localhost:5173 |
| Mailcatcher | http://localhost:1080 |
