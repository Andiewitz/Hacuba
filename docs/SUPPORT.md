# Support service

`services/support` accepts marketplace, safety, bug, and account reports. A
report description, email, listing reference, and optional authenticated user
ID are private Support database data. They are never Prometheus labels or log
fields.

## Metrics

The service exposes `GET /metrics` for an internal Prometheus scraper:

- `hacuba_support_reports_total{category}`
- `hacuba_support_reports_open`
- `hacuba_support_report_rejections_total{reason}`

Labels are bounded categories and rejection reasons only. Do not add report
IDs, listing references, emails, user IDs, descriptions, or IPs as labels.

## Local development

Run Support with its explicit SQLite development store on port 8082. The
client forwards reports through `/api/support/reports`; authenticated reports
forward the existing Bearer token so Support can attach the reporter UUID.
Set the same server-only `SUPPORT_PROXY_SECRET` in Next and Support so the
service can rate-limit by the forwarded client address instead of treating all
proxied users as one client.

To view aggregate metrics, start the optional observability stack from
`observability/` and open Grafana at `http://localhost:3001` (local user:
`hacuba`; password: `hacuba-local-only`). It scrapes the local Support service
at port 8082. The dashboard intentionally cannot display report contents.

Production uses a separate private Support PostgreSQL database and an
internal Prometheus scrape route. Configure the database DSN and `JWT_SECRET`
through the workload's secret manager; do not use the SQLite path.
