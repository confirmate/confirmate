#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UI_PORT=5173
API_PORT=8080

cleanup() {
  echo ""
  echo "Shutting down demo..."
  kill "${PID_CONFIRMATE:-}" "${PID_UI:-}" "${PID_COLLECTOR:-}" 2>/dev/null || true
  wait 2>/dev/null || true
  echo "Done."
}
trap cleanup EXIT INT TERM

# ── Intro ─────────────────────────────────────────────────────────────────────
cat <<'EOF'

   ██████╗ ██████╗ ███╗   ██╗███████╗██╗██████╗ ███╗   ███╗ █████╗ ████████╗███████╗
  ██╔════╝██╔═══██╗████╗  ██║██╔════╝██║██╔══██╗████╗ ████║██╔══██╗╚══██╔══╝██╔════╝
  ██║     ██║   ██║██╔██╗ ██║█████╗  ██║██████╔╝██╔████╔██║███████║   ██║   █████╗
  ██║     ██║   ██║██║╚██╗██║██╔══╝  ██║██╔══██╗██║╚██╔╝██║██╔══██║   ██║   ██╔══╝
  ╚██████╗╚██████╔╝██║ ╚████║██║     ██║██║  ██║██║ ╚═╝ ██║██║  ██║   ██║   ███████╗
   ╚═════╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝     ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝  ╚═╝   ╚═╝   ╚══════╝

  Welcome to the Confirmate demo!

  This script will start:
    • The Confirmate all-in-one service  (API on http://localhost:${API_PORT})
    • The Confirmate UI                  (http://localhost:${UI_PORT})
    • The code-analysis collector        (Gradle/Kotlin)

  Press Ctrl+C at any time to stop everything.

EOF

# ── 1. Sync submodules ────────────────────────────────────────────────────────
echo "[1/4] Syncing submodules..."
cd "${REPO_ROOT}"
git submodule sync --quiet
git submodule update --init --remote collectors/code-analysis

# ── 2. Build & start the confirmate all-in-one binary ─────────────────────────
echo "[2/4] Building confirmate..."
cd "${REPO_ROOT}"
go build -o bin/confirmate ./core/cmd/confirmate

echo "[2/4] Starting confirmate (API on :${API_PORT})..."
./bin/confirmate \
  --db-in-memory \
  --api-port "${API_PORT}" \
  --catalogs-load-default \
  --metrics-load-default \
  &>"${REPO_ROOT}/logs/confirmate.log" &
PID_CONFIRMATE=$!
echo "      PID ${PID_CONFIRMATE} — logs: logs/confirmate.log"

# ── 3. Start the UI ───────────────────────────────────────────────────────────
echo "[3/4] Starting UI (http://localhost:${UI_PORT})..."
mkdir -p "${REPO_ROOT}/logs"
cd "${REPO_ROOT}/ui-legacy"
pnpm install --silent
pnpm run dev -- --port "${UI_PORT}" &>"${REPO_ROOT}/logs/ui.log" &
PID_UI=$!
echo "      PID ${PID_UI} — logs: logs/ui.log"

# ── 4. Start the code-analysis collector ──────────────────────────────────────
echo "[4/4] Starting code-analysis collector..."
cd "${REPO_ROOT}/collectors/code-analysis"

./gradlew --quiet run &>"${REPO_ROOT}/logs/code-analysis.log" &
PID_COLLECTOR=$!
echo "      PID ${PID_COLLECTOR} — logs: logs/code-analysis.log"

# ── Wait for the UI to be ready, then open the browser ────────────────────────
echo ""
echo "Waiting for services to be ready..."
cd "${REPO_ROOT}"

for i in $(seq 1 30); do
  if curl -sf "http://localhost:${UI_PORT}" >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

echo ""
echo "Opening http://localhost:${UI_PORT} in your browser..."
if command -v open &>/dev/null; then
  open "http://localhost:${UI_PORT}"
elif command -v xdg-open &>/dev/null; then
  xdg-open "http://localhost:${UI_PORT}"
fi

echo ""
echo "All services are running. Press Ctrl+C to stop."
echo ""

# Keep the script alive until interrupted
wait
