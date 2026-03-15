# Development Guide

## Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [Node.js](https://nodejs.org/) 24+ (via [fnm](https://github.com/Schniz/fnm) or [nvm](https://github.com/nvm-sh/nvm))
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose (optional, for containerized dev)

## Quick Start (Docker Compose)

The fastest way to get everything running:

```bash
cp .env.example .env    # Edit with your values
docker compose up -d
```

The app will be available at `http://localhost:8090`.

To include the local mail server for testing emails:

```bash
docker compose --profile dev up -d
```

## Local Development (without Docker)

### Backend

```bash
cd backend
cp ../.env.example ../.env   # Edit with your values

go run . serve --http=0.0.0.0:8090
```

On first run, PocketBase will auto-migrate the database schema and seed data. The backend serves the API at `http://localhost:8090` and the PocketBase admin dashboard at `http://localhost:8090/_/`.

### Frontend

In a separate terminal:

```bash
cd frontend
fnm use          # or: nvm use
npm install
npm run dev
```

The frontend dev server runs at `http://localhost:5173` with hot module replacement. It proxies API requests to the backend at `:8090`.

### Modem Agent (optional)

Only needed if you have physical USB modems to send/receive SMS:

```bash
cd modem-agent
cp .env.example .env    # Set your device API key + serial ports
go run .
```

## Environment Variables

Copy `.env.example` to `.env` at the project root. Required variables:

| Variable | Description |
|---|---|
| `FIRST_SUPERUSER` | Email for the initial admin account |
| `FIRST_SUPERUSER_PASSWORD` | Password for the initial admin account |
| `WEBHOOK_ENCRYPTION_KEY` | AES key for webhook secret encryption (random 32+ char string) |

All other variables (OAuth, FCM, payments, SMTP, Litestream) are optional for local development.

## Project Structure

```
vendel/
├── backend/              # Go (PocketBase) API server
│   ├── main.go           # Thin wiring layer (~80 lines)
│   ├── config.go         # App bootstrap helpers
│   ├── hooks/            # Record lifecycle hooks (one file per domain)
│   ├── cronjobs/         # Periodic background tasks
│   ├── handlers/         # Custom API routes
│   ├── services/         # Business logic
│   ├── middleware/        # Request middleware
│   └── migrations/       # DB schema + seed data
├── frontend/             # React 19 + TypeScript + Vite
│   ├── src/routes/       # TanStack Router file-based pages
│   ├── src/components/   # UI components
│   ├── src/hooks/        # Custom hooks (PocketBase SDK)
│   ├── src/types/        # TypeScript interfaces
│   └── src/lib/          # Utilities and constants
├── modem-agent/          # Standalone Go agent for USB modems
├── docker-compose.yml
└── Dockerfile
```

## Useful Commands

### Backend

```bash
cd backend
go build ./...                          # Verify compilation
go run . serve --http=0.0.0.0:8090      # Run dev server
```

### Frontend

```bash
cd frontend
npm run dev       # Dev server with HMR
npm run build     # TypeScript check + production build
npm run lint      # Biome linter with auto-fix
```

### E2E Tests

Requires the backend running on `:8090`:

```bash
cd frontend
npx playwright test           # Run all tests
npx playwright test --ui      # Interactive UI mode
```

### Docker Compose

```bash
docker compose up -d                              # Start app
docker compose --profile dev up -d                # Start with Mailcatcher
docker compose --profile modem up -d              # Start with modem agent
docker compose logs -f app                        # View app logs
docker compose down -v                            # Clean up with volumes
```

## Development URLs

| Service | URL |
|---|---|
| App (API + Dashboard) | http://localhost:8090 |
| PocketBase Admin | http://localhost:8090/_/ |
| Frontend (Vite dev) | http://localhost:5173 |
| Mailcatcher | http://localhost:1080 |
