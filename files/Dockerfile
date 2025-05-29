# --- Build stage ---
FROM golang:1.22 as builder
WORKDIR /app
COPY . .  
RUN GOOS=linux GOARCH=amd64 go build -o /opt/unbound/unbound-control-api ./cmd/api

# --- Runtime stage ---
FROM mvance/unbound:latest
WORKDIR /opt/unbound
COPY --from=builder /opt/unbound/unbound-control-api .
COPY ./files/config.yaml .
COPY ./files/wait-for-unbound.sh .
RUN chmod +x wait-for-unbound.sh

# Entrypoint: run unbound, wait for socket, then run API
CMD ["/opt/unbound/wait-for-unbound.sh"]