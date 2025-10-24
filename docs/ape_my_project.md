# Ape_my

> A minimalist mock API server that apes your schema

## Overview

Ape_my is a zero-config mock API server that turns a simple JSON schema into a fully functional REST API with stateful CRUD operations. Perfect for frontend development, testing, and prototyping when your backend doesn't exist yet.

## The Name

The name works on multiple levels:
- **Semantically**: "Aping" means mimicking or imitating, which is exactly what a mock server does
- **Command line magic**: When installed via Go, you literally type `go ape_my schema.json` - as in "GO APE!"

## Core Concept

Define your API structure in a simple JSON schema file. Run one command. Get a working REST API with realistic endpoints and full CRUD support. No configuration files, no complex setup, no learning curve.

## Key Features

### Zero Configuration
- Single command to start: `go ape_my schema.json`
- Minimal schema format - just define your data structures
- No boilerplate, no config files

### Stateful by Default
- Starts empty (unless seeded)
- POST creates records
- GET retrieves data
- PUT/PATCH updates records
- DELETE removes records
- All mutations persist during the session

### Optional Seed Data
- Load initial data with: `go ape_my schema.json with seed_data.json`
- Useful for testing against known datasets
- Simple JSON format matching your schema

### Smart Conventions
- Auto-generates RESTful routes from schema
- Handles standard HTTP methods appropriately
- Returns proper status codes
- Content-Type negotiation

## Technology Stack

**Language**: Go

**Why Go?**
- **Single binary distribution** - No runtime dependencies, just download and run
- **Fast startup** - Compiled binary starts instantly
- **Built for network services** - Go's `net/http` is production-grade and simple
- **Native JSON support** - `encoding/json` in stdlib handles everything
- **Trivial concurrency** - Goroutines make handling multiple requests easy
- **Easy cross-compilation** - Build for all platforms from one machine
- **Simple installation** - `go install` or download binary from releases

## Schema Format

Simple, intuitive JSON structure defining your data models:
- Resource names as top-level keys
- Field names with type hints
- Type system that infers realistic mock data

## Command Examples

Start empty server:
```
go ape_my schema.json
```

Start with seed data:
```
go ape_my schema.json with seed_data.json
```

Specify port:
```
go ape_my schema.json on 3000
```

## Project Goals

1. **Simplicity First** - If it takes more than 30 seconds to understand, it's too complex
2. **Developer Experience** - The command should read like natural language
3. **Practical Utility** - Solve real problems in frontend/API development workflows
4. **Portfolio Showcase** - Demonstrate Go expertise alongside JavaScript, Python, and Rust projects

## Portfolio Context

This project complements:
- **Word Beauty** (JavaScript) - Creative web tool
- **liteconfig_py** (Python) - Configuration utilities
- **Taskline** (Rust) - System-level scheduling

Ape_my demonstrates proficiency in Go while solving a genuine pain point in modern development workflows.

## Distribution

- **GitHub repository** - Open source, MIT licensed
- **Go package** - Installable via `go install`
- **Binary releases** - Pre-compiled for major platforms (Linux, macOS, Windows)
- **Documentation** - Clear README with examples and use cases

## Future Enhancements (Post-MVP)

- Chaos mode (random latency, failures)
- Response schema validation
- Request logging
- GraphQL support
- WebSocket endpoints
- OpenAPI spec generation
- Docker image

## Success Metrics

A successful portfolio project:
- Demonstrates clean Go code and idiomatic patterns
- Solves a real problem developers face
- Has a memorable, shareable name/concept
- Is actually useful beyond just being a portfolio piece
- Shows understanding of REST APIs, HTTP, and developer workflows

---

**Status**: Pre-development  
**License**: MIT  
**Author**: [Your Name]
