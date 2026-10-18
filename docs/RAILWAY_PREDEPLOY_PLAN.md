# Railway pre-deploy plan

This is a production-readiness plan for Hacuba's existing architecture: a
public Next.js client that proxies to private Auth, Listings, and Support Go
services. Each service owns its own PostgreSQL database; Listings uses private
S3 objects and CloudFront for public images. Railway hosts the application
workloads, not development SQLite databases or local image uploads.

The plan deliberately keeps browser traffic on one public origin. Auth,
Listings, and Support stay private Railway services and are reached by Next
route handlers over Railway private networking. That preserves HttpOnly
session-cookie behavior and means application APIs do not need public CORS.

## Baseline already in place

- Playwright runs a full-stack browser journey in GitHub Actions. It covers
  registration, seller activation, draft saving, two-image upload, publishing,
  public detail/gallery, reporting, staff moderation, session restoration, and
  logout.
- Each Go service exposes `GET /healthz`; auth has graceful shutdown; auth and
  Support already rate-limit their sensitive public writes.
- Service CI runs Go unit/integration checks and Compose smoke contracts.
- Client API routes proxy to the private services. Listings production storage
  is already S3 + CloudFront; SQLite and local uploads are explicit development
  settings only.

## Deployment topology

Create these Railway services from this monorepo:

| Railway service | Source | Public | Required healthcheck |
| --- | --- | --- | --- |
| `client` | `client/Dockerfile`, repository root build context | Yes | `/api/health` |
| `auth` | `services/auth/Dockerfile` | No | `/readyz` |
| `listings` | `services/listings/Dockerfile`, repository root build context | No | `/readyz` |
| `support` | `services/support/Dockerfile`, repository root build context | No | `/readyz` |
| `auth-db`, `listings-db`, `support-db` | Railway PostgreSQL | No | managed by Railway |

Use three managed Postgres services, one for each owner service. Do not share
database credentials between Auth, Listings, and Support. Keep S3 and
CloudFront as the image layer already defined by Terraform; do not use a
Railway volume for production listing images.

Railway's deployment healthcheck must receive a `2xx` response before it
switches traffic, and it uses the service `PORT` variable. Configure that in
the dashboard after the readiness endpoints land. See Railway's
[healthcheck documentation](https://docs.railway.com/deployments/healthchecks).

## Commit sequence

Each commit is independently reviewable and must pass its named tests before
the next one starts.

### 1. `test(e2e): cover public browse failures and access boundaries`

Extend the existing Playwright suite with focused, independent specs:

- Search/filter a published listing by city, property type, mode, and price.
- Attempt to publish an incomplete draft and assert the server validation
  messages remain visible and the record stays a draft.
- Visit a random listing UUID and assert the client returns the real 404 page.
- Verify anonymous users cannot save a seller listing or reach the staff queue;
  verify a non-staff user gets no staff UI or report data.

Use the real Compose services and Postgres fixtures. Do not mock fetch,
browser storage, or authentication state.

**Checks:** `npm run test:e2e --workspace client` against
`docker-compose.e2e.yml`.

### 2. `feat(runtime): add readiness and production configuration checks`

Make startup and readiness honest:

- Add `GET /readyz` to each Go service. It must confirm that its own database
  is reachable; `/healthz` remains a cheap process liveness response.
- Add a client `GET /api/health` route that confirms the Next server is live
  without returning service URLs, credentials, or upstream error bodies.
- Require production configuration at startup: database URL, JWT secret,
  service-to-service proxy secrets, and Listings' S3 settings. Reject
  `DEV_SQLITE_PATH` and `DEV_UPLOAD_PATH` in production.
- Remove client-side production fallbacks to `localhost`; a missing private
  service URL must fail the deployment clearly instead of serving a broken
  application.

**Checks:** focused Go router/config tests, `go vet ./...`, service Compose
smokes, client typecheck/lint/build, and an E2E readiness wait using `/readyz`.

### 3. `feat(deploy): make database migrations and Railway builds repeatable`

Add a migration command for each database owner. It must record applied
migration filenames in a schema-migrations table, take a PostgreSQL advisory
lock, run files in lexical order, and stop on the first error. It must be safe
to run repeatedly and never run automatically inside an application request.

Add deployment configuration and documentation for the four application
services:

- Auth, Listings, and Support use their existing Dockerfiles and their private
  database reference variables.
- Client uses the root-context Dockerfile so workspace dependencies build
  reliably.
- Railway pre-deploy commands run the relevant migration binary before the
  service can become active.
- The public client receives `AUTH_API_URL`, `LISTINGS_API_URL`, and
  `SUPPORT_API_URL` only as Railway private-network URLs; it receives the
  CloudFront image domain as `LISTINGS_IMAGE_BASE_URL`.
- `COOKIE_SECURE=true`; leave `COOKIE_DOMAIN` unset unless a verified custom
  client domain requires it.

**Checks:** empty-database and already-migrated-database integration tests;
build every production image; run a Railway staging deployment with all
healthchecks configured.

### 4. `feat(reliability): handle shutdown, upstream failure, and abuse safely`

- Give Listings and Support the same `SIGTERM`/`SIGINT` drain behavior as Auth;
  make the drain timeout configurable and ensure new requests stop while
  in-flight requests finish.
- Add bounded database, S3, and outgoing HTTP timeouts. Return generic `5xx`
  API messages while structured logs retain the cause; never leak DSNs, stack
  traces, bearer tokens, cookies, passwords, or reporter details.
- Finish deliberate rate limits for public report, auth, discovery-event, and
  upload/presign endpoints. Limits must identify the endpoint and retry policy
  without punishing all users behind the client proxy.
- Add failure tests for disconnected Postgres, unavailable image storage,
  malformed/absent production variables, and graceful process termination.

**Checks:** Go integration tests for each failure contract and a deployment
smoke that sends `SIGTERM` during an active request.

### 5. `feat(observability): add structured logs, Sentry, and service metrics`

- Standardize Go services on `log/slog` JSON in production with request ID,
  service name, route, status, latency, and error class. Use a redaction test
  suite for sensitive headers and request fields.
- Add Sentry to Next.js and each Go service behind optional `SENTRY_DSN`
  configuration. Enable error capture only; disable default PII collection and
  scrub authorization/cookie headers and report descriptions before sending
  events.
- Extend Prometheus metrics consistently for requests, latency, dependency
  failures, rate limits, image-upload outcomes, and report intake. Keep
  `/metrics` private. Configure the existing Prometheus dashboard/scrape job
  against private service networking rather than exposing metrics publicly.

**Checks:** logger redaction tests, Sentry transport tests, metric assertions,
and a staging exception that appears once in Sentry with no sensitive fields.

### 6. `ci: gate Railway deployments on verified artifacts`

Keep the current GitHub Actions test workflows and make them the only path to
production:

- Protect `main`; require client checks, all service checks, Terraform
  validation, and the full-stack browser job before merge.
- Disable direct Railway GitHub autodeploys for production. Add a deployment
  workflow that runs only after required CI succeeds for a `main` commit, uses
  a repository secret scoped to the Railway production project, and deploys
  the exact commit to each service.
- Use a protected Railway/GitHub production environment so the deployment
  secret is unavailable to pull requests.
- Have the workflow wait for deployment healthchecks and publish the deployed
  commit and public URL in its summary.

**Checks:** staging deployment from a test commit, failed-healthcheck rollback
exercise, and proof that a failing E2E job cannot invoke the Railway deploy.

### 7. `docs: publish operational runbook and portfolio deployment guide`

Rewrite the README around what Hacuba actually ships, not planned features:

- One-paragraph product description, accurate stack, local setup for all four
  services, test commands, and the verified public URL once it exists.
- Link the browser test contract and state exactly which production services
  are private.
- Document environment-variable names and owners, never values.
- Add an incident runbook: inspect Railway/Sentry logs, verify `/readyz`, roll
  back the affected service, then fix forward. Railway rollback restores a
  previous deployment image and its variables while retained; see its
  [deployment actions documentation](https://docs.railway.com/deployments/deployment-actions).
- Add the manual setup for an external uptime monitor: one HTTPS check against
  the public client health endpoint and alert recipients. Do not claim an
  uptime monitor, a live URL, or Sentry is configured until each is verified.

**Checks:** new-contributor setup walkthrough and a staging rollback drill.

## Release criteria

Production deployment can begin only when commits 1–6 are green in a staging
Railway environment, migrations have been tested from both empty and upgraded
schemas, healthchecks are configured for every application service, Sentry has
received a scrubbed test exception, and an external uptime monitor has
confirmed the public client health endpoint. Commit 7 records the real URL
and operational facts after that verification.
