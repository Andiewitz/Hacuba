#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:${PORT:-8081}}"
code="$(curl -s -o /tmp/listings-health.json -w '%{http_code}' "$BASE_URL/healthz")"
[ "$code" = "200" ] || { cat /tmp/listings-health.json; exit 1; }
grep -q '"service":"listings"' /tmp/listings-health.json
echo "listings compose smoke passed"
