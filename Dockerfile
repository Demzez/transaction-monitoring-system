
# syntax=docker/dockerfile:1

FROM golang:1.27.4-alpine3.24 AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" \
    -o /out/transaction-monitoring-system \
    ./cmd/transaction-monitoring-system

FROM alpine:3.24

RUN apk add --no-cache ca-certificates netcat-openbsd \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/transaction-monitoring-system /app/server

COPY deploy/config.yaml /app/config.yaml

ENV CONFIG_PATH=/app/config.yaml

USER app

EXPOSE 8082 8084

ENTRYPOINT ["/app/server"]
