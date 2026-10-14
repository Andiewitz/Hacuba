# Testing and release gate

Hacuba uses fast tests to identify local regressions and a separate browser
gate to prove that the marketplace works as one application. The browser gate
runs Auth, Listings, Support, their isolated Postgres databases, and the
production-mode Next server together.

## Test tiers

| Tier | Runs | Purpose |
| --- | --- | --- |
| Client checks | `npm run lint`, `npm run lint:ox`, `npm run typecheck`, `npm run build` | Type, lint, and production-build regressions. |
| Service unit and Postgres integration | service GitHub workflows | Domain rules, authorization, migrations, and persistence. |
| Service Compose smoke | Auth, Listings, and Support workflows | Health, public/unauthorized HTTP contracts, cookie/CSRF behavior, and report persistence. |
| Browser E2E | `end-to-end` workflow | Real browser journeys across all application services. |

## Browser contract

The E2E suite performs these actions without mocked browser storage or mocked
service responses:

1. Register a user, activate seller access, create a complete listing, upload
   two images, publish it, and restore the session after a browser reload.
2. Open the public property page in a separate browser context, navigate its
   gallery, see direct seller contact information, and submit a linked report.
3. Register the configured staff account, open the private support queue, hide
   the reported listing with a reason, verify the public page returns `404`,
   restore it, and verify it becomes public again.
4. Log the seller out and prove that the seller workspace is protected.

The CI stack uses disposable Postgres containers. Image uploads use the
explicit development object store only inside that disposable E2E Compose
environment; production keeps using S3 configuration and never receives
`DEV_UPLOAD_PATH`.

## Run locally

```bash
docker compose -f docker-compose.e2e.yml up -d --build
npx playwright install chromium
npm run test:e2e --workspace client
docker compose -f docker-compose.e2e.yml down -v --remove-orphans
```

If a browser test fails, inspect `client/e2e/playwright-report`,
`client/e2e/test-results`, and `docker compose -f docker-compose.e2e.yml logs`.

## CI/CD gate

The `end-to-end / browser journeys (full stack)` check is the deployment gate.
A future deployment workflow must require it before deployment:

```yaml
deploy:
  needs: [browser]
```

Repository branch protection must mark the check as required before releases
are allowed to merge. This repository does not currently contain a deployment
workflow or branch-protection configuration, so the workflow supplies the gate
but cannot enforce GitHub repository settings by itself.
