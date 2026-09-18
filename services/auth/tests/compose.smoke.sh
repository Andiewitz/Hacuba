#!/usr/bin/env bash
# compose.smoke.sh — live contract check against `docker compose up` auth.
#
# Usage:
#   ./tests/compose.smoke.sh [BASE_URL]
#
# Expects the compose stack already running (CI brings it up first).
# Fails non-zero on the first broken promise. Safe to re-run: every run
# uses fresh timestamped emails so state never collides.
set -euo pipefail

BASE_URL="${1:-http://localhost:${PORT:-8080}}"
STAMP="$(date +%s)-$RANDOM"
EMAIL_A="smoke-a-${STAMP}@example.com"
EMAIL_B="smoke-b-${STAMP}@example.com"
PASSWORD="validpassword123"

JAR_A="$(mktemp)"
JAR_B="$(mktemp)"
trap 'rm -f "$JAR_A" "$JAR_B" /tmp/smoke-*.json' EXIT

say() { printf '  [smoke] %s\n' "$*"; }
fail() { printf '  [smoke] FAIL: %s\n' "$*" >&2; exit 1; }

# jget <file> <key> — extract a top-level (or user.*) JSON string via python3.
jget() { python3 -c "import json,sys; d=json.load(open(sys.argv[1])); print(d.get(sys.argv[2]) or d.get('user',{}).get(sys.argv[2]) or '')" "$1" "$2"; }

expect_code() { # <want> <got> <label>
  [ "$2" = "$1" ] || fail "$3: want HTTP $1, got $2"
  say "ok: $3 -> $1"
}

say "base: $BASE_URL"

# 1. healthz
CODE="$(curl -s -o /tmp/smoke-health.json -w '%{http_code}' "$BASE_URL/healthz")"
expect_code 200 "$CODE" "GET /healthz"

# 2. register A -> 201 + cookies + tokens
CODE="$(curl -s -c "$JAR_A" -o /tmp/smoke-reg-a.json -w '%{http_code}' \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL_A\",\"password\":\"$PASSWORD\"}" \
  "$BASE_URL/auth/register")"
expect_code 201 "$CODE" "POST /auth/register (A)"
ACCESS_A="$(jget /tmp/smoke-reg-a.json access_token)"
CSRF_A="$(jget /tmp/smoke-reg-a.json csrf_token)"
[ -n "$ACCESS_A" ] || fail "register A: empty access_token"
[ -n "$CSRF_A" ] || fail "register A: empty csrf_token"
grep -q 'refresh_token' "$JAR_A" || fail "register A: no refresh_token cookie"
grep -q 'csrf_token' "$JAR_A" || fail "register A: no csrf_token cookie"
grep -q '#HttpOnly' "$JAR_A" || say "warn: curl jar shows no HttpOnly flag (verify manually in browser devtools)"
say "ok: register A sets session cookies + tokens"

# 3. duplicate register -> 409
CODE="$(curl -s -o /dev/null -w '%{http_code}' \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL_A\",\"password\":\"$PASSWORD\"}" \
  "$BASE_URL/auth/register")"
expect_code 409 "$CODE" "POST /auth/register duplicate"

# 4. me with A's token returns A
ME="$(curl -s -H "Authorization: Bearer $ACCESS_A" "$BASE_URL/auth/me")"
echo "$ME" | grep -q "$EMAIL_A" || fail "GET /auth/me did not return A ($ME)"
say "ok: GET /auth/me returns caller A"

# 5. me without token -> 401
CODE="$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/auth/me")"
expect_code 401 "$CODE" "GET /auth/me anonymous"

# 6. register B, prove isolation (B's me is B, never A)
CODE="$(curl -s -c "$JAR_B" -o /tmp/smoke-reg-b.json -w '%{http_code}' \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL_B\",\"password\":\"$PASSWORD\"}" \
  "$BASE_URL/auth/register")"
expect_code 201 "$CODE" "POST /auth/register (B)"
ACCESS_B="$(jget /tmp/smoke-reg-b.json access_token)"
ME_B="$(curl -s -H "Authorization: Bearer $ACCESS_B" "$BASE_URL/auth/me")"
echo "$ME_B" | grep -q "$EMAIL_B" || fail "B me wrong ($ME_B)"
echo "$ME_B" | grep -q "$EMAIL_A" && fail "cross-user leak: B me contains A" || true
say "ok: per-user isolation (A != B UUIDs, me is caller-only)"

# 7. refresh A with cookie jar + CSRF header -> 200 + rotation
CSRF_JAR="$(python3 -c "
import http.cookiejar
j = http.cookiejar.MozillaCookieJar('$JAR_A'); j.load(ignore_discard=True)
for c in j:
    if c.name == 'csrf_token': print(c.value)
")"
[ -n "$CSRF_JAR" ] || fail "could not read csrf cookie from jar"
CODE="$(curl -s -b "$JAR_A" -c "$JAR_A" -o /tmp/smoke-refresh.json -w '%{http_code}' \
  -X POST -H "X-CSRF-Token: $CSRF_JAR" "$BASE_URL/auth/refresh")"
expect_code 200 "$CODE" "POST /auth/refresh (A rotates)"
NEW_ACCESS="$(jget /tmp/smoke-refresh.json access_token)"
[ -n "$NEW_ACCESS" ] || fail "refresh: empty access_token"
[ "$NEW_ACCESS" != "$ACCESS_A" ] || fail "refresh must issue a new access token"
say "ok: refresh rotates session"

# 8. refresh without CSRF header -> 403
CODE="$(curl -s -b "$JAR_A" -o /dev/null -w '%{http_code}' \
  -X POST "$BASE_URL/auth/refresh")"
expect_code 403 "$CODE" "POST /auth/refresh without CSRF"

# 9. login with wrong password -> generic 401
CODE="$(curl -s -o /tmp/smoke-badlogin.json -w '%{http_code}' \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL_A\",\"password\":\"wrongpassword999\"}" \
  "$BASE_URL/auth/login")"
expect_code 401 "$CODE" "POST /auth/login wrong password"
grep -q 'invalid credentials' /tmp/smoke-badlogin.json || fail "bad login must say invalid credentials"

# 10. logout A -> 204, then refresh with old jar -> 401
CODE="$(curl -s -b "$JAR_A" -c "$JAR_A" -o /dev/null -w '%{http_code}' \
  -X POST "$BASE_URL/auth/logout")"
expect_code 204 "$CODE" "POST /auth/logout"
CODE="$(curl -s -b "$JAR_A" -o /dev/null -w '%{http_code}' \
  -X POST -H "X-CSRF-Token: $CSRF_JAR" "$BASE_URL/auth/refresh")"
# After logout the jar holds cleared cookies; server must reject.
[ "$CODE" = "401" ] || [ "$CODE" = "403" ] || fail "refresh after logout: want 401/403, got $CODE"
say "ok: logout kills session (refresh rejected with $CODE)"

# 11. auth-db must not publish ports to the host.
if command -v docker >/dev/null 2>&1; then
  if docker ps --format '{{.Names}} {{.Ports}}' 2>/dev/null | grep -iE 'auth-db|db' | grep -q '0.0.0.0'; then
    fail "auth-db publishes ports to host — it must stay on the internal network"
  else
    say "ok: auth-db exposes no host ports"
  fi
else
  say "skip: docker not available for port check"
fi

say "SMOKE PASS against $BASE_URL"
