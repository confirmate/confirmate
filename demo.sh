#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UI_PORT=5173
API_PORT=8080
# The host address the demo is reachable at (for OAuth redirects and CORS).
# Defaults to localhost; override with --host <ip> for remote access.
DEMO_HOST="localhost"

# ── Parameters ────────────────────────────────────────────────────────────────
CODE_ANALYSIS_DIR="${REPO_ROOT}/collectors/code-analysis"
GRADLE_PROJECT="example-project"

DAEMON_MODE=false
IMPORT_BSI_C3A=false
DOC_ANALYSER=false

usage() {
  echo "Usage: $0 [options]"
  echo ""
  echo "  --code-analysis-dir <path>   Path to the code-analysis repo (default: collectors/code-analysis)"
  echo "  --gradle-project <name>      Gradle sub-project to run (default: example-project)"
  echo "  --import-bsi-c3a             Run BSI C3A catalog importer before starting demo"
  echo "  --doc-analyser               Run the document analyser on sample PDFs after startup"
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
    --doc-analyser)
      DOC_ANALYSER=true; shift ;;
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
    • The document analyser              (--doc-analyser flag, optional)

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
  --auth-enabled \
  --oauth2-embedded \
  --demo-seed-file ./example-data/demo/seed.json \
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

# ── Seed demo audit scope via REST ────────────────────────────────────────────
# Create the "Demo Audit" scope via the CreateAuditScope RPC so that
# autoCreateControlsInScope and audit trail events run through the regular
# service path (the old Go-level seed func bypassed them).
echo "Seeding demo audit scope..."
API="http://localhost:${API_PORT}"

# Temporarily disable exit-on-error for the seeding section
set +e

# Fetch a service token (service client has ROLE_ADMIN)
TOKEN=""
for i in $(seq 1 15); do
  RESP="$(curl -sf -X POST "${API}/v1/auth/token" \
    -d 'grant_type=client_credentials' \
    -u 'confirmate:confirmate' 2>/dev/null)"
  TOKEN="$(echo "${RESP}" | python3 -c 'import sys,json; print(json.load(sys.stdin).get("access_token",""))' 2>/dev/null)"
  if [[ -n "${TOKEN}" ]]; then
    break
  fi
  sleep 1
done

if [[ -z "${TOKEN}" ]]; then
  echo "WARNING: Could not obtain service token, skipping audit scope seeding"
else
  # Get the default target of evaluation ID
  TOE_ID="$(
    curl -sf "${API}/v1/orchestrator/targets_of_evaluation?pageSize=1" \
      -H "Authorization: Bearer ${TOKEN}" 2>/dev/null \
    | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("targetsOfEvaluation",[{}])[0].get("id",""))' 2>/dev/null
  )"

  if [[ -z "${TOE_ID}" ]]; then
    echo "WARNING: Could not find default target of evaluation, skipping audit scope seeding"
  else
    # Create the "Demo Audit" audit scope via REST — this triggers
    # autoCreateControlsInScope (with audit trail events) and grants the
    # service client ADMIN permission on the scope.
    SCOPE_ID="$(
      curl -sf -X POST "${API}/v1/orchestrator/audit_scopes" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"Demo Audit\",\"targetOfEvaluationId\":\"${TOE_ID}\",\"catalogId\":\"CRA_Catalog\",\"status\":\"AUDIT_SCOPE_STATUS_SETUP\"}" 2>/dev/null \
      | python3 -c 'import sys,json; print(json.load(sys.stdin).get("id",""))' 2>/dev/null
    )"

    if [[ -z "${SCOPE_ID}" ]]; then
      echo "WARNING: Could not create audit scope (may already exist), skipping permission grants"
    else
      echo "  Created audit scope: ${SCOPE_ID}"

      # Grant alice and bob CONTRIBUTOR access to the audit scope.
      USERS_JSON="$(
        curl -sf "${API}/v1/users" \
          -H "Authorization: Bearer ${TOKEN}" 2>/dev/null
      )"

      for USERNAME in alice bob; do
        USER_ID="$(
          echo "${USERS_JSON}" \
          | python3 -c "import sys,json; users=json.load(sys.stdin).get('users',[]); print(next((u['id'] for u in users if u.get('username')=='${USERNAME}'),''))" 2>/dev/null
        )"
        if [[ -n "${USER_ID}" ]]; then
          curl -sf -X PUT \
            "${API}/v1/users/permissions/5/${SCOPE_ID}/users/${USER_ID}/2" \
            -H "Authorization: Bearer ${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{}' >/dev/null 2>/dev/null
          echo "  Granted ${USERNAME} CONTRIBUTOR on audit scope"
        fi
      done
    fi
  fi
fi

# ── Pre-populate implementation view with sample data ─────────────────────────
# Loads implementation states (e.g. IN_PROGRESS, IMPLEMENTED, etc.) and
# implementation details from demo/implementation-seed.json and applies them
# to the ControlInScope records via REST.
SEED_FILE="${REPO_ROOT}/example-data/demo/implementation-seed.json"
if [[ -n "${SCOPE_ID:-}" && -f "${SEED_FILE}" ]]; then
  echo "  Pre-populating implementation data..."

  # Write the token and API URL to temp files for the Python script
  TMP_DIR="$(mktemp -d)"
  echo -n "${TOKEN}" > "${TMP_DIR}/token"
  echo -n "${API}" > "${TMP_DIR}/api_url"
  echo -n "${SCOPE_ID}" > "${TMP_DIR}/scope_id"

  # Save API responses to temp files to avoid shell interpolation issues
  curl -sf "${API}/v1/orchestrator/controls?filter.catalogId=CRA_Catalog&pageSize=1000" \
    -H "Authorization: Bearer ${TOKEN}" 2>/dev/null > "${TMP_DIR}/controls.json" || true
  curl -sf "${API}/v1/orchestrator/controls_in_scope?filter.auditScopeId=${SCOPE_ID}&pageSize=1000" \
    -H "Authorization: Bearer ${TOKEN}" 2>/dev/null > "${TMP_DIR}/cis.json" || true
  curl -sf "${API}/v1/users" \
    -H "Authorization: Bearer ${TOKEN}" 2>/dev/null > "${TMP_DIR}/users.json" || true

  python3 "${REPO_ROOT}/example-data/demo/apply-implementation-seed.py" \
    "${SEED_FILE}" \
    "${TMP_DIR}/controls.json" \
    "${TMP_DIR}/cis.json" \
    "${TMP_DIR}/users.json" \
    "${TMP_DIR}/token" \
    "${TMP_DIR}/api_url" 2>&1 || echo "  WARNING: could not pre-populate implementation data"

  rm -rf "${TMP_DIR}"
fi

# ── Run document analyser on sample PDFs ──────────────────────────────────────
if [[ "${DOC_ANALYSER}" == "true" ]]; then
  echo "Running document analyser on sample PDFs..."
  DOC_DIR="${REPO_ROOT}/collectors/documents"
  PDF_DIR="${DOC_DIR}/vibe-pythonpass-pdfs"

  # Set up a venv for the document analyser if needed
  DOC_VENV="${HOME}/.virtualenvs/document-analyser"
  if [[ ! -d "${DOC_VENV}" ]]; then
    python3 -m venv "${DOC_VENV}"
  fi
  # shellcheck disable=SC1091
  source "${DOC_VENV}/bin/activate"
  pip install --quiet --upgrade pip
  pip install --quiet -r "${DOC_DIR}/requirements.txt"

  # Use the LLM endpoint from environment or default to the Fraunhofer GLM model
  export DOC_ANALYSER_BASE_URL="${DOC_ANALYSER_BASE_URL:-https://dgx-b200-node1.aisec.fraunhofer.de:32008/v1}"
  export DOC_ANALYSER_API_KEY="${DOC_ANALYSER_API_KEY:-not-needed}"
  export DOC_ANALYSER_MAX_TOKENS="${DOC_ANALYSER_MAX_TOKENS:-4096}"
  export INSECURE_TLS=1

  echo "  Analysing PDFs in ${PDF_DIR}..."
  PYTHONPATH="${DOC_DIR}/src" python3 -m document_analyser.cli "${PDF_DIR}" \
    --mode resources \
    --model "${DOC_ANALYSER_MODEL:-glm-5.2-fp8}" \
    --push-evidence \
    --evidence-url "${API}" \
    --evidence-token "${TOKEN}" \
    &>"${REPO_ROOT}/logs/document-analyser.log" && \
    echo "  Document analysis complete (see logs/document-analyser.log)" || \
    echo "  WARNING: document analysis failed (see logs/document-analyser.log)"

  deactivate 2>/dev/null || true
fi

set -e
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
