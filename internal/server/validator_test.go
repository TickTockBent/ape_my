package server

import (
	"testing"
)

func TestValidateCreate(t *testing.T) {
	loader := setupTestSchema(t)
	validator := NewValidator(loader)

	tests := []struct {
		name       string
		entityName string
		data       map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "valid user data",
			entityName: "users",
			data: map[string]interface{}{
				"name":  "Alice",
				"email": "alice@example.com",
			},
			wantErr: false,
		},
		{
			name:       "missing required field",
			entityName: "users",
			data: map[string]interface{}{
				"email": "alice@example.com",
			},
			wantErr: true,
		},
		{
			name:       "wrong field type",
			entityName: "users",
			data: map[string]interface{}{
				"name": "Alice",
				"age":  "not a number",
			},
			wantErr: true,
		},
		{
			name:       "extra fields allowed",
			entityName: "users",
			data: map[string]interface{}{
				"name":     "Alice",
				"extraKey": "extraValue",
			},
			wantErr: false,
		},
		{
			name:       "null values allowed",
			entityName: "users",
			data: map[string]interface{}{
				"name":  "Alice",
				"email": nil,
			},
			wantErr: false,
		},
		{
			name:       "unknown entity type",
			entityName: "unknown",
			data: map[string]interface{}{
				"name": "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreate(tt.entityName, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUpdate(t *testing.T) {
	loader := setupTestSchema(t)
	validator := NewValidator(loader)

	tests := []struct {
		name       string
		entityName string
		data       map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "valid update",
			entityName: "users",
			data: map[string]interface{}{
				"name":  "Alice Updated",
				"email": "new@example.com",
			},
			wantErr: false,
		},
		{
			name:       "missing required field",
			entityName: "users",
			data: map[string]interface{}{
				"email": "alice@example.com",
			},
			wantErr: true,
		},
		{
			name:       "wrong field type",
			entityName: "users",
			data: map[string]interface{}{
				"name": "Alice",
				"age":  "not a number",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUpdate(tt.entityName, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePatch(t *testing.T) {
	loader := setupTestSchema(t)
	validator := NewValidator(loader)

	tests := []struct {
		name       string
		entityName string
		data       map[string]interface{}
		wantErr    bool
	}{
		{
			name:       "valid partial update",
			entityName: "users",
			data: map[string]interface{}{
				"age": float64(30),
			},
			wantErr: false,
		},
		{
			name:       "partial update without required fields (allowed for PATCH)",
			entityName: "users",
			data: map[string]interface{}{
				"email": "newemail@example.com",
			},
			wantErr: false,
		},
		{
			name:       "wrong field type",
			entityName: "users",
			data: map[string]interface{}{
				"age": "not a number",
			},
			wantErr: true,
		},
		{
			name:       "empty patch (valid)",
			entityName: "users",
			data:       map[string]interface{}{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePatch(tt.entityName, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFieldType(t *testing.T) {
	tests := []struct {
		name         string
		expectedType string
		value        interface{}
		wantErr      bool
	}{
		// String tests
		{"valid string", "string", "hello", false},
		{"invalid string", "string", 123, true},

		// Number tests
		{"valid number", "number", float64(42), false},
		{"invalid number", "number", "not a number", true},

		// Boolean tests
		{"valid boolean", "boolean", true, false},
		{"invalid boolean", "boolean", "not a bool", true},

		// Object tests
		{"valid object", "object", map[string]interface{}{"key": "value"}, false},
		{"invalid object", "object", []interface{}{}, true},

		// Array tests
		{"valid array", "array", []interface{}{1, 2, 3}, false},
		{"invalid array", "array", map[string]interface{}{}, true},

		// Null values
		{"null string", "string", nil, false},
		{"null number", "number", nil, false},
		{"null boolean", "boolean", nil, false},

		// Unknown type
		{"unknown type", "invalid_type", "value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFieldType(tt.expectedType, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFieldType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	loader := setupTestSchema(t)
	validator := NewValidator(loader)

	if validator == nil {
		t.Fatal("expected validator to not be nil")
	}

	if validator.loader == nil {
		t.Error("expected validator.loader to not be nil")
	}
}
