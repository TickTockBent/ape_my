package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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
