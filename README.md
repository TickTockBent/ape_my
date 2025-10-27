# Ape_my

[![CI](https://github.com/ticktockbent/ape_my/workflows/CI/badge.svg)](https://github.com/ticktockbent/ape_my/actions)
[![codecov](https://codecov.io/gh/ticktockbent/ape_my/branch/main/graph/badge.svg)](https://codecov.io/gh/ticktockbent/ape_my)
[![Go Report Card](https://goreportcard.com/badge/github.com/ticktockbent/ape_my)](https://goreportcard.com/report/github.com/ticktockbent/ape_my)
[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/ticktockbent/ape_my)](https://github.com/ticktockbent/ape_my/releases)

> A minimalist mock API server that converts JSON schemas into fully functional REST APIs

Ape_my is a zero-configuration mock API server written in Go. Point it at a JSON schema file, and it instantly generates a stateful RESTful API - perfect for frontend development, testing, and rapid prototyping.

## Features

- **Zero Configuration**: Just run `go ape_my schema.json` and you're done
- **Stateful by Default**: In-memory storage with full CRUD operations
- **Natural Language Commands**: Intuitive CLI syntax
- **Auto-generated Routes**: RESTful endpoints created from your schema
- **Optional Seed Data**: Start with pre-populated data
- **Single Binary**: No dependencies, just download and run

## Quick Start

### Installation

```bash
go install github.com/ticktockbent/ape_my/cmd/ape_my@latest
```

Or build from source:

```bash
git clone https://github.com/ticktockbent/ape_my.git
cd ape_my
go build -o bin/ape_my ./cmd/ape_my
```

### Basic Usage

```bash
# Start with an empty API
ape_my schema.json

# Start with seed data
ape_my schema.json with seed.json

# Specify a custom port
ape_my schema.json on 3000

# Combine options
ape_my schema.json with seed.json on 8080
```

## Example

**schema.json**:
```json
{
  "entities": {
    "users": {
      "fields": {
        "id": {"type": "string", "required": true},
        "name": {"type": "string", "required": true},
        "email": {"type": "string", "required": true},
        "active": {"type": "boolean", "required": false}
      }
    }
  }
}
```

**Start the server**:
```bash
ape_my schema.json
```

**Use the API**:
```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "active": true}'

# Get all users
curl http://localhost:8080/users

# Get a specific user
curl http://localhost:8080/users/1

# Update a user
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Smith", "email": "alice@example.com", "active": true}'

# Partially update a user
curl -X PATCH http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"active": false}'

# Delete a user
curl -X DELETE http://localhost:8080/users/1
```

## Schema Format

See [docs/schema_format.md](docs/schema_format.md) for complete schema format documentation.

## Generated Endpoints

For each entity in your schema, Ape_my generates these RESTful endpoints:

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/entity` | Create a new entity |
| GET | `/entity` | List all entities |
| GET | `/entity/:id` | Get specific entity |
| PUT | `/entity/:id` | Replace entire entity |
| PATCH | `/entity/:id` | Partially update entity |
| DELETE | `/entity/:id` | Delete entity |

## Status Codes

Ape_my returns proper HTTP status codes:

- `200 OK` - Successful GET, PUT, PATCH
- `201 Created` - Successful POST
- `204 No Content` - Successful DELETE
- `400 Bad Request` - Invalid request body or validation error
- `404 Not Found` - Entity or endpoint not found
- `500 Internal Server Error` - Server error

## Project Status

**Current Version**: 0.1.0

## Why Go?

- **Single Binary Distribution**: No runtime dependencies
- **Fast Startup**: Near-instant server startup
- **Production-grade Standard Library**: Built on `net/http`
- **Native JSON Support**: Built-in JSON encoding/decoding
- **Easy Cross-compilation**: Build for any platform

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details

## Project Goals

Ape_my is designed to be:

1. **Simple**: No configuration files, no complex setup
2. **Fast**: Instant startup, minimal overhead
3. **Focused**: Does one thing well - mock REST APIs
4. **Portable**: Single binary, runs anywhere

## Roadmap

See [docs/build_plan_v0.1.0.md](docs/build_plan_v0.1.0.md) for detailed development phases.

Future enhancements beyond v0.1.0 may include:
- Query parameters and filtering
- Pagination support
- Entity relationships
- Custom validation rules
- OpenAPI/Swagger documentation generation
- CORS configuration

## Support

- **Issues**: [GitHub Issues](https://github.com/ticktockbent/ape_my/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ticktockbent/ape_my/discussions)

## Acknowledgments

Ape_my is inspired by tools like json-server but reimagined with Go's simplicity and performance.
