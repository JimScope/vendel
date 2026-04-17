#!/usr/bin/env bash
# Run Vendel backend and frontend in development mode.
# Usage: ./scripts/run-development.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

if [ ! -d "$BACKEND_DIR" ] || [ ! -d "$FRONTEND_DIR" ]; then
  echo "error: expected backend/ and frontend/ in $ROOT_DIR" >&2
  exit 1
fi

command -v go >/dev/null 2>&1 || { echo "error: go not found in PATH" >&2; exit 1; }
command -v bun >/dev/null 2>&1 || { echo "error: bun not found in PATH" >&2; exit 1; }

pids=()

cleanup() {
  echo ""
  echo "→ stopping dev servers..."
  for pid in "${pids[@]}"; do
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
    fi
  done
  wait 2>/dev/null || true
}
trap cleanup INT TERM EXIT

echo "→ installing frontend dependencies (if needed)..."
(cd "$FRONTEND_DIR" && bun install --silent)

echo "→ starting backend on :8090..."
(cd "$BACKEND_DIR" && go run . serve --http=0.0.0.0:8090) &
pids+=($!)

echo "→ starting frontend on :5173..."
(cd "$FRONTEND_DIR" && bun run dev) &
pids+=($!)

echo ""
echo "Backend:  http://localhost:8090"
echo "Admin:    http://localhost:8090/_/"
echo "Frontend: http://localhost:5173"
echo ""
echo "Press Ctrl+C to stop both servers."

wait -n
