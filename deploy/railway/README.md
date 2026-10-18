# Railway service configuration

Railway deploys one service from each configuration file in this directory.
Select the matching absolute path (for example,
`/deploy/railway/auth.toml`) as Railway's **Config File** setting. Set the
repository root for all services because their Dockerfiles copy workspace
dependencies.
These TOML files use Railway's supported `DOCKERFILE` builder, healthchecks,
and pre-deploy commands.

| Railway service | Config file | Network | Required variables |
| --- | --- | --- | --- |
| `auth` | `/deploy/railway/auth.toml` | private | `APP_ENV=production`, `DATABASE_URL` from `auth-db`, `JWT_SECRET`, `COOKIE_SECURE=true` |
| `listings` | `/deploy/railway/listings.toml` | private | `APP_ENV=production`, `DATABASE_URL` from `listings-db`, `JWT_SECRET`, `S3_REGION`, `S3_BUCKET`, `DISCOVERY_PROXY_SECRET` |
| `support` | `/deploy/railway/support.toml` | private | `APP_ENV=production`, `DATABASE_URL` from `support-db`, `JWT_SECRET`, `SUPPORT_PROXY_SECRET` |
| `client` | `/deploy/railway/client.toml` | public | `APP_ENV=production`, private `AUTH_API_URL`, `LISTINGS_API_URL`, `SUPPORT_API_URL`, `LISTINGS_IMAGE_BASE_URL`, `SUPPORT_PROXY_SECRET`, `DISCOVERY_PROXY_SECRET` |

Create one Railway Postgres service for each database owner. Reference the
matching private Postgres URL only from that owner; never share a database URL
between services. `AUTH_API_URL`, `LISTINGS_API_URL`, and `SUPPORT_API_URL`
must be Railway private-network URLs, never public domains or localhost.

Each backend image contains `/migrate` and the immutable SQL files at
`/migrations`. Railway runs `/migrate -dir /migrations` before the application
process starts. The runner holds a Postgres advisory lock, records filenames in
`schema_migrations`, runs files lexically, and is safe to rerun. A migration
failure exits non-zero, so Railway does not activate the new deployment.

Set a five-minute pre-deploy timeout in Railway for the three backend services.
The runner does not need an application process or persistent volume. Railway
pre-deploy commands run with service variables and private-network access, and
their filesystem changes are not retained.

## CI deployment gate

GitHub Actions runs `.github/workflows/verify.yml` for every pull request and
every push to `main`. It calls the client quality/build checks, all Go service
tests and smoke contracts, Terraform validation, and the full Compose-backed
browser journey. Only a successful `main` run can trigger
`.github/workflows/deploy-railway.yml`.

Before enabling deployments, create a protected GitHub environment named
`production`. Add its `RAILWAY_TOKEN` secret as a Railway project token scoped
to the production environment. Add repository variables `RAILWAY_PROJECT_ID`,
`RAILWAY_ENVIRONMENT` (normally `production`), and the optional
`RAILWAY_PUBLIC_URL`. Configure GitHub branch protection to require `verify`
before merging to `main`.

The deployment workflow uploads the verified SHA in order: Auth, Listings,
Support, then Client. Railway waits for each configured readiness healthcheck;
if any deployment fails, the workflow stops before the next service. Roll back
from Railway's deployment history by redeploying the preceding successful
deployment, then investigate service logs and `/readyz`.
