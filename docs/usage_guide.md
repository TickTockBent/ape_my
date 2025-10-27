# Ape_my Usage Guide

## Table of Contents
1. [Getting Started](#getting-started)
2. [Command-Line Usage](#command-line-usage)
3. [Creating Your First API](#creating-your-first-api)
4. [Working with the API](#working-with-the-api)
5. [Advanced Usage](#advanced-usage)
6. [Troubleshooting](#troubleshooting)

---

## Getting Started

### Installation

#### Option 1: Install via `go install` (Recommended)

```bash
go install github.com/ticktockbent/ape_my/cmd/ape_my@latest
```

This installs the latest version to your `$GOPATH/bin` directory.

#### Option 2: Build from Source

```bash
git clone https://github.com/ticktockbent/ape_my.git
cd ape_my
go build -o bin/ape_my ./cmd/ape_my
```

#### Option 3: Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/ticktockbent/ape_my/releases).

### Verify Installation

```bash
ape_my --version
```

---

## Command-Line Usage

Ape_my uses a natural language syntax for intuitive command-line operations.

### Basic Syntax

```bash
ape_my <schema-file> [with <seed-file>] [on <port>]
```

### Arguments

| Argument | Required | Description | Example |
|----------|----------|-------------|---------|
| `schema-file` | Yes | Path to JSON schema file | `schema.json` |
| `with seed-file` | No | Path to seed data file | `with seed.json` |
| `on port` | No | Custom port number (default: 8080) | `on 3000` |

### Flags

| Flag | Description |
|------|-------------|
| `-h, --help` | Show help message |
| `-v, --version` | Show version information |

### Examples

```bash
# Start with an empty API on default port 8080
ape_my schema.json

# Start with seed data
ape_my schema.json with seed.json

# Start on a custom port
ape_my schema.json on 3000

# Combine seed data and custom port
ape_my schema.json with seed.json on 8080

# Show help
ape_my --help

# Show version
ape_my --version
```

---

## Creating Your First API

### Step 1: Define Your Schema

Create a file named `todos_schema.json`:

```json
{
  "entities": {
    "todos": {
      "fields": {
        "id": {
          "type": "string",
          "required": true
        },
        "task": {
          "type": "string",
          "required": true
        },
        "completed": {
          "type": "boolean",
          "required": false
        },
        "priority": {
          "type": "number",
          "required": false
        }
      }
    }
  }
}
```

### Step 2: (Optional) Create Seed Data

Create a file named `todos_seed.json`:

```json
{
  "todos": [
    {
      "id": "1",
      "task": "Buy groceries",
      "completed": false,
      "priority": 2
    },
    {
      "id": "2",
      "task": "Walk the dog",
      "completed": true,
      "priority": 1
    }
  ]
}
```

### Step 3: Start the Server

```bash
ape_my todos_schema.json with todos_seed.json
```

You should see output like:

```
ape_my v0.1.0
Configuration: Schema: todos_schema.json, Seed: todos_seed.json, Port: 8080

Loading schema...
Loaded 1 entities: [todos]
Initializing storage...
Loading seed data from todos_seed.json...
Seeded 2 todos
Registered routes: /todos and /todos/

=== Ape_my is ready! ===
API endpoints available:
  - /todos (GET, POST)
  - /todos/<id> (GET, PUT, PATCH, DELETE)

Starting server on http://localhost:8080
Press Ctrl+C to stop
```

---

## Working with the API

### CREATE - Add a new todo

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Finish documentation",
    "completed": false,
    "priority": 3
  }'
```

**Response** (201 Created):
```json
{
  "id": "3",
  "task": "Finish documentation",
  "completed": false,
  "priority": 3
}
```

### READ - Get all todos

```bash
curl http://localhost:8080/todos
```

**Response** (200 OK):
```json
[
  {
    "id": "1",
    "task": "Buy groceries",
    "completed": false,
    "priority": 2
  },
  {
    "id": "2",
    "task": "Walk the dog",
    "completed": true,
    "priority": 1
  },
  {
    "id": "3",
    "task": "Finish documentation",
    "completed": false,
    "priority": 3
  }
]
```

### READ - Get a specific todo

```bash
curl http://localhost:8080/todos/1
```

**Response** (200 OK):
```json
{
  "id": "1",
  "task": "Buy groceries",
  "completed": false,
  "priority": 2
}
```

### UPDATE - Replace an entire todo (PUT)

```bash
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Buy groceries and cook dinner",
    "completed": false,
    "priority": 1
  }'
```

**Response** (200 OK):
```json
{
  "id": "1",
  "task": "Buy groceries and cook dinner",
  "completed": false,
  "priority": 1
}
```

### UPDATE - Partially update a todo (PATCH)

```bash
curl -X PATCH http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{
    "completed": true
  }'
```

**Response** (200 OK):
```json
{
  "id": "1",
  "task": "Buy groceries and cook dinner",
  "completed": true,
  "priority": 1
}
```

### DELETE - Remove a todo

```bash
curl -X DELETE http://localhost:8080/todos/1
```

**Response** (204 No Content)

---

## Advanced Usage

### Multiple Entities

You can define multiple entities in a single schema:

```json
{
  "entities": {
    "users": {
      "fields": {
        "id": {"type": "string", "required": true},
        "name": {"type": "string", "required": true},
        "email": {"type": "string", "required": true}
      }
    },
    "posts": {
      "fields": {
        "id": {"type": "string", "required": true},
        "title": {"type": "string", "required": true},
        "content": {"type": "string", "required": true},
        "authorId": {"type": "string", "required": false}
      }
    }
  }
}
```

This creates:
- `/users` and `/users/:id` endpoints
- `/posts` and `/posts/:id` endpoints

### Field Types

Ape_my supports all JSON types:

```json
{
  "entities": {
    "examples": {
      "fields": {
        "id": {"type": "string", "required": true},
        "text": {"type": "string", "required": false},
        "count": {"type": "number", "required": false},
        "enabled": {"type": "boolean", "required": false},
        "metadata": {"type": "object", "required": false},
        "tags": {"type": "array", "required": false}
      }
    }
  }
}
```

### Auto-generated IDs

If you don't provide an `id` field when creating an entity, Ape_my will generate one automatically:

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Auto-generated ID example"
  }'
```

**Response**:
```json
{
  "id": "4",
  "task": "Auto-generated ID example",
  "completed": null,
  "priority": null
}
```

### Testing with HTTPie

If you prefer HTTPie over curl:

```bash
# Create
http POST localhost:8080/todos task="Test with HTTPie" completed:=false priority:=1

# List all
http GET localhost:8080/todos

# Get one
http GET localhost:8080/todos/1

# Update
http PUT localhost:8080/todos/1 task="Updated task" completed:=true priority:=1

# Partial update
http PATCH localhost:8080/todos/1 completed:=true

# Delete
http DELETE localhost:8080/todos/1
```

---

## Troubleshooting

### Server won't start

**Problem**: Port already in use

```
Error: listen tcp :8080: bind: address already in use
```

**Solution**: Use a different port

```bash
ape_my schema.json on 3000
```

---

### Schema validation error

**Problem**: Missing required field

```
Error: schema validation failed: entity "todos": field "id" is missing
```

**Solution**: Ensure all entities have an `id` field with type `string`

```json
{
  "entities": {
    "todos": {
      "fields": {
        "id": {
          "type": "string",
          "required": true
        }
      }
    }
  }
}
```

---

### Request validation error

**Problem**: Missing required field in request

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Response** (400 Bad Request):
```json
{
  "error": "validation failed: required field 'task' is missing"
}
```

**Solution**: Include all required fields in your request

---

### 415 Unsupported Media Type

**Problem**: Missing Content-Type header

```bash
curl -X POST http://localhost:8080/todos \
  -d '{"task": "Test"}'
```

**Response** (415):
```json
{
  "error": "Content-Type must be application/json"
}
```

**Solution**: Always include the Content-Type header for POST, PUT, and PATCH requests

```bash
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"task": "Test"}'
```

---

### 404 Not Found

**Problem**: Entity doesn't exist

```bash
curl http://localhost:8080/todos/999
```

**Response** (404):
```json
{
  "error": "entity not found"
}
```

**Solution**: Verify the entity ID exists by listing all entities first

---

## Best Practices

### 1. Use Descriptive Entity Names

```json
{
  "entities": {
    "users": { ... },         // Good
    "posts": { ... },         // Good
    "blogPosts": { ... },     // Good
    "x": { ... }              // Bad - not descriptive
  }
}
```

### 2. Define Required Fields Appropriately

Mark fields as required only if they are absolutely necessary for entity creation.

### 3. Use Appropriate Field Types

- Use `string` for text, IDs, and enums
- Use `number` for counts, prices, and numeric values
- Use `boolean` for flags and toggles
- Use `object` for nested data
- Use `array` for lists

### 4. Keep Seed Data Realistic

Use realistic test data in your seed files to make testing more effective.

### 5. Version Your Schemas

Keep your schema files in version control and track changes over time.

---

## Next Steps

- Explore the [Schema Format Documentation](schema_format.md)
- Check out the [API Reference](api_reference.md)
- Read about the [Development Roadmap](build_plan_v0.1.0.md)
- Contribute to the project on [GitHub](https://github.com/ticktockbent/ape_my)

---

## Getting Help

- **Documentation**: https://github.com/ticktockbent/ape_my/tree/main/docs
- **Issues**: https://github.com/ticktockbent/ape_my/issues
- **Discussions**: https://github.com/ticktockbent/ape_my/discussions

---

**Happy API Mocking!** 🎉
