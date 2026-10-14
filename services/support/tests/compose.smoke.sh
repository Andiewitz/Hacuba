#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:${PORT:-8082}}"

health="$(curl -fsS "$BASE_URL/healthz")"
grep -q '"service":"support"' <<<"$health"

report="$(curl -fsS -X POST "$BASE_URL/reports" \
  -H 'Content-Type: application/json' \
  --data '{"category":"listing","reason":"misleading_information","listing_reference":"https://hacuba.test/listings/smoke","description":"The smoke suite verifies that a complete support report is accepted."}')"
grep -q '"status":"open"' <<<"$report"

status="$(curl -sS -o /tmp/support-staff.json -w '%{http_code}' "$BASE_URL/staff/reports")"
[ "$status" = "403" ] || { cat /tmp/support-staff.json; exit 1; }

metrics="$(curl -fsS "$BASE_URL/metrics")"
grep -q 'hacuba_support_reports_accepted_total' <<<"$metrics"
echo "support compose smoke passed"
