package types

// Schema represents the entire schema definition
type Schema struct {
	BasePath string             `json:"basePath,omitempty"`
	Entities map[string]*Entity `json:"entities"`
}

// Entity represents a single entity type (e.g., "users", "posts")
type Entity struct {
	Fields map[string]*Field `json:"fields"`
}

// Field represents a field definition within an entity
type Field struct {
	Type     string `json:"type"`     // string, number, boolean, object, array
	Required bool   `json:"required"` // whether the field is required
}

// FieldType constants for validation
const (
	FieldTypeString  = "string"
	FieldTypeNumber  = "number"
	FieldTypeBoolean = "boolean"
	FieldTypeObject  = "object"
	FieldTypeArray   = "array"
)

// SeedData represents the seed data structure
type SeedData struct {
	Data map[string][]map[string]interface{} `json:"-"`
}

// EntityData represents a collection of entities of a specific type
type EntityData map[string]interface{} // key is entity ID, value is entity data
