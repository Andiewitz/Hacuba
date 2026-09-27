FROM golang:1.26-alpine AS build
WORKDIR /app
COPY services/support/go.mod services/support/go.sum ./services/support/
COPY shared/authjwt ./shared/authjwt
WORKDIR /app/services/support
RUN go mod download
COPY services/support .
RUN CGO_ENABLED=0 go build -o /support ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 support
USER support
COPY --from=build /support /support
EXPOSE 8082
ENTRYPOINT ["/support"]
