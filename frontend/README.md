# Vendel - Frontend

The frontend is built with [React 19](https://react.dev/), [TypeScript](https://www.typescriptlang.org/), [Vite](https://vitejs.dev/), [TanStack Query](https://tanstack.com/query), [TanStack Router](https://tanstack.com/router), and [Tailwind CSS](https://tailwindcss.com/) + [shadcn/ui](https://ui.shadcn.com/).

## Setup

Before you begin, ensure that you have either the Node Version Manager (nvm) or Fast Node Manager (fnm) installed.

```bash
cd frontend

# Switch to the correct Node.js version
fnm use   # or: nvm use

# Install dependencies
bun install
```

## Development

```bash
# Start dev server at http://localhost:5173
bun run dev

# Build for production
bun run build

# Lint and format (Biome)
bun run lint
```

The dev server connects to the PocketBase backend via the [PocketBase JS SDK](https://github.com/pocketbase/js-sdk). Set `VITE_API_URL` in `frontend/.env` to point to a different backend:

```env
VITE_API_URL=https://api.my-domain.example.com
```

## Code Structure

```
src/
├── routes/          # Pages (TanStack Router file-based routing)
├── components/      # React components (shadcn/ui)
├── hooks/           # Custom hooks (PocketBase SDK, realtime subscriptions)
└── lib/
    └── pocketbase.ts  # PocketBase client instance
```

## E2E Tests

End-to-end tests use [Playwright](https://playwright.dev/). The backend must be running.

```bash
# Run tests
bunx playwright test

# Interactive UI mode
bunx playwright test --ui
```
