# --- Build stage ---
FROM golang:1.22 as builder
WORKDIR /app
COPY api/ .  
RUN go build -o /unbound-control-api ./cmd/api

# --- Runtime stage ---
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /unbound-control-api .
COPY config.yaml .
# The certs will be mounted at /unbound-certs by docker-compose
CMD ["./unbound-control-api"]