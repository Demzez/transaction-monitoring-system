FROM golang:1.25.6-alpine AS builder
WORKDIR /app
COPY ../go.mod go.sum ./
RUN go mod download
COPY .. .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/transaction-monitoring-system

FROM alpine:3.22
WORKDIR /
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
EXPOSE 8082
EXPOSE 8084
ENTRYPOINT ["/server"]