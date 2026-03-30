package types

import (
	"testing"
)

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

func TestValidateFieldType(t *testing.T) {
	tests := []struct {
		name         string
		expectedType string
		value        interface{}
		wantErr      bool
	}{
		{"valid string", FieldTypeString, "hello", false},
		{"invalid string", FieldTypeString, 123, true},
		{"valid number", FieldTypeNumber, float64(42), false},
		{"invalid number", FieldTypeNumber, "not a number", true},
		{"valid boolean", FieldTypeBoolean, true, false},
		{"invalid boolean", FieldTypeBoolean, "not a bool", true},
		{"valid object", FieldTypeObject, map[string]interface{}{"key": "value"}, false},
		{"invalid object", FieldTypeObject, "not an object", true},
		{"valid array", FieldTypeArray, []interface{}{1, 2, 3}, false},
		{"invalid array", FieldTypeArray, map[string]interface{}{}, true},
		{"null string", FieldTypeString, nil, false},
		{"null number", FieldTypeNumber, nil, false},
		{"null boolean", FieldTypeBoolean, nil, false},
		{"null object", FieldTypeObject, nil, false},
		{"null array", FieldTypeArray, nil, false},
		{"unknown type", "invalid_type", "value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFieldType(tt.expectedType, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFieldType(%q, %v) error = %v, wantErr %v", tt.expectedType, tt.value, err, tt.wantErr)
			}
		})
	}
}
