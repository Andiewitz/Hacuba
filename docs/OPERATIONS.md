# Operations runbook

## Before production deployment

1. Create a protected GitHub environment named `production`.
2. Add a Railway project token as that environment's `RAILWAY_TOKEN` secret.
3. Add `RAILWAY_PROJECT_ID`, `RAILWAY_ENVIRONMENT`, and optionally `RAILWAY_PUBLIC_URL` as repository variables.
4. Configure branch protection so a merge to `main` requires the `verify` workflow.
5. Configure four Railway application services with the TOML files in `deploy/railway/` and the private environment variables documented there.

`verify` runs client quality and build checks, all Go test/smoke suites, Terraform validation, and browser journeys. The Railway workflow only starts for a successful `main` SHA that is still the current `main` head.

## Diagnose an incident

Start with the affected service's Railway deployment and structured logs. Requests include a `request_id`, service name, method, route, response status, and latency. Do not paste credentials, cookies, report contents, or user data into incident notes.

| Symptom                                   | Check                                    | Next action                                                                           |
| ----------------------------------------- | ---------------------------------------- | ------------------------------------------------------------------------------------- |
| Client is unavailable                     | `GET /api/health`                        | Inspect Client deployment logs and verify its three private API URLs.                 |
| Auth, Listings, or Support is unavailable | `GET /readyz` through private networking | Inspect the owning Postgres service, its migration logs, and database URL.            |
| Listings upload fails                     | Listings logs and S3/IAM configuration   | Verify the bucket, region, and task role; do not replace production S3 with a volume. |
| User report cannot be submitted           | Support logs and rate-limit response     | Check the proxy secret and `Retry-After`; preserve private report data.               |
| Unexpected exception                      | Sentry issue and request ID in logs      | Confirm event scrubbers removed request PII, then fix and deploy normally.            |

`/healthz` only proves a process is running. Use `/readyz` to determine whether that service can reach its database.

## Roll back safely

1. Identify the most recent successful deployment for the affected Railway service.
2. Redeploy that deployment from Railway's deployment history.
3. Confirm the service's `/readyz` check is successful and exercise the affected public flow through the client.
4. Record the incident, inspect logs and Sentry, fix forward on a branch, and let `verify` gate the replacement deployment.

Backend migrations are forward-only. Do not manually reverse schema changes in a rollback. If a migration causes the incident, restore service traffic to the last compatible application version and prepare a new, forward corrective migration.
