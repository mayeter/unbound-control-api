# Unbound Control API

A REST API for managing Unbound DNS resolver remotely. This project provides a secure HTTP interface to control Unbound instances using its native control interface.

## Features

- Direct communication with Unbound's control interface
- RESTful API endpoints for common operations
- Secure authentication and authorization
- Support for all Unbound control commands
- Easy to deploy and configure
- Hot-reloadable configuration
- TLS certificate reloading

## Prerequisites

- Go 1.17 or later
- Unbound DNS resolver
- Access to Unbound's control interface (default port: 8953)

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/unbound-control-api.git
cd unbound-control-api

# Build the project
go build -o unbound-control-api ./cmd/api

# Run the server
./unbound-control-api
```

## Configuration

The API can be configured using environment variables or a configuration file:

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  use_tls: true
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"

unbound:
  control_port: 8953
  control_host: "127.0.0.1"
  control_key: "/path/to/control.key"
  control_cert: "/path/to/control.cert"

security:
  api_key: "your-secure-api-key"

rate_limit:
  requests_per_second: 10
  burst_size: 20

logging:
  level: "info"
  use_syslog: false
  app_name: "unbound-control-api"
```

### Hot-Reloadable Configuration

The API supports hot-reloading of configuration using the SIGHUP signal. The following settings can be updated without restarting the server:

- **Security Settings**:
  - API key (`security.api_key`)
  - TLS certificates (`server.cert_file`, `server.key_file`)

- **Rate Limiting**:
  - Requests per second (`rate_limit.requests_per_second`)
  - Burst size (`rate_limit.burst_size`)

- **Unbound Connection Settings**:
  - Control host (`unbound.control_host`)
  - Control port (`unbound.control_port`)
  - Control certificate (`unbound.control_cert`)
  - Control key (`unbound.control_key`)

To reload the configuration, send a SIGHUP signal to the process:
```bash
kill -HUP <pid>
```

Note: The following settings require a server restart to take effect:
- Server host and port (`server.host`, `server.port`)
- TLS enablement (`server.use_tls`)
- Logging configuration (`logging.*`)

## API Endpoints

- `GET /api/v1/status` - Get Unbound server status
- `POST /api/v1/reload` - Reload Unbound configuration
- `POST /api/v1/flush` - Flush DNS cache
- `GET /api/v1/stats` - Get Unbound statistics
- `GET /api/v1/info` - Get detailed server information

## Security

- All API endpoints require authentication using an API key
- Communication with Unbound uses TLS encryption
- Rate limiting to prevent abuse
- Input validation for all commands

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
