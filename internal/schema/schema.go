package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ticktockbent/ape_my/pkg/types"
)

var (
	// ErrEmptySchema is returned when the schema has no entities
	ErrEmptySchema = errors.New("schema contains no entities")

	// ErrInvalidFieldType is returned when a field has an invalid type
	ErrInvalidFieldType = errors.New("invalid field type")

	// ErrNoFields is returned when an entity has no fields
	ErrNoFields = errors.New("entity has no fields")

	// ErrMissingIDField is returned when an entity doesn't have an id field
	ErrMissingIDField = errors.New("entity must have an 'id' field")
)

// Loader handles loading and validating schemas
type Loader struct {
	schema *types.Schema
}

// NewLoader creates a new schema loader
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFromFile loads a schema from a JSON file
func (l *Loader) LoadFromFile(filepath string) error {
	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Parse JSON
	var schema types.Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	l.schema = &schema

	// Validate schema
	if err := l.Validate(); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}

// Validate validates the loaded schema
func (l *Loader) Validate() error {
	if l.schema == nil {
		return errors.New("no schema loaded")
	}

	// Check if schema has entities
	if len(l.schema.Entities) == 0 {
		return ErrEmptySchema
	}

	// Validate each entity
	for entityName, entity := range l.schema.Entities {
		if err := l.validateEntity(entityName, entity); err != nil {
			return fmt.Errorf("entity %q: %w", entityName, err)
		}
	}

	return nil
}

// validateEntity validates a single entity
func (l *Loader) validateEntity(name string, entity *types.Entity) error {
	if entity == nil {
		return errors.New("entity is nil")
	}

	// Check if entity has fields
	if len(entity.Fields) == 0 {
		return ErrNoFields
	}

	// Check for id field
	idField, hasID := entity.Fields["id"]
	if !hasID {
		return ErrMissingIDField
	}

	// Validate id field is string type
	if idField.Type != types.FieldTypeString {
		return fmt.Errorf("id field must be of type 'string', got '%s'", idField.Type)
	}

	// Validate each field
	for fieldName, field := range entity.Fields {
		if err := l.validateField(fieldName, field); err != nil {
			return fmt.Errorf("field %q: %w", fieldName, err)
		}
	}

	return nil
}

// validateField validates a single field
func (l *Loader) validateField(name string, field *types.Field) error {
	if field == nil {
		return errors.New("field is nil")
	}

	// Validate field type
	validTypes := map[string]bool{
		types.FieldTypeString:  true,
		types.FieldTypeNumber:  true,
		types.FieldTypeBoolean: true,
		types.FieldTypeObject:  true,
		types.FieldTypeArray:   true,
	}

	if !validTypes[field.Type] {
		return fmt.Errorf("%w: %s (must be one of: string, number, boolean, object, array)", ErrInvalidFieldType, field.Type)
	}

	return nil
}

// GetSchema returns the loaded schema
func (l *Loader) GetSchema() *types.Schema {
	return l.schema
}

// GetEntityNames returns a list of all entity names in the schema
func (l *Loader) GetEntityNames() []string {
	if l.schema == nil {
		return nil
	}

	names := make([]string, 0, len(l.schema.Entities))
	for name := range l.schema.Entities {
		names = append(names, name)
	}

	return names
}

// GetEntity returns a specific entity by name
func (l *Loader) GetEntity(name string) (*types.Entity, bool) {
	if l.schema == nil {
		return nil, false
	}

	entity, exists := l.schema.Entities[name]
	return entity, exists
}

// LoadSeedData loads seed data from a JSON file
func LoadSeedData(filepath string) (map[string][]map[string]interface{}, error) {
	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed file: %w", err)
	}

	// Parse JSON
	var seedData map[string][]map[string]interface{}
	if err := json.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("failed to parse seed JSON: %w", err)
	}

	return seedData, nil
}

// ValidateSeedData validates that seed data matches the schema
func (l *Loader) ValidateSeedData(seedData map[string][]map[string]interface{}) error {
	if l.schema == nil {
		return errors.New("no schema loaded")
	}

	// Check each entity in seed data
	for entityName, entities := range seedData {
		// Check if entity exists in schema
		entity, exists := l.schema.Entities[entityName]
		if !exists {
			return fmt.Errorf("seed data contains unknown entity: %s", entityName)
		}

		// Validate each entity instance
		for i, entityData := range entities {
			if err := l.validateEntityData(entityName, entity, entityData); err != nil {
				return fmt.Errorf("seed data for %s[%d]: %w", entityName, i, err)
			}
		}
	}

	return nil
}

// validateEntityData validates a single entity instance against the schema
func (l *Loader) validateEntityData(entityName string, entity *types.Entity, data map[string]interface{}) error {
	// Check required fields
	for fieldName, field := range entity.Fields {
		if field.Required {
			if _, exists := data[fieldName]; !exists {
				return fmt.Errorf("required field %q is missing", fieldName)
			}
		}
	}

	// Validate field types (basic validation)
	for fieldName, value := range data {
		// Check if field exists in schema
		field, exists := entity.Fields[fieldName]
		if !exists {
			// Allow extra fields in seed data (flexibility)
			continue
		}

		// Basic type checking
		if err := validateFieldValue(field.Type, value); err != nil {
			return fmt.Errorf("field %q: %w", fieldName, err)
		}
	}

	return nil
}

// validateFieldValue performs basic type validation on a field value
func validateFieldValue(fieldType string, value interface{}) error {
	if value == nil {
		return nil // Allow null values
	}

	switch fieldType {
	case types.FieldTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case types.FieldTypeNumber:
		// JSON numbers can be float64
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected number, got %T", value)
		}
	case types.FieldTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case types.FieldTypeObject:
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	case types.FieldTypeArray:
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	}

	return nil
}
