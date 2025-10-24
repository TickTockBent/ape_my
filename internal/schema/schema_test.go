package schema

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ticktockbent/ape_my/pkg/types"
)

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	validSchema := `{
		"entities": {
			"users": {
				"fields": {
					"id": {"type": "string", "required": true},
					"name": {"type": "string", "required": true},
					"email": {"type": "string", "required": true}
				}
			}
		}
	}`

	invalidJSON := `{invalid json}`

	emptySchema := `{"entities": {}}`

	noIDSchema := `{
		"entities": {
			"users": {
				"fields": {
					"name": {"type": "string", "required": true}
				}
			}
		}
	}`

	invalidTypeSchema := `{
		"entities": {
			"users": {
				"fields": {
					"id": {"type": "string", "required": true},
					"name": {"type": "invalid_type", "required": true}
				}
			}
		}
	}`

	tests := []struct {
		name        string
		schemaJSON  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid schema",
			schemaJSON: validSchema,
			wantErr:    false,
		},
		{
			name:        "invalid JSON",
			schemaJSON:  invalidJSON,
			wantErr:     true,
			errContains: "failed to parse schema JSON",
		},
		{
			name:        "empty schema",
			schemaJSON:  emptySchema,
			wantErr:     true,
			errContains: "schema contains no entities",
		},
		{
			name:        "no id field",
			schemaJSON:  noIDSchema,
			wantErr:     true,
			errContains: "must have an 'id' field",
		},
		{
			name:        "invalid field type",
			schemaJSON:  invalidTypeSchema,
			wantErr:     true,
			errContains: "invalid field type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp schema file
			schemaFile := filepath.Join(tmpDir, tt.name+".json")
			if err := os.WriteFile(schemaFile, []byte(tt.schemaJSON), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			loader := NewLoader()
			err := loader.LoadFromFile(schemaFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("LoadFromFile() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestGetEntityNames(t *testing.T) {
	loader := NewLoader()
	loader.schema = &types.Schema{
		Entities: map[string]*types.Entity{
			"users": {
				Fields: map[string]*types.Field{
					"id": {Type: types.FieldTypeString, Required: true},
				},
			},
			"posts": {
				Fields: map[string]*types.Field{
					"id": {Type: types.FieldTypeString, Required: true},
				},
			},
		},
	}

	names := loader.GetEntityNames()

	if len(names) != 2 {
		t.Errorf("expected 2 entity names, got %d", len(names))
	}

	// Check both names are present
	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	if !found["users"] || !found["posts"] {
		t.Errorf("expected to find 'users' and 'posts', got %v", names)
	}
}

func TestGetEntity(t *testing.T) {
	loader := NewLoader()
	loader.schema = &types.Schema{
		Entities: map[string]*types.Entity{
			"users": {
				Fields: map[string]*types.Field{
					"id": {Type: types.FieldTypeString, Required: true},
				},
			},
		},
	}

	// Test existing entity
	entity, exists := loader.GetEntity("users")
	if !exists {
		t.Error("expected entity 'users' to exist")
	}
	if entity == nil {
		t.Error("expected entity to not be nil")
	}

	// Test non-existing entity
	_, exists = loader.GetEntity("nonexistent")
	if exists {
		t.Error("expected entity 'nonexistent' to not exist")
	}
}

func TestLoadSeedData(t *testing.T) {
	tmpDir := t.TempDir()

	validSeed := `{
		"users": [
			{"id": "1", "name": "Alice", "email": "alice@example.com"}
		]
	}`

	invalidJSON := `{invalid}`

	tests := []struct {
		name        string
		seedJSON    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid seed data",
			seedJSON: validSeed,
			wantErr:  false,
		},
		{
			name:        "invalid JSON",
			seedJSON:    invalidJSON,
			wantErr:     true,
			errContains: "failed to parse seed JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seedFile := filepath.Join(tmpDir, tt.name+".json")
			if err := os.WriteFile(seedFile, []byte(tt.seedJSON), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			data, err := LoadSeedData(seedFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSeedData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && data == nil {
				t.Error("expected seed data to not be nil")
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("LoadSeedData() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidateSeedData(t *testing.T) {
	loader := NewLoader()
	loader.schema = &types.Schema{
		Entities: map[string]*types.Entity{
			"users": {
				Fields: map[string]*types.Field{
					"id":    {Type: types.FieldTypeString, Required: true},
					"name":  {Type: types.FieldTypeString, Required: true},
					"email": {Type: types.FieldTypeString, Required: true},
					"age":   {Type: types.FieldTypeNumber, Required: false},
				},
			},
		},
	}

	tests := []struct {
		name        string
		seedData    map[string][]map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "valid seed data",
			seedData: map[string][]map[string]interface{}{
				"users": {
					{"id": "1", "name": "Alice", "email": "alice@example.com"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			seedData: map[string][]map[string]interface{}{
				"users": {
					{"id": "1", "name": "Alice"},
				},
			},
			wantErr:     true,
			errContains: "required field",
		},
		{
			name: "unknown entity",
			seedData: map[string][]map[string]interface{}{
				"unknown": {
					{"id": "1"},
				},
			},
			wantErr:     true,
			errContains: "unknown entity",
		},
		{
			name: "wrong field type",
			seedData: map[string][]map[string]interface{}{
				"users": {
					{"id": "1", "name": "Alice", "email": "alice@example.com", "age": "not a number"},
				},
			},
			wantErr:     true,
			errContains: "expected number",
		},
		{
			name: "extra fields allowed",
			seedData: map[string][]map[string]interface{}{
				"users": {
					{"id": "1", "name": "Alice", "email": "alice@example.com", "extra": "field"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.ValidateSeedData(tt.seedData)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSeedData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateSeedData() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestBuildRouteMap(t *testing.T) {
	loader := NewLoader()
	loader.schema = &types.Schema{
		Entities: map[string]*types.Entity{
			"users": {
				Fields: map[string]*types.Field{
					"id": {Type: types.FieldTypeString, Required: true},
				},
			},
			"posts": {
				Fields: map[string]*types.Field{
					"id": {Type: types.FieldTypeString, Required: true},
				},
			},
		},
	}

	routeMap, err := loader.BuildRouteMap()
	if err != nil {
		t.Fatalf("BuildRouteMap() error = %v", err)
	}

	if len(routeMap) != 2 {
		t.Errorf("expected 2 routes, got %d", len(routeMap))
	}

	// Check users route
	usersRoute, exists := routeMap.GetRouteInfo("users")
	if !exists {
		t.Error("expected 'users' route to exist")
	}
	if usersRoute.CollectionPath != "/users" {
		t.Errorf("expected collection path '/users', got %s", usersRoute.CollectionPath)
	}
	if usersRoute.ItemPath != "/users/{id}" {
		t.Errorf("expected item path '/users/{id}', got %s", usersRoute.ItemPath)
	}

	// Check posts route
	postsRoute, exists := routeMap.GetRouteInfo("posts")
	if !exists {
		t.Error("expected 'posts' route to exist")
	}
	if postsRoute.CollectionPath != "/posts" {
		t.Errorf("expected collection path '/posts', got %s", postsRoute.CollectionPath)
	}
	if postsRoute.ItemPath != "/posts/{id}" {
		t.Errorf("expected item path '/posts/{id}', got %s", postsRoute.ItemPath)
	}
}

func TestValidateFieldValue(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		value     interface{}
		wantErr   bool
	}{
		{"string valid", types.FieldTypeString, "test", false},
		{"string invalid", types.FieldTypeString, 123, true},
		{"number valid", types.FieldTypeNumber, 123.45, false},
		{"number invalid", types.FieldTypeNumber, "not a number", true},
		{"boolean valid", types.FieldTypeBoolean, true, false},
		{"boolean invalid", types.FieldTypeBoolean, "not a bool", true},
		{"object valid", types.FieldTypeObject, map[string]interface{}{"key": "value"}, false},
		{"object invalid", types.FieldTypeObject, "not an object", true},
		{"array valid", types.FieldTypeArray, []interface{}{1, 2, 3}, false},
		{"array invalid", types.FieldTypeArray, "not an array", true},
		{"null allowed", types.FieldTypeString, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldValue(tt.fieldType, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
