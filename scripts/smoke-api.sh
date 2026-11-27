#!/usr/bin/env bash
# Smoke-test the MoCA Worker against the routes it actually exposes.
# Usage: BASE_URL=https://... [PATIENT_PHONE=... PATIENT_PASSWORD=...] ./scripts/smoke-api.sh

set -euo pipefail

BASE="${BASE_URL:-http://localhost:8787}"
PASS=0
FAIL=0

check() {
  local name="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    echo "OK  $name -> $actual"
    PASS=$((PASS + 1))
  else
    echo "FAIL $name -> $actual (expected $expected)"
    FAIL=$((FAIL + 1))
  fi
}

status() {
  curl -sS -o /dev/null -w "%{http_code}" "$@"
}

post_status() {
  local path="$1" body="$2"
  status -X POST "$BASE$path" -H "Content-Type: application/json" -d "$body"
}

echo "=== MoCA API smoke test @ $BASE ==="

check "GET /api/v1/health" 200 "$(status "$BASE/api/v1/health")"
check "GET /api/v1/doctors/list" 200 "$(status "$BASE/api/v1/doctors/list")"
check "GET /api/v1/doctor/list?id=1" 200 "$(status "$BASE/api/v1/doctor/list?id=1")"
check "GET /error (blocked)" 404 "$(status "$BASE/error")"
check "GET /internal/d1/query (blocked)" 404 "$(status "$BASE/internal/d1/query")"
check "POST /api/v1/auth/patient/login (no password)" 400 \
  "$(post_status /api/v1/auth/patient/login '{"phone_number":"0123456789"}')"
check "POST /api/v1/auth/patient/status (bad phone)" 400 \
  "$(post_status /api/v1/auth/patient/status '{"phone_number":"123"}')"
check "POST /api/v1/auth/doctor/login (bad JSON)" 400 \
  "$(post_status /api/v1/auth/doctor/login '{')"
check "POST /api/v1/auth/forgot-password (bad email)" 400 \
  "$(post_status /api/v1/auth/forgot-password '{"email":"not-an-email"}')"

if [ -n "${PATIENT_PHONE:-}" ] && [ -n "${PATIENT_PASSWORD:-}" ]; then
  body=$(PATIENT_PHONE="$PATIENT_PHONE" PATIENT_PASSWORD="$PATIENT_PASSWORD" python3 -c \
    'import json, os; print(json.dumps({"phone_number": os.environ["PATIENT_PHONE"], "password": os.environ["PATIENT_PASSWORD"]}))')
  check "POST /api/v1/auth/patient/login" 200 \
    "$(post_status /api/v1/auth/patient/login "$body")"
  status_body=$(PATIENT_PHONE="$PATIENT_PHONE" python3 -c \
    'import json, os; print(json.dumps({"phone_number": os.environ["PATIENT_PHONE"]}))')
  check "POST /api/v1/auth/patient/status" 200 \
    "$(post_status /api/v1/auth/patient/status "$status_body")"
else
  echo "SKIP authenticated patient routes (set PATIENT_PHONE and PATIENT_PASSWORD)"
fi

echo "=== $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ]
