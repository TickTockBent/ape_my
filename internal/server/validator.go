package server

import (
	"fmt"

	"github.com/ticktockbent/ape_my/internal/schema"
	"github.com/ticktockbent/ape_my/pkg/types"
)

// Validator validates entity data against schema
type Validator struct {
	loader *schema.Loader
}

// NewValidator creates a new validator
func NewValidator(loader *schema.Loader) *Validator {
	return &Validator{
		loader: loader,
	}
}

// ValidateCreate validates data for creating an entity
func (v *Validator) ValidateCreate(entityName string, data map[string]interface{}) error {
	entity, exists := v.loader.GetEntity(entityName)
	if !exists {
		return fmt.Errorf("entity type %q not found in schema", entityName)
	}

	return v.validateEntityData(entity, data, true)
}

// ValidateUpdate validates data for updating an entity (PUT)
func (v *Validator) ValidateUpdate(entityName string, data map[string]interface{}) error {
	entity, exists := v.loader.GetEntity(entityName)
	if !exists {
		return fmt.Errorf("entity type %q not found in schema", entityName)
	}

	return v.validateEntityData(entity, data, true)
}

// ValidatePatch validates data for patching an entity (PATCH)
func (v *Validator) ValidatePatch(entityName string, data map[string]interface{}) error {
	entity, exists := v.loader.GetEntity(entityName)
	if !exists {
		return fmt.Errorf("entity type %q not found in schema", entityName)
	}

	// For PATCH, required fields are not required (partial update)
	return v.validateEntityData(entity, data, false)
}

// validateEntityData validates entity data against schema
func (v *Validator) validateEntityData(entity *types.Entity, data map[string]interface{}, checkRequired bool) error {
	// Check required fields (except for PATCH)
	if checkRequired {
		for fieldName, field := range entity.Fields {
			// Skip ID field - it's auto-generated or provided
			if fieldName == "id" {
				continue
			}

			if field.Required {
				if _, exists := data[fieldName]; !exists {
					return fmt.Errorf("required field %q is missing", fieldName)
				}
			}
		}
	}

	// Validate field types
	for fieldName, value := range data {
		// Skip ID field
		if fieldName == "id" {
			continue
		}

		// Check if field exists in schema
		field, exists := entity.Fields[fieldName]
		if !exists {
			// Allow extra fields for flexibility
			continue
		}

		// Validate type
		if err := validateFieldType(field.Type, value); err != nil {
			return fmt.Errorf("field %q: %w", fieldName, err)
		}
	}

	return nil
}

// validateFieldType validates that a value matches the expected type
func validateFieldType(expectedType string, value interface{}) error {
	if value == nil {
		return nil // Allow null values
	}

	switch expectedType {
	case types.FieldTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case types.FieldTypeNumber:
		// JSON numbers are float64
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
	default:
		return fmt.Errorf("unknown field type: %s", expectedType)
	}

	return nil
}
