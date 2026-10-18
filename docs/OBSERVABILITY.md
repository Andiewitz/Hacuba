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
