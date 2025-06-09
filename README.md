# Unbound Control API

A REST API for managing Unbound DNS resolver remotely. This project provides a secure HTTP interface to control Unbound instances using its native control interface.

## Features

- Direct communication with Unbound's control interface (via UNIX socket)
- RESTful API endpoints for common operations
- Secure authentication and authorization
- Support for all Unbound control commands
- Easy to deploy and configure
- Hot-reloadable configuration
- TLS certificate reloading (for the API, not Unbound)
- Zone management capabilities
- Zone file management for authoritative DNS

## Roadmap

### Phase 1: Core Unbound Control Interface (Current)
- [x] Basic Unbound control commands (status, reload, flush, stats)
- [x] UNIX socket communication
- [x] API authentication and rate limiting
- [ ] Complete mapping of all unbound-control commands:
  - [ ] List and manage local zones
  - [ ] List and manage forward zones
  - [ ] List and manage stub zones
  - [ ] Manage local data records
  - [ ] Cache management commands
  - [ ] Module management commands
  - [ ] DNSSEC management commands

### Phase 2: Advanced Zone Management
- [ ] Zone File Management:
  - [ ] Create and manage zone files in BIND format
  - [ ] Validate zone file syntax
  - [ ] Support for SOA record management
  - [ ] Support for all common record types (A, AAAA, MX, CNAME, etc.)
  - [ ] Zone file import/export
  - [ ] Zone file versioning and rollback

- [ ] Auth Zone Support:
  - [ ] Primary (master) zone configuration
  - [ ] Secondary (slave) zone configuration
  - [ ] Zone transfer (AXFR/IXFR) management
  - [ ] DNSSEC key management
  - [ ] Zone signing and validation

### Phase 3: Enhanced Features
- [ ] Zone Templates:
  - [ ] Predefined zone configurations
  - [ ] Common record patterns
  - [ ] Bulk zone creation

- [ ] Zone Monitoring:
  - [ ] Zone health checks
  - [ ] Record validation
  - [ ] DNSSEC status monitoring
  - [ ] Zone transfer status

- [ ] Advanced Security:
  - [ ] Role-based access control
  - [ ] Audit logging
  - [ ] Zone access policies
  - [ ] API key management

### Phase 4: Integration and Automation
- [ ] Webhook Support:
  - [ ] Zone change notifications
  - [ ] Health check alerts
  - [ ] Integration with external systems

- [ ] Automation Tools:
  - [ ] Zone deployment automation
  - [ ] Record update automation
  - [ ] Bulk operations API
  - [ ] Scheduled tasks

- [ ] Documentation and Examples:
  - [ ] API usage examples
  - [ ] Common use cases
  - [ ] Best practices
  - [ ] Integration guides

## API Response Structure

The API converts Unbound's text-based responses into structured JSON objects for better usability and consistency.

### Common Response Format
```json
{
  "success": true,
  "data": {
    // Command-specific data
  },
  "error": null
}
```

### Command-Specific Responses

#### Status Response
```json
{
  "success": true,
  "data": {
    "version": "1.22.0",
    "verbosity": 1,
    "threads": 4,
    "modules": ["validator", "iterator"],
    "uptime": {
      "seconds": 123,
      "formatted": "2m 3s"
    },
    "options": {
      "control": "open"
    }
  }
}
```

#### Statistics Response
```json
{
  "success": true,
  "data": {
    "queries": {
      "total": 1234,
      "ip_ratelimited": 0
    },
    "cache": {
      "hits": 567,
      "misses": 667,
      "prefetch": 0,
      "zero_ttl": 0
    },
    "recursion": {
      "replies": 667,
      "time": {
        "average": 0.012345,
        "median": 0.01
      }
    },
    "request_list": {
      "average": 0.5,
      "max": 1,
      "overwritten": 0,
      "exceeded": 0,
      "current": {
        "all": 0,
        "user": 0
      }
    },
    "tcp_usage": 0
  }
}
```


#### Error Response
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "INVALID_COMMAND",
    "message": "Unknown command: invalid_command",
    "details": "Available commands: status, stats, list_local_zones, ..."
  }
}
```

### Response Types

1. **Status Information**
   - Server status
   - Module status
   - Configuration status

2. **Statistics**
   - Query statistics
   - Cache statistics
   - Recursion statistics
   - Resource usage

3. **Operation Results**
   - Command success/failure
   - Operation details
   - Error information

### Implementation Plan

1. **Phase 1: Basic Response Structures**
   - [ ] Define Go structs for each response type
   - [ ] Implement response parsers for basic commands
   - [ ] Add response validation

2. **Phase 2: Enhanced Response Features**
   - [ ] Add formatted values (e.g., human-readable uptime)
   - [ ] Implement response caching
   - [ ] Add response compression

3. **Phase 3: Advanced Response Features**
   - [ ] Add response filtering
   - [ ] Implement response pagination
   - [ ] Add response metadata

## Prerequisites

- Go 1.17 or later (for development/building)
- Docker or Podman (for running the stack)
- Unbound DNS resolver (included in the container)

## Quick Start (Docker)

To run the API and Unbound together locally:

```bash
# Clone the repository
git clone https://github.com/mayeter/unbound-control-api.git
cd unbound-control-api/files

# Build and run the stack
docker-compose up --build
```

- This will build a single container image with both Unbound and the API.
- The API will be available on port 8080.
- Unbound and the API communicate via a UNIX socket (`/opt/unbound/unbound.sock`).

### **Note: UNIX Socket Only**
- This API **only supports controlling Unbound via a UNIX socket**.
- **Remote-control over TCP is NOT supported** in this version.
- This means the API and Unbound must run on the same machine/container.
- No TLS/certificates are needed for Unbound control (the socket file permissions provide security).

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
  control_socket: "/opt/unbound/unbound.sock"

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
  - Control socket path (`unbound.control_socket`)

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

## Security

- All API endpoints require authentication using an API key
- Communication with Unbound uses a UNIX socket (no TCP, no TLS)
- Rate limiting to prevent abuse
- Input validation for all commands

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
