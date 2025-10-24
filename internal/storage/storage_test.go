package storage

import (
	"sync"
	"testing"
)

func TestNewInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()

	if store == nil {
		t.Fatal("expected store to not be nil")
	}

	if store.data == nil {
		t.Error("expected data map to be initialized")
	}

	if store.counter == nil {
		t.Error("expected counter map to be initialized")
	}
}

func TestInitialize(t *testing.T) {
	store := NewInMemoryStore()

	entityTypes := []string{"users", "posts"}
	err := store.Initialize(entityTypes)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Check that entity types were initialized
	for _, entityType := range entityTypes {
		if store.data[entityType] == nil {
			t.Errorf("expected %s to be initialized", entityType)
		}
		if _, exists := store.counter[entityType]; !exists {
			t.Errorf("expected counter for %s to be initialized", entityType)
		}
	}
}

func TestCreate(t *testing.T) {
	store := NewInMemoryStore()
	store.Initialize([]string{"users"})

	tests := []struct {
		name       string
		entityType string
		data       map[string]interface{}
		wantErr    bool
		wantID     string
	}{
		{
			name:       "create with auto-generated ID",
			entityType: "users",
			data:       map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
			wantErr:    false,
			wantID:     "1",
		},
		{
			name:       "create with provided ID",
			entityType: "users",
			data:       map[string]interface{}{"id": "custom-123", "name": "Bob"},
			wantErr:    false,
			wantID:     "custom-123",
		},
		{
			name:       "create in non-existent entity type",
			entityType: "nonexistent",
			data:       map[string]interface{}{"name": "Charlie"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := store.Create(tt.entityType, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if id != tt.wantID {
					t.Errorf("Create() id = %v, want %v", id, tt.wantID)
				}

				// Verify entity was stored
				entity, err := store.Get(tt.entityType, id)
				if err != nil {
					t.Errorf("Get() after Create() error = %v", err)
				}
				if entity["id"] != id {
					t.Errorf("stored entity id = %v, want %v", entity["id"], id)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	store := NewInMemoryStore()
	store.Initialize([]string{"users"})

	// Create a test entity
	testData := map[string]interface{}{"name": "Alice", "email": "alice@example.com"}
	id, _ := store.Create("users", testData)

	tests := []struct {
		name       string
		entityType string
		id         string
		wantErr    bool
	}{
		{
			name:       "get existing entity",
			entityType: "users",
			id:         id,
			wantErr:    false,
		},
		{
			name:       "get non-existent entity",
			entityType: "users",
			id:         "999",
			wantErr:    true,
		},
		{
			name:       "get from non-existent entity type",
			entityType: "nonexistent",
			id:         "1",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := store.Get(tt.entityType, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && entity == nil {
				t.Error("expected entity to not be nil")
			}
		})
	}
}

func TestList(t *testing.T) {
	store := NewInMemoryStore()
	store.Initialize([]string{"users"})

	// Create multiple entities
	store.Create("users", map[string]interface{}{"name": "Alice"})
	store.Create("users", map[string]interface{}{"name": "Bob"})
	store.Create("users", map[string]interface{}{"name": "Charlie"})

	tests := []struct {
		name       string
		entityType string
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "list all entities",
			entityType: "users",
			wantCount:  3,
			wantErr:    false,
		},
		{
			name:       "list from non-existent entity type",
			entityType: "nonexistent",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities, err := store.List(tt.entityType)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(entities) != tt.wantCount {
				t.Errorf("List() returned %d entities, want %d", len(entities), tt.wantCount)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	store := NewInMemoryStore()
	store.Initialize([]string{"users"})

	// Create a test entity
	testData := map[string]interface{}{"name": "Alice", "email": "alice@example.com"}
	id, _ := store.Create("users", testData)

	tests := []struct {
		name       string
		entityType string
		id         string
		data       map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "update existing entity",
			entityType: "users",
			id:         id,
			data:       map[string]interface{}{"name": "Alice Updated", "email": "newemail@example.com"},
			wantErr:    false,
		},
		{
			name:       "update non-existent entity",
			entityType: "users",
			id:         "999",
			data:       map[string]interface{}{"name": "Nobody"},
			wantErr:    true,
		},
		{
			name:       "update in non-existent entity type",
			entityType: "nonexistent",
			id:         "1",
			data:       map[string]interface{}{"name": "Nobody"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Update(tt.entityType, tt.id, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify update
				entity, _ := store.Get(tt.entityType, tt.id)
				if entity["name"] != tt.data["name"] {
					t.Errorf("Update() name = %v, want %v", entity["name"], tt.data["name"])
				}
			}
		})
	}
}

func TestPatch(t *testing.T) {
	store := NewInMemoryStore()
	store.Initialize([]string{"users"})

	// Create a test entity
	testData := map[string]interface{}{"name": "Alice", "email": "alice@example.com", "age": 30}
	id, _ := store.Create("users", testData)

	tests := []struct {
		name       string
		entityType string
		id         string
		data       map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "patch existing entity",
			entityType: "users",
			id:         id,
			data:       map[string]interface{}{"age": 31}, // Only update age
			wantErr:    false,
		},
		{
			name:       "patch non-existent entity",
			entityType: "users",
			id:         "999",
			data:       map[string]interface{}{"name": "Nobody"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Patch(tt.entityType, tt.id, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("Patch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify patch - should have updated field and kept others
				entity, _ := store.Get(tt.entityType, tt.id)
				if entity["age"] != tt.data["age"] {
					t.Errorf("Patch() age = %v, want %v", entity["age"], tt.data["age"])
				}
				if entity["name"] != "Alice" {
					t.Errorf("Patch() name = %v, want Alice (should be unchanged)", entity["name"])
				}
				if entity["email"] != "alice@example.com" {
					t.Errorf("Patch() email = %v, want alice@example.com (should be unchanged)", entity["email"])
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	store := NewInMemoryStore()
	store.Initialize([]string{"users"})

	// Create a test entity
	testData := map[string]interface{}{"name": "Alice"}
	id, _ := store.Create("users", testData)

	tests := []struct {
		name       string
		entityType string
		id         string
		wantErr    bool
	}{
		{
			name:       "delete existing entity",
			entityType: "users",
			id:         id,
			wantErr:    false,
		},
		{
			name:       "delete non-existent entity",
			entityType: "users",
			id:         "999",
			wantErr:    true,
		},
		{
			name:       "delete from non-existent entity type",
			entityType: "nonexistent",
			id:         "1",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Delete(tt.entityType, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify deletion
				_, err := store.Get(tt.entityType, tt.id)
				if err != ErrNotFound {
					t.Errorf("Get() after Delete() error = %v, want ErrNotFound", err)
				}
			}
		})
	}
}

func TestSeed(t *testing.T) {
	store := NewInMemoryStore()
	store.Initialize([]string{"users"})

	seedData := []map[string]interface{}{
		{"id": "1", "name": "Alice", "email": "alice@example.com"},
		{"id": "2", "name": "Bob", "email": "bob@example.com"},
		{"id": "3", "name": "Charlie", "email": "charlie@example.com"},
	}

	err := store.Seed("users", seedData)
	if err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	// Verify all entities were seeded
	entities, _ := store.List("users")
	if len(entities) != 3 {
		t.Errorf("Seed() loaded %d entities, want 3", len(entities))
	}

	// Verify we can get individual entities
	entity, err := store.Get("users", "1")
	if err != nil {
		t.Errorf("Get() after Seed() error = %v", err)
	}
	if entity["name"] != "Alice" {
		t.Errorf("Seeded entity name = %v, want Alice", entity["name"])
	}

	// Verify counter was updated
	newID, _ := store.Create("users", map[string]interface{}{"name": "David"})
	if newID == "1" || newID == "2" || newID == "3" {
		t.Errorf("Create() after Seed() generated duplicate ID: %v", newID)
	}
}

func TestConcurrentAccess(t *testing.T) {
	store := NewInMemoryStore()
	store.Initialize([]string{"users"})

	// Create some initial data
	for i := 0; i < 10; i++ {
		store.Create("users", map[string]interface{}{"name": "User"})
	}

	// Concurrent reads and writes
	var wg sync.WaitGroup
	iterations := 100

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				store.List("users")
			}
		}()
	}

	// Concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				store.Create("users", map[string]interface{}{"name": "ConcurrentUser"})
			}
		}()
	}

	wg.Wait()

	// Verify data integrity
	entities, err := store.List("users")
	if err != nil {
		t.Errorf("List() after concurrent access error = %v", err)
	}

	expectedCount := 10 + (5 * iterations)
	if len(entities) != expectedCount {
		t.Errorf("After concurrent access, got %d entities, want %d", len(entities), expectedCount)
	}
}

func TestFormatID(t *testing.T) {
	tests := []struct {
		counter int
		want    string
	}{
		{1, "1"},
		{5, "5"},
		{9, "9"},
		{10, "10"},
		{42, "42"},
		{100, "100"},
		{999, "999"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatID(tt.counter)
			if got != tt.want {
				t.Errorf("formatID(%d) = %v, want %v", tt.counter, got, tt.want)
			}
		})
	}
}

func TestParseIDNumber(t *testing.T) {
	tests := []struct {
		id   string
		want int
	}{
		{"1", 1},
		{"5", 5},
		{"10", 10},
		{"42", 42},
		{"100", 100},
		{"999", 999},
		{"abc", 0},
		{"12abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := parseIDNumber(tt.id)
			if got != tt.want {
				t.Errorf("parseIDNumber(%s) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestCopyMap(t *testing.T) {
	original := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	}

	copied := copyMap(original)

	// Verify contents are the same
	if copied["name"] != original["name"] {
		t.Errorf("copied name = %v, want %v", copied["name"], original["name"])
	}

	// Verify it's a separate map (modifying copy doesn't affect original)
	copied["name"] = "Bob"
	if original["name"] == "Bob" {
		t.Error("modifying copied map affected original map")
	}
}
