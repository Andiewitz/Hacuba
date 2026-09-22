FROM golang:1.26-alpine AS build
WORKDIR /app
COPY shared/authjwt ./shared/authjwt
COPY services/listings ./services/listings
WORKDIR /app/services/listings
RUN go mod download
RUN CGO_ENABLED=0 go build -o /listings ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 listings
USER listings
COPY --from=build /listings /listings
EXPOSE 8081
ENTRYPOINT ["/listings"]
