# Vendel - Development

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

To view logs:

```bash
docker compose logs -f app
```

## Local Development

### Backend

```bash
cd backend

# Run dev server
go run . serve --http=0.0.0.0:8090

# Verify compilation
go build ./...

# Build binary
go build -o vendel .
./vendel serve --http=0.0.0.0:8090
```

The backend runs at `http://localhost:8090`. The PocketBase admin dashboard is available at `http://localhost:8090/_/`.

Collections are defined in `migrations/1740000000_initial.go`. With `ENVIRONMENT != production`, auto-migrations are enabled — changes made in the admin UI are automatically reflected in migration files.

### Frontend

```bash
cd frontend

fnm use          # or: nvm use

# Install dependencies
npm install

# Dev server (with HMR)
npm run dev

# Production build
npm run build

# Lint
npm run lint
```

The frontend runs at `http://localhost:5173` and connects to the backend via the PocketBase JS SDK.

### E2E Tests

Requires the backend running on `:8090`:

```bash
cd frontend
npx playwright test           # Run all tests
npx playwright test --ui      # Interactive UI mode
```

### Modem Agent (optional)

Only needed if you have physical USB modems to send/receive SMS:

```bash
cd modem-agent
cp .env.example .env    # Set your device API key + serial ports
go run .
```

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

## Environment Variables

Copy `.env.example` to `.env` at the project root. See the file for all available options.

| Variable | Description | Required |
|----------|-------------|----------|
| `ENVIRONMENT` | Runtime environment (`local`, `staging`, `production`) | No (default: `local`) |
| `FIRST_SUPERUSER` | Email for the initial admin account | Yes |
| `FIRST_SUPERUSER_PASSWORD` | Password for the initial admin account | Yes |
| `WEBHOOK_ENCRYPTION_KEY` | AES key for webhook secret encryption (random 32+ char string) | Yes |
| `FIREBASE_SERVICE_ACCOUNT_JSON` | Firebase service account JSON for FCM | No |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Google OAuth credentials | No |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | GitHub OAuth credentials | No |
| `QVAPAY_APP_ID` / `QVAPAY_APP_SECRET` | QvaPay payment credentials | No |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` | SMTP configuration | No |
| `LITESTREAM_REPLICA_URL` | S3 URL for continuous backup | No |
| `APP_URL` | Public app URL | No (default: `http://localhost:8090`) |
| `FRONTEND_URL` | Frontend URL | No (default: `http://localhost:5173`) |

## Development URLs

| Service | URL |
|---------|-----|
| App (API + Dashboard) | http://localhost:8090 |
| PocketBase Admin | http://localhost:8090/_/ |
| Frontend (Vite dev) | http://localhost:5173 |
| Mailcatcher | http://localhost:1080 |
