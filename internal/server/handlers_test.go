package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ticktockbent/ape_my/internal/storage"
)

func TestHandleCreate(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name          string
		entityType    string
		body          string
		wantStatus    int
		checkResponse bool
		checkID       bool
	}{
		{
			name:          "create valid entity",
			entityType:    "users",
			body:          `{"name": "Alice", "email": "alice@example.com"}`,
			wantStatus:    http.StatusCreated,
			checkResponse: true,
			checkID:       true,
		},
		{
			name:       "create with invalid JSON",
			entityType: "users",
			body:       `{invalid json}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "create in non-existent entity type",
			entityType: "nonexistent",
			body:       `{"name": "Bob"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:          "create with custom ID",
			entityType:    "users",
			body:          `{"id": "custom-123", "name": "Charlie"}`,
			wantStatus:    http.StatusCreated,
			checkResponse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/"+tt.entityType, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.checkResponse {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if tt.checkID {
					if _, exists := response["id"]; !exists {
						t.Error("response should contain 'id' field")
					}
				}
			}
		})
	}
}

func TestHandleList(t *testing.T) {
	server := setupTestServer(t)

	// Create some test entities
	server.store.Create("users", map[string]interface{}{"name": "Alice"})
	server.store.Create("users", map[string]interface{}{"name": "Bob"})
	server.store.Create("users", map[string]interface{}{"name": "Charlie"})

	tests := []struct {
		name       string
		entityType string
		wantStatus int
		wantCount  int
	}{
		{
			name:       "list all users",
			entityType: "users",
			wantStatus: http.StatusOK,
			wantCount:  3,
		},
		{
			name:       "list non-existent entity type",
			entityType: "nonexistent",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "list empty collection",
			entityType: "posts",
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.entityType, http.NoBody)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response []map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if len(response) != tt.wantCount {
					t.Errorf("response count = %d, want %d", len(response), tt.wantCount)
				}
			}
		})
	}
}

func TestHandleGetOne(t *testing.T) {
	server := setupTestServer(t)

	// Create a test entity
	id, _ := server.store.Create("users", map[string]interface{}{"name": "Alice", "email": "alice@example.com"})

	tests := []struct {
		name       string
		entityType string
		id         string
		wantStatus int
	}{
		{
			name:       "get existing entity",
			entityType: "users",
			id:         id,
			wantStatus: http.StatusOK,
		},
		{
			name:       "get non-existent entity",
			entityType: "users",
			id:         "999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "get from non-existent entity type",
			entityType: "nonexistent",
			id:         "1",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.entityType+"/"+tt.id, http.NoBody)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response["id"] != tt.id {
					t.Errorf("response id = %v, want %v", response["id"], tt.id)
				}
			}
		})
	}
}

func TestHandleUpdate(t *testing.T) {
	server := setupTestServer(t)

	// Create a test entity
	id, _ := server.store.Create("users", map[string]interface{}{"name": "Alice", "email": "alice@example.com"})

	tests := []struct {
		name       string
		entityType string
		id         string
		body       string
		wantStatus int
	}{
		{
			name:       "update existing entity",
			entityType: "users",
			id:         id,
			body:       `{"name": "Alice Updated", "email": "newemail@example.com"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "update non-existent entity",
			entityType: "users",
			id:         "999",
			body:       `{"name": "Nobody"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "update with invalid JSON",
			entityType: "users",
			id:         id,
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/"+tt.entityType+"/"+tt.id, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				// Verify ID is preserved
				if response["id"] != tt.id {
					t.Errorf("response id = %v, want %v", response["id"], tt.id)
				}
			}
		})
	}
}

func TestHandlePatch(t *testing.T) {
	server := setupTestServer(t)

	// Create a test entity
	id, _ := server.store.Create("users", map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	})

	tests := []struct {
		name       string
		entityType string
		id         string
		body       string
		wantStatus int
		checkField string
		wantValue  interface{}
	}{
		{
			name:       "patch existing entity",
			entityType: "users",
			id:         id,
			body:       `{"age": 31}`,
			wantStatus: http.StatusOK,
			checkField: "age",
			wantValue:  float64(31), // JSON numbers are float64
		},
		{
			name:       "patch non-existent entity",
			entityType: "users",
			id:         "999",
			body:       `{"name": "Nobody"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "patch with invalid JSON",
			entityType: "users",
			id:         id,
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/"+tt.entityType+"/"+tt.id, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK && tt.checkField != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response[tt.checkField] != tt.wantValue {
					t.Errorf("response[%s] = %v, want %v", tt.checkField, response[tt.checkField], tt.wantValue)
				}

				// Verify other fields are unchanged
				if response["name"] != "Alice" {
					t.Errorf("name should be unchanged, got %v", response["name"])
				}
			}
		})
	}
}

func TestHandleDelete(t *testing.T) {
	server := setupTestServer(t)

	// Create a test entity
	id, _ := server.store.Create("users", map[string]interface{}{"name": "Alice"})

	tests := []struct {
		name       string
		entityType string
		id         string
		wantStatus int
	}{
		{
			name:       "delete existing entity",
			entityType: "users",
			id:         id,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "delete non-existent entity",
			entityType: "users",
			id:         "999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "delete from non-existent entity type",
			entityType: "nonexistent",
			id:         "1",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/"+tt.entityType+"/"+tt.id, http.NoBody)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			// Verify deletion if successful
			if tt.wantStatus == http.StatusNoContent {
				_, err := server.store.Get(tt.entityType, tt.id)
				if err != storage.ErrNotFound {
					t.Errorf("entity should be deleted, but Get returned: %v", err)
				}
			}
		})
	}
}

func TestFullCRUDWorkflow(t *testing.T) {
	server := setupTestServer(t)

	// 1. Create an entity
	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name": "Alice", "email": "alice@example.com"}`))
	createReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, createReq)

	if w.Code != http.StatusCreated {
		t.Fatalf("Create failed with status %d", w.Code)
	}

	var created map[string]interface{}
	json.NewDecoder(w.Body).Decode(&created)
	id := created["id"].(string)

	// 2. Get the entity
	getReq := httptest.NewRequest(http.MethodGet, "/users/"+id, http.NoBody)
	w = httptest.NewRecorder()
	server.mux.ServeHTTP(w, getReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Get failed with status %d", w.Code)
	}

	// 3. List entities
	listReq := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w = httptest.NewRecorder()
	server.mux.ServeHTTP(w, listReq)

	if w.Code != http.StatusOK {
		t.Fatalf("List failed with status %d", w.Code)
	}

	var list []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) < 1 {
		t.Fatal("List should contain at least one entity")
	}

	// 4. Update the entity
	updateReq := httptest.NewRequest(http.MethodPut, "/users/"+id, bytes.NewBufferString(`{"name": "Alice Updated", "email": "newemail@example.com"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.mux.ServeHTTP(w, updateReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Update failed with status %d", w.Code)
	}

	var updated map[string]interface{}
	json.NewDecoder(w.Body).Decode(&updated)
	if updated["name"] != "Alice Updated" {
		t.Errorf("name not updated, got %v", updated["name"])
	}

	// 5. Patch the entity
	patchReq := httptest.NewRequest(http.MethodPatch, "/users/"+id, bytes.NewBufferString(`{"email": "patched@example.com"}`))
	patchReq.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.mux.ServeHTTP(w, patchReq)

	if w.Code != http.StatusOK {
		t.Fatalf("Patch failed with status %d", w.Code)
	}

	var patched map[string]interface{}
	json.NewDecoder(w.Body).Decode(&patched)
	if patched["email"] != "patched@example.com" {
		t.Errorf("email not patched, got %v", patched["email"])
	}
	if patched["name"] != "Alice Updated" {
		t.Errorf("name should be preserved, got %v", patched["name"])
	}

	// 6. Delete the entity
	deleteReq := httptest.NewRequest(http.MethodDelete, "/users/"+id, http.NoBody)
	w = httptest.NewRecorder()
	server.mux.ServeHTTP(w, deleteReq)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Delete failed with status %d", w.Code)
	}

	// 7. Verify deletion
	getAfterDelete := httptest.NewRequest(http.MethodGet, "/users/"+id, http.NoBody)
	w = httptest.NewRecorder()
	server.mux.ServeHTTP(w, getAfterDelete)

	if w.Code != http.StatusNotFound {
		t.Errorf("Get after delete should return 404, got %d", w.Code)
	}
}

func TestConcurrentHTTPRequests(t *testing.T) {
	server := setupTestServer(t)

	var wg sync.WaitGroup
	var successfulCreates atomic.Int32
	var successfulReads atomic.Int32
	var successfulUpdates atomic.Int32
	numGoroutines := 10
	requestsPerGoroutine := 20

	// Concurrent POST requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				body := bytes.NewBufferString(`{"name": "ConcurrentUser", "email": "test@example.com"}`)
				req := httptest.NewRequest(http.MethodPost, "/users", body)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.mux.ServeHTTP(w, req)

				if w.Code == http.StatusCreated {
					successfulCreates.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify creates
	expectedCreates := int32(numGoroutines * requestsPerGoroutine)
	if successfulCreates.Load() != expectedCreates {
		t.Errorf("expected %d successful creates, got %d", expectedCreates, successfulCreates.Load())
	}

	// Get all created IDs
	req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	w := httptest.NewRecorder()
	server.mux.ServeHTTP(w, req)

	var users []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&users)

	if len(users) != int(expectedCreates) {
		t.Errorf("expected %d users, got %d", expectedCreates, len(users))
	}

	// Concurrent READ requests
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				req := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
				w := httptest.NewRecorder()

				server.mux.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					successfulReads.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	// Verify reads
	expectedReads := int32(numGoroutines * requestsPerGoroutine)
	if successfulReads.Load() != expectedReads {
		t.Errorf("expected %d successful reads, got %d", expectedReads, successfulReads.Load())
	}

	// Concurrent UPDATE requests on different entities
	for i := 0; i < numGoroutines && i < len(users); i++ {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				body := bytes.NewBufferString(`{"name": "UpdatedUser", "email": "updated@example.com"}`)
				req := httptest.NewRequest(http.MethodPut, "/users/"+userID, body)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.mux.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					successfulUpdates.Add(1)
				}
			}
		}(users[i]["id"].(string))
	}

	wg.Wait()

	// Verify updates (at least some should succeed)
	if successfulUpdates.Load() == 0 {
		t.Error("expected at least some successful updates")
	}

	t.Logf("Concurrent test stats: Creates=%d, Reads=%d, Updates=%d",
		successfulCreates.Load(), successfulReads.Load(), successfulUpdates.Load())
}
