# Hacuba

Hacuba is a marketplace for browsing real-estate listings, contacting sellers, creating and publishing listings, and reporting safety or product issues. The public Next.js application is the only internet-facing application service; it proxies requests to private Auth, Listings, and Support APIs.

## Stack

- Next.js 16, React 19, TypeScript, Tailwind CSS, Redux Toolkit
- Go services for authentication, listings/discovery, and support triage
- PostgreSQL in production; explicit SQLite files for local development only
- S3 and CloudFront for production listing images
- Prometheus, Grafana, structured logs, and optional Sentry error reporting
- Terraform for AWS image infrastructure and GitHub Actions for verification

## Architecture

```text
Browser
  │ HTTPS
  ▼
Next.js client (public)
  ├── private ──► Auth API ─────► auth PostgreSQL
  ├── private ──► Listings API ─► listings PostgreSQL + S3/CloudFront
  └── private ──► Support API ──► support PostgreSQL
```

The APIs are private Railway services. This keeps authentication cookies on one public origin, avoids public API CORS, and limits each service to its own database credentials.

## Run locally without Docker

Requirements: Node.js 22+, Go 1.26+, and PowerShell. Copy `client/.env.example` to ignored `client/.env.local`, leaving `APP_ENV=development` and the local service URLs in place.

Install client packages once:

```powershell
npm ci
```

Start each API in a separate PowerShell terminal. Use the same development JWT secret in all three commands.

```powershell
# services/auth
$env:DEV_SQLITE_PATH = ".dev/auth.db"
$env:JWT_SECRET = "development-secret-at-least-32-characters"
$env:COOKIE_SECURE = "false"
go run ./cmd/server
```

```powershell
# services/listings
$env:DEV_SQLITE_PATH = ".dev/listings.db"
$env:JWT_SECRET = "development-secret-at-least-32-characters"
$env:DEV_UPLOAD_URL = "http://localhost:3000/api/dev-uploads"
go run ./cmd/server
```

```powershell
# services/support
$env:DEV_SQLITE_PATH = ".dev/support.db"
$env:JWT_SECRET = "development-secret-at-least-32-characters"
go run ./cmd/server
```

Then start the client:

```powershell
npm run dev --workspace client
```

Open [http://localhost:3000](http://localhost:3000). SQLite paths and local uploads are rejected when `APP_ENV=production`.

## Verification

Run the client checks:

```powershell
npm run typecheck --workspace client
npm run lint --workspace client
npm run build --workspace client
npm audit --omit=dev --audit-level=high
```

Run a service suite from its directory:

```powershell
go test ./...
go vet ./...
```

The Compose-backed Playwright suite covers registration, seller activation, listing creation with image upload and publishing, filtered browsing, listing detail/gallery, report submission, staff moderation, validation failures, protected routes, 404s, session restoration, and logout:

```powershell
docker compose -f docker-compose.e2e.yml up -d --build
npm run test:e2e --workspace client
docker compose -f docker-compose.e2e.yml down -v --remove-orphans
```

GitHub Actions runs all service suites, the client production build and audit, Terraform validation, and the browser journey through the required `verify` workflow.

## Production deployment

Railway deploys four application services: public `client`, plus private `auth`, `listings`, and `support`. Each backend owns a private PostgreSQL database. The backend services run their versioned migrations before a new deployment serves traffic and expose `/readyz` for Railway health checks.

See [the Railway configuration guide](deploy/railway/README.md) for required environment variables, service setup, protected production deployment, and rollback. See [the infrastructure guide](terraform/README.md) for production image storage.

## Operations

- `GET /healthz` confirms each API process is alive.
- `GET /readyz` confirms that API's own database is reachable.
- `GET /api/health` confirms the public Next server is live.
- Prometheus metrics are private. The optional local Grafana stack is in `observability/` and is documented in [Observability](docs/OBSERVABILITY.md).
- Sentry is opt-in: set server-only `SENTRY_DSN` and, for browser errors, `NEXT_PUBLIC_SENTRY_DSN`.

Use the [operations runbook](docs/OPERATIONS.md) for incident diagnosis and Railway rollback.

## Design system

The design tokens, typography, spacing, and accessible color pairings live in [DESIGN.md](DESIGN.md). The interface uses solid white surfaces for primary search and content controls, with the forest/terracotta palette as accents.

## License

[MIT](LICENSE)
