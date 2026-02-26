package types

// Schema represents the entire schema definition
type Schema struct {
	BasePath        string                `json:"basePath,omitempty"`
	Entities        map[string]*Entity    `json:"entities"`
	ResponseHeaders map[string]string     `json:"responseHeaders,omitempty"`
	Auth            *AuthConfig           `json:"auth,omitempty"`
	ResponseWrapper *ResponseWrapperConfig `json:"responseWrapper,omitempty"`
	Pagination      *PaginationConfig     `json:"pagination,omitempty"`
	Routes          []*CustomRoute        `json:"routes,omitempty"`
}

// AuthConfig defines bearer token authentication settings
type AuthConfig struct {
	Token string `json:"token"`
}

// ResponseWrapperConfig defines response envelope templates
type ResponseWrapperConfig struct {
	Single interface{} `json:"single,omitempty"`
	List   interface{} `json:"list,omitempty"`
}

// PaginationConfig defines pagination behavior
type PaginationConfig struct {
	Style        string `json:"style"`                  // "cursor" or "offset"
	DefaultLimit int    `json:"defaultLimit,omitempty"`
	MaxLimit     int    `json:"maxLimit,omitempty"`
}

// CustomRoute defines a custom route pattern
type CustomRoute struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Entity  string            `json:"entity"`
	Filters map[string]string `json:"filters,omitempty"`
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

// QueryOpts defines options for querying entities from storage
type QueryOpts struct {
	Filters map[string]string
	Limit   int
	Offset  int
	Cursor  string
}

// QueryResult holds the results of a storage query
type QueryResult struct {
	Items      []map[string]interface{}
	TotalCount int
	NextCursor string
}

// SeedData represents the seed data structure
type SeedData struct {
	Data map[string][]map[string]interface{} `json:"-"`
}

// EntityData represents a collection of entities of a specific type
type EntityData map[string]interface{} // key is entity ID, value is entity data
