package tests

import (
	"testing"

	"github.com/ticktockbent/ape_my/internal/schema"
)

// TestExampleSchemas tests that our example schema files are valid
func TestExampleSchemas(t *testing.T) {
	tests := []struct {
		name       string
		schemaPath string
		seedPath   string
	}{
		{
			name:       "todos example",
			schemaPath: "../examples/todos_schema.json",
			seedPath:   "../examples/todos_seed.json",
		},
		{
			name:       "users example",
			schemaPath: "../examples/users_schema.json",
			seedPath:   "../examples/users_seed.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load schema
			loader := schema.NewLoader()
			err := loader.LoadFromFile(tt.schemaPath)
			if err != nil {
				t.Fatalf("failed to load schema: %v", err)
			}

			// Verify we got entities
			entityNames := loader.GetEntityNames()
			if len(entityNames) == 0 {
				t.Error("expected at least one entity")
			}

			t.Logf("Loaded entities: %v", entityNames)

			// Build route map
			routeMap, err := loader.BuildRouteMap()
			if err != nil {
				t.Fatalf("failed to build route map: %v", err)
			}

			// Verify routes
			routes := routeMap.GetRoutes()
			if len(routes) == 0 {
				t.Error("expected at least one route")
			}

			for _, route := range routes {
				t.Logf("Route: %s -> %s, %s", route.EntityName, route.CollectionPath, route.ItemPath)
			}

			// Load seed data
			seedData, err := schema.LoadSeedData(tt.seedPath)
			if err != nil {
				t.Fatalf("failed to load seed data: %v", err)
			}

			// Validate seed data against schema
			err = loader.ValidateSeedData(seedData)
			if err != nil {
				t.Fatalf("seed data validation failed: %v", err)
			}

			t.Logf("Successfully validated seed data for %d entities", len(seedData))
		})
	}
}

// TestSchemaValidation tests various schema validation scenarios
func TestSchemaValidation(t *testing.T) {
	// This test loads the actual example schema and validates it
	loader := schema.NewLoader()
	err := loader.LoadFromFile("../examples/todos_schema.json")
	if err != nil {
		t.Fatalf("expected valid schema to load: %v", err)
	}

	// Test getting a specific entity
	entity, exists := loader.GetEntity("todos")
	if !exists {
		t.Fatal("expected 'todos' entity to exist")
	}

	// Verify it has the expected fields
	expectedFields := []string{"id", "task", "completed", "priority"}
	for _, fieldName := range expectedFields {
		if _, exists := entity.Fields[fieldName]; !exists {
			t.Errorf("expected field %q to exist", fieldName)
		}
	}
}
