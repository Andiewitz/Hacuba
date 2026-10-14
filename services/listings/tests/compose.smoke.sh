#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:${PORT:-8081}}"
health="$(curl -fsS "$BASE_URL/healthz")"
grep -q '"service":"listings"' <<<"$health"

public="$(curl -fsS "$BASE_URL/listings")"
grep -q '"listings"' <<<"$public"

status="$(curl -sS -o /tmp/listings-invalid-filter.json -w '%{http_code}' "$BASE_URL/listings?type=not-a-property")"
[ "$status" = "400" ] || { cat /tmp/listings-invalid-filter.json; exit 1; }

status="$(curl -sS -o /tmp/listings-write.json -w '%{http_code}' -X POST "$BASE_URL/listings" -H 'Content-Type: application/json' --data '{"listing_mode":"for_sale"}')"
[ "$status" = "401" ] || { cat /tmp/listings-write.json; exit 1; }
echo "listings compose smoke passed"
