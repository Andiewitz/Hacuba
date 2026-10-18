FROM golang:1.26-alpine AS build
WORKDIR /app
COPY services/auth/go.mod services/auth/go.sum ./services/auth/
COPY shared/authjwt ./shared/authjwt
WORKDIR /app/services/auth
RUN go mod download
COPY services/auth .
RUN CGO_ENABLED=0 go build -o /auth ./cmd/server
RUN CGO_ENABLED=0 go build -o /migrate ./cmd/migrate

FROM alpine:3.20
RUN adduser -D -u 10001 auth
USER auth
COPY --from=build /auth /auth
COPY --from=build /migrate /migrate
COPY services/auth/migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/auth"]
