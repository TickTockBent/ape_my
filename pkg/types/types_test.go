package types

import "testing"

// TestFieldTypeConstants verifies field type constants are defined
func TestFieldTypeConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"string type", FieldTypeString},
		{"number type", FieldTypeNumber},
		{"boolean type", FieldTypeBoolean},
		{"object type", FieldTypeObject},
		{"array type", FieldTypeArray},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s is empty", tt.name)
			}
		})
	}
}

// TestSchemaStructure verifies the Schema struct can be instantiated
func TestSchemaStructure(t *testing.T) {
	schema := &Schema{
		Entities: make(map[string]*Entity),
	}

	if schema.Entities == nil {
		t.Error("Schema.Entities should not be nil")
	}
}

// TestEntityStructure verifies the Entity struct can be instantiated
func TestEntityStructure(t *testing.T) {
	entity := &Entity{
		Fields: make(map[string]*Field),
	}

	if entity.Fields == nil {
		t.Error("Entity.Fields should not be nil")
	}
}

// TestFieldStructure verifies the Field struct can be instantiated
func TestFieldStructure(t *testing.T) {
	field := &Field{
		Type:     FieldTypeString,
		Required: true,
	}

	if field.Type != FieldTypeString {
		t.Errorf("expected type %s, got %s", FieldTypeString, field.Type)
	}

	if !field.Required {
		t.Error("expected Required to be true")
	}
}
