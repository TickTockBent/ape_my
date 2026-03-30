package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ticktockbent/ape_my/internal/schema"
	"github.com/ticktockbent/ape_my/internal/storage"
)

func setupTestSchema(t *testing.T) *schema.Loader {
	// Create a temporary schema file for testing
	schemaData := `{
		"entities": {
			"users": {
				"fields": {
					"id": {
						"type": "string",
						"required": true
					},
					"name": {
						"type": "string",
						"required": true
					},
					"email": {
						"type": "string",
						"required": false
					},
					"age": {
						"type": "number",
						"required": false
					}
				},
				"collectionPath": "/users"
			},
			"posts": {
				"fields": {
					"id": {
						"type": "string",
						"required": true
					},
					"title": {
						"type": "string",
						"required": true
					},
					"content": {
						"type": "string",
						"required": false
					}
				},
				"collectionPath": "/posts"
			}
		}
	}`

	// Create temp file
	tmpFile := t.TempDir() + "/test-schema.json"
	if err := os.WriteFile(tmpFile, []byte(schemaData), 0o644); err != nil {
		t.Fatalf("failed to create test schema file: %v", err)
	}

	// Load schema
	loader := schema.NewLoader()
	if err := loader.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("failed to load test schema: %v", err)
	}

	return loader
}

func setupTestServer(t *testing.T) *Server {
	store := storage.NewInMemoryStore()
	store.Initialize([]string{"users", "posts"})

	routeMap := schema.RouteMap{
		"users": {
			EntityName:     "users",
			CollectionPath: "/users",
			ItemPath:       "/users/{id}",
		},
		"posts": {
			EntityName:     "posts",
			CollectionPath: "/posts",
			ItemPath:       "/posts/{id}",
		},
	}

	loader := setupTestSchema(t)
	server := New(8080, store, routeMap, loader)
	server.RegisterRoutes()

	return server
}

func TestNew(t *testing.T) {
	store := storage.NewInMemoryStore()
	routeMap := schema.RouteMap{}
	loader := schema.NewLoader()

	server := New(8080, store, routeMap, loader)

	if server == nil {
		t.Fatal("expected server to not be nil")
	}

	if server.port != 8080 {
		t.Errorf("expected port 8080, got %d", server.port)
	}

	if server.store == nil {
		t.Error("expected store to not be nil")
	}

	if server.mux == nil {
		t.Error("expected mux to not be nil")
	}

	if server.validator == nil {
		t.Error("expected validator to not be nil")
	}
}

func TestRegisterRoutes(t *testing.T) {
	server := setupTestServer(t)

	// Test that routes respond with CRUD operations
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"GET collection", http.MethodGet, "/users", http.StatusOK},
		{"POST collection", http.MethodPost, "/users", http.StatusUnsupportedMediaType}, // No Content-Type
		{"GET item", http.MethodGet, "/users/1", http.StatusNotFound},                   // Entity doesn't exist
		{"PUT item", http.MethodPut, "/users/1", http.StatusUnsupportedMediaType},       // No Content-Type
		{"PATCH item", http.MethodPatch, "/users/1", http.StatusUnsupportedMediaType},   // No Content-Type
		{"DELETE item", http.MethodDelete, "/users/1", http.StatusNotFound},             // Entity doesn't exist
		{"unknown route", http.MethodGet, "/unknown", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestMiddleware_Logging(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	server.mux.ServeHTTP(w, req)

	// Check that response has JSON content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", contentType)
	}
}

func TestMiddleware_ContentType(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name        string
		method      string
		path        string
		contentType string
		wantStatus  int
	}{
		{
			name:        "POST with JSON",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest, // Empty body
		},
		{
			name:        "POST without content-type",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "",
			wantStatus:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "POST with wrong content-type",
			method:      http.MethodPost,
			path:        "/users",
			contentType: "text/plain",
			wantStatus:  http.StatusUnsupportedMediaType,
		},
		{
			name:        "PUT with JSON",
			method:      http.MethodPut,
			path:        "/users/1",
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest, // Empty body
		},
		{
			name:        "PATCH with JSON",
			method:      http.MethodPatch,
			path:        "/users/1",
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest, // Empty body
		},
		{
			name:        "GET doesn't require content-type",
			method:      http.MethodGet,
			path:        "/users",
			contentType: "",
			wantStatus:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandle404(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"root path", "/", http.StatusNotFound},
		{"unknown path", "/unknown", http.StatusNotFound},
		{"unknown nested path", "/unknown/nested", http.StatusNotFound},
		{"valid collection path", "/users", http.StatusOK},
		{"valid item path", "/users/123", http.StatusNotFound}, // Entity doesn't exist
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d for path %s", w.Code, tt.wantStatus, tt.path)
			}
		})
	}
}

func TestHandleItem_IDExtraction(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"valid ID", "/users/123", http.StatusNotFound},                  // Entity doesn't exist
		{"valid alphanumeric ID", "/users/abc-123", http.StatusNotFound}, // Entity doesn't exist
		{"nested path (invalid)", "/users/123/nested", http.StatusNotFound},
		{"no ID", "/users/", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d for path %s", w.Code, tt.wantStatus, tt.path)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"OPTIONS on collection", http.MethodOptions, "/users"},
		{"HEAD on collection", http.MethodHead, "/users"},
		{"OPTIONS on item", http.MethodOptions, "/users/1"},
		{"HEAD on item", http.MethodHead, "/users/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestRespondJSON(t *testing.T) {
	server := setupTestServer(t)

	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	server.respondJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "test") {
		t.Errorf("body = %s, want to contain 'test'", body)
	}
}

func TestBasePathRouting(t *testing.T) {
	// Setup a server with a base path prefix — using "users" entity which exists in test schema
	store := storage.NewInMemoryStore()
	store.Initialize([]string{"users"})

	routeMap := schema.RouteMap{
		"users": {
			EntityName:     "users",
			CollectionPath: "/api/v2/users",
			ItemPath:       "/api/v2/users/{id}",
		},
	}

	loader := setupTestSchema(t)
	srv := New(8080, store, routeMap, loader)
	srv.RegisterRoutes()

	// Create a user via the prefixed path
	createReq := httptest.NewRequest(http.MethodPost, "/api/v2/users", strings.NewReader(`{"name": "Alice"}`))
	createReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, createReq)

	if w.Code != http.StatusCreated {
		t.Fatalf("POST /api/v2/users: status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	// List users via the prefixed path
	listReq := httptest.NewRequest(http.MethodGet, "/api/v2/users", http.NoBody)
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, listReq)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v2/users: status = %d, want %d", w.Code, http.StatusOK)
	}

	// Get a user by ID via the prefixed path
	getReq := httptest.NewRequest(http.MethodGet, "/api/v2/users/1", http.NoBody)
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, getReq)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v2/users/1: status = %d, want %d", w.Code, http.StatusOK)
	}

	// Non-prefixed path should 404
	badReq := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, badReq)

	if w.Code != http.StatusNotFound {
		t.Errorf("GET /users (no base path): status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func setupTestServerWithSchema(t *testing.T, schemaJSON string) *Server {
	// Write the schema to a temp file and load it
	tmpFile := t.TempDir() + "/test-schema.json"
	if err := os.WriteFile(tmpFile, []byte(schemaJSON), 0o644); err != nil {
		t.Fatalf("failed to write test schema: %v", err)
	}
	loader := schema.NewLoader()
	if err := loader.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("failed to load test schema: %v", err)
	}

	// Initialize store with all entity types from the schema
	store := storage.NewInMemoryStore()
	store.Initialize(loader.GetEntityNames())

	routeMap, err := loader.BuildRouteMap()
	if err != nil {
		t.Fatalf("failed to build route map: %v", err)
	}

	srv := New(8080, store, routeMap, loader)
	srv.RegisterRoutes()
	return srv
}

func TestCustomResponseHeaders(t *testing.T) {
	schemaJSON := `{
		"responseHeaders": {
			"x-rate-limit-limit": "100",
			"x-rate-limit-remaining": "99",
			"x-custom-header": "mock-server"
		},
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify custom headers are present
	if got := w.Header().Get("x-rate-limit-limit"); got != "100" {
		t.Errorf("x-rate-limit-limit = %q, want %q", got, "100")
	}
	if got := w.Header().Get("x-rate-limit-remaining"); got != "99" {
		t.Errorf("x-rate-limit-remaining = %q, want %q", got, "99")
	}
	if got := w.Header().Get("x-custom-header"); got != "mock-server" {
		t.Errorf("x-custom-header = %q, want %q", got, "mock-server")
	}
	// Content-Type should not be overridden
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
}

func TestCustomHeadersDoNotOverrideProtected(t *testing.T) {
	schemaJSON := `{
		"responseHeaders": {
			"Content-Type": "text/plain",
			"Content-Length": "999"
		},
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	// Content-Type should remain application/json
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
}

func TestAuthMiddleware(t *testing.T) {
	schemaJSON := `{
		"auth": {
			"token": "test-secret-token"
		},
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "valid token",
			authHeader: "Bearer test-secret-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing auth header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong token",
			authHeader: "Bearer wrong-token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong auth type",
			authHeader: "Basic dGVzdDp0ZXN0",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			srv.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestNoAuthWhenNotConfigured(t *testing.T) {
	// Default server has no auth configured — all requests should pass
	server := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (no auth configured should allow all)", w.Code, http.StatusOK)
	}
}

func TestQueryParameterFiltering(t *testing.T) {
	schemaJSON := `{
		"entities": {
			"users": {
				"fields": {
					"id":    {"type": "string", "required": true},
					"name":  {"type": "string", "required": true},
					"email": {"type": "string", "required": false}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	// Seed some data
	srv.store.Create("users", map[string]interface{}{"name": "Alice", "email": "alice@example.com"})
	srv.store.Create("users", map[string]interface{}{"name": "Bob", "email": "bob@example.com"})
	srv.store.Create("users", map[string]interface{}{"name": "Alice", "email": "alice2@example.com"})

	tests := []struct {
		name      string
		query     string
		wantCount int
	}{
		{"no filter", "/users", 3},
		{"filter by name", "/users?name=Alice", 2},
		{"filter by email", "/users?email=bob@example.com", 1},
		{"unknown param ignored", "/users?unknown=value", 3},
		{"no match", "/users?name=Nobody", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.query, http.NoBody)
			w := httptest.NewRecorder()
			srv.mux.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
			}

			var response []map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if len(response) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(response), tt.wantCount)
			}
		})
	}
}

func TestResponseWrapper(t *testing.T) {
	schemaJSON := `{
		"responseWrapper": {
			"single": {"data": "$entity"},
			"list": {"data": "$entities", "meta": {"result_count": "$count"}}
		},
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	// Create a user
	createReq := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name": "Alice"}`))
	createReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, createReq)

	if w.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	// Check wrapped single response
	var createResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&createResp)
	if _, ok := createResp["data"]; !ok {
		t.Fatalf("wrapped response should have 'data' key, got: %v", createResp)
	}

	// Check wrapped list response
	listReq := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, listReq)

	var listResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&listResp)
	if _, ok := listResp["data"]; !ok {
		t.Fatalf("wrapped list response should have 'data' key, got: %v", listResp)
	}
	if _, ok := listResp["meta"]; !ok {
		t.Fatalf("wrapped list response should have 'meta' key, got: %v", listResp)
	}
}

func TestPaginationCursor(t *testing.T) {
	schemaJSON := `{
		"pagination": {
			"style": "cursor",
			"defaultLimit": 2,
			"maxLimit": 5
		},
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	// Create 5 users
	for i := 0; i < 5; i++ {
		body := fmt.Sprintf(`{"name": "User%d"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.mux.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("Create user %d failed: %d", i, w.Code)
		}
	}

	// First page (default limit 2)
	req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	var firstPage map[string]interface{}
	json.NewDecoder(w.Body).Decode(&firstPage)

	data, ok := firstPage["data"].([]interface{})
	if !ok {
		// If no wrapper, response is a plain array — pagination with no wrapper still returns meta
		t.Logf("response: %v", firstPage)
	}
	if ok && len(data) != 2 {
		t.Errorf("first page: got %d items, want 2", len(data))
	}
}

func TestPaginationOffset(t *testing.T) {
	schemaJSON := `{
		"pagination": {
			"style": "offset",
			"defaultLimit": 2,
			"maxLimit": 10
		},
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	// Create 5 users
	for i := 0; i < 5; i++ {
		body := fmt.Sprintf(`{"name": "User%d"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.mux.ServeHTTP(w, req)
	}

	// Get first page (limit=2, offset=0) — response has data/meta since totalCount > limit
	req := httptest.NewRequest(http.MethodGet, "/users?limit=2&offset=0", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatalf("expected 'data' array in response, got: %v", resp)
	}
	if len(data) != 2 {
		t.Errorf("first page: got %d items, want 2", len(data))
	}

	// Get second page
	req2 := httptest.NewRequest(http.MethodGet, "/users?limit=2&offset=2", http.NoBody)
	w2 := httptest.NewRecorder()
	srv.mux.ServeHTTP(w2, req2)

	var resp2 map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&resp2)
	data2, ok := resp2["data"].([]interface{})
	if !ok {
		t.Fatalf("expected 'data' array in response, got: %v", resp2)
	}
	if len(data2) != 2 {
		t.Errorf("second page: got %d items, want 2", len(data2))
	}
}

func TestCustomRoutes(t *testing.T) {
	schemaJSON := `{
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			},
			"tweets": {
				"fields": {
					"id":        {"type": "string", "required": true},
					"text":      {"type": "string", "required": true},
					"author_id": {"type": "string", "required": false}
				}
			}
		},
		"routes": [
			{
				"method": "GET",
				"path": "/users/:userId/tweets",
				"entity": "tweets",
				"filters": {"userId": "author_id"}
			},
			{
				"method": "GET",
				"path": "/users/me",
				"entity": "users",
				"filters": {"id": "1"}
			}
		]
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	// Seed tweets with author_id
	srv.store.Create("tweets", map[string]interface{}{"text": "Hello from Alice", "author_id": "1"})
	srv.store.Create("tweets", map[string]interface{}{"text": "Another from Alice", "author_id": "1"})
	srv.store.Create("tweets", map[string]interface{}{"text": "Hello from Bob", "author_id": "2"})

	// Seed users
	srv.store.Create("users", map[string]interface{}{"id": "1", "name": "Alice"})

	// Test nested resource route: /users/1/tweets
	req := httptest.NewRequest(http.MethodGet, "/users/1/tweets", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /users/1/tweets: status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var tweets []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&tweets)
	if len(tweets) != 2 {
		t.Errorf("GET /users/1/tweets: got %d tweets, want 2", len(tweets))
	}

	// Test /users/2/tweets should only return Bob's tweet
	req2 := httptest.NewRequest(http.MethodGet, "/users/2/tweets", http.NoBody)
	w2 := httptest.NewRecorder()
	srv.mux.ServeHTTP(w2, req2)

	var tweets2 []map[string]interface{}
	json.NewDecoder(w2.Body).Decode(&tweets2)
	if len(tweets2) != 1 {
		t.Errorf("GET /users/2/tweets: got %d tweets, want 1", len(tweets2))
	}

	// Test alias route: /users/me should return user with id "1"
	req3 := httptest.NewRequest(http.MethodGet, "/users/me", http.NoBody)
	w3 := httptest.NewRecorder()
	srv.mux.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("GET /users/me: status = %d, want %d, body: %s", w3.Code, http.StatusOK, w3.Body.String())
	}

	var user map[string]interface{}
	json.NewDecoder(w3.Body).Decode(&user)
	if user["name"] != "Alice" {
		t.Errorf("GET /users/me: name = %v, want Alice", user["name"])
	}
}

func TestCustomRoutesWithBasePath(t *testing.T) {
	schemaJSON := `{
		"basePath": "/api/v2",
		"entities": {
			"tweets": {
				"fields": {
					"id":        {"type": "string", "required": true},
					"text":      {"type": "string", "required": true},
					"author_id": {"type": "string", "required": false}
				}
			}
		},
		"routes": [
			{
				"method": "GET",
				"path": "/users/:userId/tweets",
				"entity": "tweets",
				"filters": {"userId": "author_id"}
			}
		]
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	srv.store.Create("tweets", map[string]interface{}{"text": "Hello", "author_id": "42"})

	// Custom route should be prefixed with basePath
	req := httptest.NewRequest(http.MethodGet, "/api/v2/users/42/tweets", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v2/users/42/tweets: status = %d, want %d, body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var tweets []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&tweets)
	if len(tweets) != 1 {
		t.Errorf("got %d tweets, want 1", len(tweets))
	}
}

func TestCustomRouteMethodRestriction(t *testing.T) {
	schemaJSON := `{
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		},
		"routes": [
			{
				"method": "GET",
				"path": "/users/me",
				"entity": "users",
				"filters": {"id": "1"}
			}
		]
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	// POST should be rejected
	req := httptest.NewRequest(http.MethodPost, "/users/me", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /users/me: status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestRespondJSON_NilData(t *testing.T) {
	server := setupTestServer(t)
	w := httptest.NewRecorder()
	server.respondJSON(w, http.StatusNoContent, nil)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if w.Body.Len() != 0 {
		t.Errorf("body should be empty, got %q", w.Body.String())
	}
}

func TestApplyTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template interface{}
		vars     map[string]interface{}
		check    func(t *testing.T, result interface{})
	}{
		{
			name:     "string variable substitution",
			template: "$entity",
			vars:     map[string]interface{}{"$entity": map[string]interface{}{"id": "1"}},
			check: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map, got %T", result)
				}
				if m["id"] != "1" {
					t.Errorf("expected id=1, got %v", m["id"])
				}
			},
		},
		{
			name:     "inline string substitution",
			template: "count is $count items",
			vars:     map[string]interface{}{"$count": 5},
			check: func(t *testing.T, result interface{}) {
				s, ok := result.(string)
				if !ok {
					t.Fatalf("expected string, got %T", result)
				}
				if s != "count is 5 items" {
					t.Errorf("expected 'count is 5 items', got %q", s)
				}
			},
		},
		{
			name:     "no variable match returns original string",
			template: "no vars here",
			vars:     map[string]interface{}{"$other": "val"},
			check: func(t *testing.T, result interface{}) {
				if result != "no vars here" {
					t.Errorf("expected original string, got %v", result)
				}
			},
		},
		{
			name:     "map template",
			template: map[string]interface{}{"data": "$entity", "status": "ok"},
			vars:     map[string]interface{}{"$entity": "resolved"},
			check: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map, got %T", result)
				}
				if m["data"] != "resolved" {
					t.Errorf("expected data=resolved, got %v", m["data"])
				}
				if m["status"] != "ok" {
					t.Errorf("expected status=ok, got %v", m["status"])
				}
			},
		},
		{
			name:     "array template",
			template: []interface{}{"$entity", "static"},
			vars:     map[string]interface{}{"$entity": "resolved"},
			check: func(t *testing.T, result interface{}) {
				arr, ok := result.([]interface{})
				if !ok {
					t.Fatalf("expected array, got %T", result)
				}
				if len(arr) != 2 {
					t.Fatalf("expected 2 items, got %d", len(arr))
				}
				if arr[0] != "resolved" {
					t.Errorf("expected first item=resolved, got %v", arr[0])
				}
			},
		},
		{
			name:     "float64 passthrough",
			template: float64(42),
			vars:     map[string]interface{}{},
			check: func(t *testing.T, result interface{}) {
				if result != float64(42) {
					t.Errorf("expected 42.0, got %v", result)
				}
			},
		},
		{
			name:     "bool passthrough",
			template: true,
			vars:     map[string]interface{}{},
			check: func(t *testing.T, result interface{}) {
				if result != true {
					t.Errorf("expected true, got %v", result)
				}
			},
		},
		{
			name:     "nil passthrough",
			template: nil,
			vars:     map[string]interface{}{},
			check: func(t *testing.T, result interface{}) {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyTemplate(tt.template, tt.vars)
			tt.check(t, result)
		})
	}
}

func TestHandleCreate_ValidationError(t *testing.T) {
	server := setupTestServer(t)

	// Send a number where a string is expected for "name"
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name": 123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleUpdate_ValidationError(t *testing.T) {
	server := setupTestServer(t)

	id, _ := server.store.Create("users", map[string]interface{}{"name": "Alice"})

	// Send wrong type for "name"
	req := httptest.NewRequest(http.MethodPut, "/users/"+id, strings.NewReader(`{"name": 123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePatch_ValidationError(t *testing.T) {
	server := setupTestServer(t)

	id, _ := server.store.Create("users", map[string]interface{}{"name": "Alice"})

	// Send wrong type for "age" (string instead of number)
	req := httptest.NewRequest(http.MethodPatch, "/users/"+id, strings.NewReader(`{"age": "not a number"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleCollection_MethodNotAllowed(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/users", http.NoBody)
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestPaginationMaxLimit(t *testing.T) {
	schemaJSON := `{
		"pagination": {
			"style": "offset",
			"defaultLimit": 100,
			"maxLimit": 5
		},
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	for i := 0; i < 10; i++ {
		body := fmt.Sprintf(`{"name": "User%d"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.mux.ServeHTTP(w, req)
	}

	// Request with limit exceeding maxLimit — should be capped to 5
	req := httptest.NewRequest(http.MethodGet, "/users?limit=100", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatalf("expected 'data' array, got: %v", resp)
	}
	if len(data) != 5 {
		t.Errorf("got %d items, want 5 (maxLimit)", len(data))
	}
}

func TestPaginationCursorNavigation(t *testing.T) {
	schemaJSON := `{
		"pagination": {
			"style": "cursor",
			"defaultLimit": 2
		},
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	for i := 0; i < 5; i++ {
		body := fmt.Sprintf(`{"name": "User%d"}`, i)
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.mux.ServeHTTP(w, req)
	}

	// First page
	req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	var page1 map[string]interface{}
	json.NewDecoder(w.Body).Decode(&page1)
	meta, _ := page1["meta"].(map[string]interface{})
	nextToken, _ := meta["next_token"].(string)

	if nextToken == "" {
		t.Fatal("expected next_token in first page")
	}

	// Second page using cursor
	req2 := httptest.NewRequest(http.MethodGet, "/users?cursor="+nextToken, http.NoBody)
	w2 := httptest.NewRecorder()
	srv.mux.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w2.Code, http.StatusOK)
	}
}

func TestRespondList_NoPagination(t *testing.T) {
	// Server with no pagination config returns raw array
	schemaJSON := `{
		"entities": {
			"users": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`
	srv := setupTestServerWithSchema(t, schemaJSON)

	srv.store.Create("users", map[string]interface{}{"name": "Alice"})
	srv.store.Create("users", map[string]interface{}{"name": "Bob"})

	req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	// Should return plain array, not wrapped
	var response []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("expected plain array response, got error: %v", err)
	}
	if len(response) != 2 {
		t.Errorf("got %d items, want 2", len(response))
	}
}

func TestConvertPathParams(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/users/:userId/tweets", "/users/{userId}/tweets"},
		{"/users/me", "/users/me"},
		{"/:a/:b/:c", "/{a}/{b}/{c}"},
		{"/no/params", "/no/params"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := convertPathParams(tt.input)
			if got != tt.want {
				t.Errorf("convertPathParams(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractParamNames(t *testing.T) {
	tests := []struct {
		path string
		want []string
	}{
		{"/users/:userId/tweets", []string{"userId"}},
		{"/users/me", nil},
		{"/:a/:b", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractParamNames(tt.path)
			if len(got) != len(tt.want) {
				t.Errorf("extractParamNames(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	server := setupTestServer(t)

	w := httptest.NewRecorder()
	server.respondError(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	body := w.Body.String()
	if !strings.Contains(body, "test error") {
		t.Errorf("body = %s, want to contain 'test error'", body)
	}
	if !strings.Contains(body, "error") {
		t.Errorf("body = %s, want to contain 'error' key", body)
	}
}

// setupMismatchedServer creates a server where "ghost" entity exists in the schema and routes
// but not in storage, allowing us to test ErrEntityTypeNotFound error paths inside handlers.
func setupMismatchedServer(t *testing.T) *Server {
	schemaJSON := `{
		"entities": {
			"users": {
				"fields": {
					"id":    {"type": "string", "required": true},
					"name":  {"type": "string", "required": true},
					"email": {"type": "string", "required": false}
				}
			},
			"ghost": {
				"fields": {
					"id":   {"type": "string", "required": true},
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`

	tmpFile := t.TempDir() + "/mismatch-schema.json"
	if err := os.WriteFile(tmpFile, []byte(schemaJSON), 0o644); err != nil {
		t.Fatalf("failed to write schema: %v", err)
	}
	loader := schema.NewLoader()
	if err := loader.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("failed to load schema: %v", err)
	}

	store := storage.NewInMemoryStore()
	store.Initialize([]string{"users"}) // Only initialize "users", not "ghost"

	routeMap := schema.RouteMap{
		"users": {EntityName: "users", CollectionPath: "/users", ItemPath: "/users/{id}"},
		"ghost": {EntityName: "ghost", CollectionPath: "/ghost", ItemPath: "/ghost/{id}"},
	}

	srv := New(8080, store, routeMap, loader)
	srv.RegisterRoutes()
	return srv
}

func TestHandleCreate_EntityTypeNotFound(t *testing.T) {
	srv := setupMismatchedServer(t)

	req := httptest.NewRequest(http.MethodPost, "/ghost", strings.NewReader(`{"name": "test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	// Entity passes validation (ghost is in schema) but store returns ErrEntityTypeNotFound
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestHandleList_EntityTypeNotFound(t *testing.T) {
	srv := setupMismatchedServer(t)

	req := httptest.NewRequest(http.MethodGet, "/ghost", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleGetOne_EntityTypeNotFound(t *testing.T) {
	srv := setupMismatchedServer(t)

	req := httptest.NewRequest(http.MethodGet, "/ghost/1", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandleUpdate_EntityTypeNotFound(t *testing.T) {
	srv := setupMismatchedServer(t)

	req := httptest.NewRequest(http.MethodPut, "/ghost/1", strings.NewReader(`{"name": "test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	// Ghost passes validation (in schema) but store returns ErrEntityTypeNotFound
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestHandlePatch_EntityTypeNotFound(t *testing.T) {
	srv := setupMismatchedServer(t)

	req := httptest.NewRequest(http.MethodPatch, "/ghost/1", strings.NewReader(`{"name": "test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	// Ghost passes validation (in schema) but store returns ErrEntityTypeNotFound
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestHandleDelete_EntityTypeNotFound(t *testing.T) {
	srv := setupMismatchedServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/ghost/1", http.NoBody)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestValidateUpdate_UnknownEntity(t *testing.T) {
	loader := setupTestSchema(t)
	v := NewValidator(loader)

	err := v.ValidateUpdate("nonexistent", map[string]interface{}{"name": "test"})
	if err == nil {
		t.Error("expected error for unknown entity type")
	}
}

func TestValidatePatch_UnknownEntity(t *testing.T) {
	loader := setupTestSchema(t)
	v := NewValidator(loader)

	err := v.ValidatePatch("nonexistent", map[string]interface{}{"name": "test"})
	if err == nil {
		t.Error("expected error for unknown entity type")
	}
}

func TestStartAndShutdown(t *testing.T) {
	store := storage.NewInMemoryStore()
	store.Initialize([]string{"users"})

	routeMap := schema.RouteMap{
		"users": {EntityName: "users", CollectionPath: "/users", ItemPath: "/users/{id}"},
	}

	loader := setupTestSchema(t)
	srv := New(0, store, routeMap, loader) // port 0 = random available port
	srv.RegisterRoutes()

	// Start in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Wait briefly for the server to initialize
	time.Sleep(50 * time.Millisecond)

	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// Start should return nil (ErrServerClosed is swallowed)
	if err := <-errCh; err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestShutdown_NilServer(t *testing.T) {
	store := storage.NewInMemoryStore()
	routeMap := schema.RouteMap{}
	loader := schema.NewLoader()
	srv := New(8080, store, routeMap, loader)

	// Shutdown without ever starting — should be a no-op
	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}
