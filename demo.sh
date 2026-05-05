#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UI_PORT=5173
API_PORT=8080

# ── Parameters ────────────────────────────────────────────────────────────────
CODE_ANALYSIS_DIR="${REPO_ROOT}/collectors/code-analysis"
GRADLE_PROJECT="example-project"

DAEMON_MODE=false
IMPORT_BSI_C3A=false

usage() {
  echo "Usage: $0 [options]"
  echo ""
  echo "  --code-analysis-dir <path>   Path to the code-analysis repo (default: collectors/code-analysis)"
  echo "  --gradle-project <name>      Gradle sub-project to run (default: example-project)"
  echo "  --import-bsi-c3a             Run BSI C3A catalog importer before starting demo"
  echo "  --daemon                     Run in background (don't open browser, don't wait)"
  echo "  --stop                       Stop a running demo"
  echo "  -h, --help                   Show this help"
  exit 1
}

stop_demo() {
  PID_FILE="${REPO_ROOT}/logs/demo.pids"
  if [[ ! -f "${PID_FILE}" ]]; then
    echo "No running demo found (no PID file: ${PID_FILE})"
    exit 1
  fi

  PIDs=$(cat "${PID_FILE}")
  echo "Stopping demo processes..."
  for pid in ${PIDs}; do
    if kill -0 "${pid}" 2>/dev/null; then
      kill "${pid}" 2>/dev/null && echo "  Killed PID ${pid}" || echo "  Failed to kill PID ${pid}"
    else
      echo "  Process ${pid} not running"
    fi
  done
  rm -f "${PID_FILE}"
  echo "Done."
  exit 0
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --code-analysis-dir)
      CODE_ANALYSIS_DIR="$2"; shift 2 ;;
    --gradle-project)
      GRADLE_PROJECT="$2"; shift 2 ;;
    --import-bsi-c3a)
      IMPORT_BSI_C3A=true; shift ;;
    --daemon)
      DAEMON_MODE=true; shift ;;
    --stop)
      stop_demo ;;
    -h|--help)
      usage ;;
    *)
      echo "Unknown argument: $1"; usage ;;
  esac
done

cleanup() {
  echo ""
  echo "Shutting down demo..."
  kill "${PID_CONFIRMATE:-}" 2>/dev/null || true
  kill "${PID_UI:-}" 2>/dev/null || true
  kill "${PID_COLLECTOR:-}" 2>/dev/null || true
  wait 2>/dev/null || true
  rm -f "${REPO_ROOT}/logs/demo.pids"
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
echo "[1/5] Syncing submodules..."
cd "${REPO_ROOT}"
git submodule sync --quiet
if [[ "${CODE_ANALYSIS_DIR}" == "${REPO_ROOT}/collectors/code-analysis" ]]; then
  git submodule update --init --remote collectors/code-analysis
else
  echo "      Using custom code-analysis dir: ${CODE_ANALYSIS_DIR} (skipping submodule update)"
fi

# ── 2. Optionally import BSI C3A catalog ──────────────────────────────────────
if [[ "${IMPORT_BSI_C3A}" == "true" ]]; then
  echo "[2/5] Importing BSI C3A catalog..."
  cd "${REPO_ROOT}"
  if [[ ! -f "importers/bsi_c3a.py" ]]; then
    echo "      Importer not found, skipping..."
  else
    python3 importers/bsi_c3a.py --output example-data/catalogs/bsi-c3a-catalog.json
    echo "      Imported BSI C3A catalog"
  fi
  STEP_OFFSET=1
else
  STEP_OFFSET=0
fi

# ── $((2 + STEP_OFFSET)). Build & start the confirmate all-in-one binary ───────
echo "[$((2 + STEP_OFFSET))/5] Building confirmate..."
cd "${REPO_ROOT}"
go build -o bin/confirmate ./core/cmd/confirmate

echo "[$((3 + STEP_OFFSET))/5] Starting confirmate (API on :${API_PORT})..."
./bin/confirmate \
  --db-in-memory \
  --api-port "${API_PORT}" \
  --catalogs-default-path ./example-data/catalogs \
  --catalogs-load-default \
  --metrics-default-path ./example-data/metrics \
  --metrics-load-default \
  &>"${REPO_ROOT}/logs/confirmate.log" &
PID_CONFIRMATE=$!
echo "      PID ${PID_CONFIRMATE} — logs: logs/confirmate.log"

# ── $((3 + STEP_OFFSET)). Start the UI ───────────────────────────────────────
echo "[$((4 + STEP_OFFSET))/5] Starting UI (http://localhost:${UI_PORT})..."
mkdir -p "${REPO_ROOT}/logs"
cd "${REPO_ROOT}/ui"
npm install --silent
npm run dev -- --port "${UI_PORT}" &>"${REPO_ROOT}/logs/ui.log" &
PID_UI=$!
echo "      PID ${PID_UI} — logs: logs/ui.log"

# ── $((4 + STEP_OFFSET)). Set up jep virtualenv for Python frontend ─────────
echo "[$((5 + STEP_OFFSET))/5] Setting up Python virtualenv with jep..."
VENV_NAME="confirmate-demo"
VENV_DIR="${HOME}/.virtualenvs/${VENV_NAME}"
# Recreate the venv if the interpreter it was built with is gone
if [[ -d "${VENV_DIR}" ]] && ! "${VENV_DIR}/bin/python3" -c "" 2>/dev/null; then
  rm -rf "${VENV_DIR}"
fi
if [[ ! -d "${VENV_DIR}" ]]; then
  mkdir -p "${HOME}/.virtualenvs"
  python3 -m venv "${VENV_DIR}"
fi
# shellcheck disable=SC1090
source "${VENV_DIR}/bin/activate"
pip3 install --quiet "jep==4.3.1"
deactivate

# Tell CPG's JepSingleton which virtualenv to use
export CPG_PYTHON_VIRTUALENV="${VENV_NAME}"

# ── $((5 + STEP_OFFSET)). Start the code-analysis collector ─────────────────
echo "[$((6 + STEP_OFFSET))/5] Starting code-analysis collector..."
echo "      dir: ${CODE_ANALYSIS_DIR}, project: ${GRADLE_PROJECT}"
cd "${CODE_ANALYSIS_DIR}"

./gradlew --no-daemon ":${GRADLE_PROJECT}:run" &>"${REPO_ROOT}/logs/code-analysis.log" &
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
echo "All services are running."
echo ""

if [[ "${DAEMON_MODE}" == "true" ]]; then
  echo "Running in daemon mode — exiting."
  echo "PID file: ${REPO_ROOT}/logs/demo.pids"
  echo "Logs:"
  echo "  - confirmate: logs/confirmate.log"
  echo "  - UI: logs/ui.log"
  echo "  - code-analysis: logs/code-analysis.log"
  echo ""
  echo "To stop: kill \$(cat logs/demo.pids)"
  echo "${PID_CONFIRMATE}
${PID_UI}
${PID_COLLECTOR}" > "${REPO_ROOT}/logs/demo.pids"
  trap - EXIT
  exit 0
fi

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
