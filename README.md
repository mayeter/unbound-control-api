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
- Zone management capabilities
- Zone file management for authoritative DNS

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

### Unbound Control
- `GET /api/v1/status` - Get Unbound server status
- `POST /api/v1/reload` - Reload Unbound configuration
- `POST /api/v1/flush` - Flush DNS cache
- `GET /api/v1/stats` - Get Unbound statistics
- `GET /api/v1/info` - Get detailed server information

### Zone Management
- `GET /api/v1/zones` - List all configured zones
- `POST /api/v1/zones` - Add a new zone
- `GET /api/v1/zones/{name}` - Get zone details
- `PUT /api/v1/zones/{name}` - Update zone configuration
- `DELETE /api/v1/zones/{name}` - Remove a zone

#### Zone Configuration Example
```json
{
  "name": "example.com",
  "type": "primary",
  "file": "/etc/unbound/zones/example.com.zone"
}
```

```json
{
  "name": "example.org",
  "type": "secondary",
  "masters": ["192.168.1.10", "192.168.1.11"]
}
```

```json
{
  "name": "example.net",
  "type": "forward",
  "forwards": ["8.8.8.8", "8.8.4.4"]
}
```

### Zone File Management
- `GET /api/v1/zones/{name}/file` - Get zone file content
- `PUT /api/v1/zones/{name}/file` - Update entire zone file
- `POST /api/v1/zones/{name}/records` - Add a new record
- `GET /api/v1/zones/{name}/records/{recordName}/{recordType}` - Get record details
- `PUT /api/v1/zones/{name}/records/{recordName}/{recordType}` - Update record
- `DELETE /api/v1/zones/{name}/records/{recordName}/{recordType}` - Remove record

#### Zone File Example
```json
{
  "name": "example.com",
  "records": [
    {
      "name": "@",
      "ttl": 3600,
      "class": "IN",
      "type": "SOA",
      "rdata": "ns1.example.com. admin.example.com. 2024031501 7200 3600 1209600 3600",
      "comments": "Start of Authority"
    },
    {
      "name": "@",
      "ttl": 3600,
      "class": "IN",
      "type": "NS",
      "rdata": "ns1.example.com.",
      "comments": "Primary nameserver"
    },
    {
      "name": "@",
      "ttl": 3600,
      "class": "IN",
      "type": "A",
      "rdata": "192.168.1.10",
      "comments": "Main IP address"
    },
    {
      "name": "www",
      "ttl": 3600,
      "class": "IN",
      "type": "CNAME",
      "rdata": "@",
      "comments": "WWW subdomain"
    }
  ]
}
```

#### Record Example
```json
{
  "name": "mail",
  "ttl": 3600,
  "class": "IN",
  "type": "MX",
  "rdata": "10 mail.example.com.",
  "comments": "Mail server"
}
```

## Security

- All API endpoints require authentication using an API key
- Communication with Unbound uses TLS encryption
- Rate limiting to prevent abuse
- Input validation for all commands

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
