# Observability

Auth, Listings, and Support each expose `GET /metrics` from an isolated
Prometheus registry. The common request metrics are:

- `hacuba_http_requests_total{service,route,method,status}`
- `hacuba_http_request_duration_seconds{service,route,method}`

The only labels are bounded service names, registered HTTP route patterns,
methods, and status codes. Request IDs, account IDs, emails, listing IDs,
search strings, IP addresses, and request bodies are never metrics labels.

The optional local stack in `observability/` scrapes the three development
ports and provisions Grafana at `http://localhost:3001`. In production,
Prometheus must reach each service over Railway private networking; do not
publish `/metrics` through the client-facing service URL.

## Sentry

Set the server-only `SENTRY_DSN` on the Next workload and any Go service that
should report exceptions. Set `NEXT_PUBLIC_SENTRY_DSN` only when browser errors
should go to a browser Sentry project. With either value absent, Sentry remains
disabled and makes no network requests.

Sentry's default PII collection remains disabled. Before events are sent, the
Next integration also removes request cookies, headers, and query strings; Go
service events remove the request entirely. Never put credentials, customer
data, report text, or listing details in exception messages.
