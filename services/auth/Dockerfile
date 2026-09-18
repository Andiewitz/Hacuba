FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /auth ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 auth
USER auth
COPY --from=build /auth /auth
EXPOSE 8080
ENTRYPOINT ["/auth"]
