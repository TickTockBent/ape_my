# Gap Analysis: ape_my as an API Mock Server

## Context

During development of [Puck](https://github.com/TickTockBent/puck) (an X/Twitter MCP server), we evaluated using ape_my as a mock API server for integration testing. The X API v2 was the target, but these gaps apply broadly to mocking any real-world REST API.

## What Works Today

ape_my v0.1.0 already handles:

- Stateful CRUD for arbitrary entities from a JSON schema
- Seed data for realistic test fixtures
- Correct HTTP status codes (200, 201, 204, 400, 404)
- Request validation against the schema
- Zero-config startup from a single schema file

For simple APIs with flat entities and standard REST patterns (`GET /entity`, `POST /entity`, `GET /entity/:id`), ape_my works perfectly.

## Gaps When Mocking Real APIs

### 1. Path Prefix / Base Path

**The gap:** Real APIs use versioned paths like `/2/tweets/:id` or `/api/v1/users`. ape_my generates routes at the root (`/tweets/:id`).

**What's needed:** A `basePath` field in the schema or a CLI flag:
```json
{
  "basePath": "/2",
  "entities": { ... }
}
```
Or: `ape_my schema.json on 3000 at /2`

**Complexity:** Low. Prepend the base path to all generated routes.

### 2. Response Envelope / Wrapper

**The gap:** Many APIs wrap responses in an envelope. X API v2 returns:
```json
{
  "data": { "id": "123", "text": "hello" },
  "includes": { "users": [...] },
  "meta": { "result_count": 1, "next_token": "abc" }
}
```

ape_my returns the entity directly:
```json
{ "id": "123", "text": "hello" }
```

**What's needed:** An optional `responseWrapper` in the schema:
```json
{
  "responseWrapper": {
    "single": { "data": "$entity" },
    "list": { "data": "$entities", "meta": { "result_count": "$count" } }
  }
}
```

**Complexity:** Medium. Needs template substitution for the wrapper fields and different wrapping for single vs. list responses.

### 3. Custom Route Patterns

**The gap:** Real APIs have non-standard routes:
- `GET /2/users/:id/tweets` (nested resource)
- `GET /2/tweets/search/recent?query=...` (query endpoint)
- `GET /2/users/me` (aliased endpoint)
- `POST /2/media/upload/initialize` (multi-step workflows)

ape_my only generates flat CRUD routes per entity.

**What's needed:** A `routes` override in the schema:
```json
{
  "entities": {
    "tweets": { ... }
  },
  "routes": [
    {
      "method": "GET",
      "path": "/users/:userId/tweets",
      "entity": "tweets",
      "filter": { "user_id": ":userId" }
    },
    {
      "method": "GET",
      "path": "/users/me",
      "entity": "users",
      "filter": { "id": "1" }
    }
  ]
}
```

**Complexity:** High. This is essentially a mini route DSL with parameter extraction and entity filtering.

### 4. Query Parameter Filtering

**The gap:** Most real APIs support filtering via query params: `GET /tweets?author_id=123&max_results=10`. ape_my's `GET /entity` returns all entities with no filtering.

**What's needed:** Automatic query parameter → field matching for list endpoints. Any query param that matches a field name filters by that value.

**Complexity:** Low-Medium. Compare query params against entity field names, filter the in-memory collection.

### 5. Pagination

**The gap:** Real APIs paginate large result sets with cursors or offset/limit. X API v2 uses `max_results` + `pagination_token` in the request and returns `meta.next_token` in the response. ape_my returns all entities at once.

**What's needed:** Support for `limit` (or `max_results`) and cursor-based pagination with `next_token` generation.

**Complexity:** Medium. Need to slice results, generate opaque cursor tokens, and return them in the response meta.

### 6. Custom Response Headers

**The gap:** X API returns rate limit info in response headers (`x-rate-limit-limit`, `x-rate-limit-remaining`, `x-rate-limit-reset`). Some clients depend on these for rate limiting logic.

**What's needed:** Optional custom headers per entity or globally:
```json
{
  "responseHeaders": {
    "x-rate-limit-limit": "450",
    "x-rate-limit-remaining": "449",
    "x-rate-limit-reset": "$timestamp+900"
  }
}
```

**Complexity:** Low. Static headers are trivial. Dynamic values (like decrementing remaining) are medium.

### 7. Authentication Simulation

**The gap:** ape_my accepts all requests with no auth checks. Real API mocking benefits from at least validating that an `Authorization: Bearer <token>` header is present, so client auth code gets exercised.

**What's needed:** An optional `auth` field:
```json
{
  "auth": {
    "type": "bearer",
    "token": "mock-token-123"
  }
}
```

Returns 401 if the header is missing or doesn't match.

**Complexity:** Low.

## Prioritized Feature Roadmap

If the goal is "mock any real REST API for testing," here's the suggested order based on impact vs. complexity:

| Priority | Feature | Complexity | Impact |
|----------|---------|-----------|--------|
| 1 | Base path prefix | Low | High — almost every real API uses versioned paths |
| 2 | Query parameter filtering | Low-Medium | High — essential for realistic list endpoints |
| 3 | Response envelope/wrapper | Medium | High — most modern APIs wrap responses |
| 4 | Pagination | Medium | Medium — needed for any API with large datasets |
| 5 | Custom response headers | Low | Medium — rate limit testing, CORS |
| 6 | Auth simulation | Low | Medium — exercises client auth code paths |
| 7 | Custom route patterns | High | Medium — needed for nested/non-standard APIs |

## Workaround Used for Puck

Since ape_my couldn't mock the X API v2 structure, Puck's test suite mocks at the `twitter-api-v2` library level using vitest's `vi.mock()`. This tests all of Puck's logic (rate limiting, error mapping, thread chaining, field resolution) without a running server. It's effective but doesn't exercise the HTTP layer.

With features 1-4 from the roadmap above, ape_my could serve as a full HTTP-level mock for Puck and similar API connector projects.
