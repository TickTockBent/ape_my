# Ape_my v0.1.0 MVP - Phased Build Plan

## Overview

This document outlines the phased development plan to take Ape_my from concept to a functional v0.1.0 MVP. Each phase builds upon the previous one, with tasks designed to be as modular as possible.

---

## Phase 0: Project Foundation (Setup & Design)

**Goal**: Establish project structure and core specifications

- **0.1** Initialize Go module and project structure
- **0.2** Define JSON schema format specification
- **0.3** Design internal data structures (schema model, storage interface)
- **0.4** Create basic project documentation (README, contributing guidelines)
- **0.5** Set up basic testing framework

---

## Phase 1: Command-Line Interface (CLI Parsing)

**Goal**: Parse and validate user commands

- **1.1** Implement CLI argument parser (handle natural language syntax)
- **1.2** Validate schema file path and existence
- **1.3** Parse optional "with seed_data.json" clause
- **1.4** Parse optional "on PORT" clause
- **1.5** Set default port (8080) when not specified
- **1.6** Add `--help` and `--version` flags

---

## Phase 2: Schema Processing (Core Logic)

**Goal**: Load and parse JSON schemas into internal models

- **2.1** Load JSON schema file from disk
- **2.2** Validate schema structure (entities, fields, types)
- **2.3** Parse entity definitions and field types
- **2.4** Build internal route map from schema entities
- **2.5** Handle schema parsing errors gracefully
- **2.6** Support basic JSON types (string, number, boolean, object, array)

---

## Phase 3: Data Storage (In-Memory Store)

**Goal**: Implement stateful in-memory data storage

- **3.1** Create in-memory storage interface
- **3.2** Implement CRUD operations on storage layer
- **3.3** Generate unique IDs for new entities (UUIDs or incremental)
- **3.4** Handle concurrent access safely (mutex/RWMutex)
- **3.5** Initialize empty collections for each entity type
- **3.6** Optional: Load seed data into storage on startup

---

## Phase 4: HTTP Server Core (Request Routing)

**Goal**: Set up HTTP server with dynamic route registration

- **4.1** Initialize `net/http` server on specified port
- **4.2** Dynamically register routes from schema (e.g., `/users`, `/posts`)
- **4.3** Implement route pattern matching for IDs (e.g., `/users/{id}`)
- **4.4** Add middleware for logging requests
- **4.5** Add middleware for content-type validation (JSON only)
- **4.6** Handle 404 for undefined routes

---

## Phase 5: CRUD Operations (HTTP Handlers)

**Goal**: Implement all RESTful CRUD endpoints

- **5.1** **POST** `/entities` - Create new entity
- **5.2** **GET** `/entities` - List all entities
- **5.3** **GET** `/entities/{id}` - Get single entity by ID
- **5.4** **PUT** `/entities/{id}` - Replace entire entity
- **5.5** **PATCH** `/entities/{id}` - Partially update entity
- **5.6** **DELETE** `/entities/{id}` - Delete entity
- **5.7** Return proper HTTP status codes (200, 201, 204, 404, 400, 500)
- **5.8** Validate request bodies against schema

---

## Phase 6: Error Handling & Validation (Robustness)

**Goal**: Proper error responses and validation

- **6.1** Validate JSON request bodies
- **6.2** Return structured error responses (JSON format)
- **6.3** Handle malformed JSON gracefully
- **6.4** Validate required fields from schema
- **6.5** Type validation for schema fields
- **6.6** Handle missing entity IDs (404 responses)

---

## Phase 7: Testing & Quality (Verification)

**Goal**: Ensure reliability and correctness

- **7.1** Unit tests for CLI parsing
- **7.2** Unit tests for schema loading/parsing
- **7.3** Unit tests for storage layer
- **7.4** Integration tests for HTTP endpoints (each CRUD operation)
- **7.5** Test concurrent request handling
- **7.6** Test seed data loading
- **7.7** Test error cases and validation

---

## Phase 8: Documentation & Polish (Release Prep)

**Goal**: Prepare for v0.1.0 release

- **8.1** Complete README with usage examples
- **8.2** Document JSON schema format specification
- **8.3** Add example schema and seed data files
- **8.4** Create simple usage guide
- **8.5** Add MIT license file
- **8.6** Build and test cross-platform binaries (Linux, macOS, Windows)
- **8.7** Tag v0.1.0 release

---

## MVP Feature Checklist

- [x] Single command execution: `go ape_my schema.json`
- [x] Optional seed data: `with seed_data.json`
- [x] Optional port specification: `on PORT`
- [x] Stateful in-memory storage (starts empty)
- [x] Full CRUD operations (POST, GET, PUT, PATCH, DELETE)
- [x] Auto-generated RESTful routes from schema
- [x] Proper HTTP status codes
- [x] JSON content negotiation
- [x] Basic error handling and validation
- [x] Single binary distribution (via `go build`)

---

## Out of Scope for v0.1.0

The following features are explicitly excluded from the initial MVP to maintain focus on core functionality:

- Persistent storage (databases)
- Authentication/authorization
- Query parameters and filtering
- Pagination
- Relationships between entities
- WebSockets or real-time features
- API documentation generation (OpenAPI/Swagger)
- Rate limiting
- CORS configuration (can be added as simple enhancement)

---

## Estimated Timeline

**Phase 0-2**: ~2-3 days (foundation)
- Project setup, CLI parsing, schema processing

**Phase 3-5**: ~3-4 days (core functionality)
- Storage layer, HTTP server, CRUD operations

**Phase 6-8**: ~2-3 days (polish & testing)
- Error handling, testing, documentation

**Total estimate**: ~7-10 days of focused development

---

## Success Criteria for v0.1.0

The MVP will be considered complete when:

1. A user can run `go ape_my schema.json` and get a working API
2. All CRUD operations work correctly for all entities defined in schema
3. The server handles errors gracefully with appropriate HTTP status codes
4. Basic test coverage exists for core functionality
5. Documentation allows a new user to get started in under 5 minutes
6. A single binary can be built for multiple platforms

---

## Next Steps

1. Begin Phase 0: Initialize Go project structure
2. Define the JSON schema format specification
3. Set up basic testing framework
4. Proceed through phases sequentially, validating each phase before moving forward
