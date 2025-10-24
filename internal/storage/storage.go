package storage

import (
	"errors"
	"sync"
)

var (
	// ErrNotFound is returned when an entity is not found
	ErrNotFound = errors.New("entity not found")

	// ErrEntityTypeNotFound is returned when an entity type doesn't exist in schema
	ErrEntityTypeNotFound = errors.New("entity type not found")
)

// Store defines the interface for data storage operations
type Store interface {
	// Create adds a new entity and returns its ID
	Create(entityType string, data map[string]interface{}) (string, error)

	// Get retrieves a single entity by ID
	Get(entityType string, id string) (map[string]interface{}, error)

	// List retrieves all entities of a given type
	List(entityType string) ([]map[string]interface{}, error)

	// Update replaces an entire entity
	Update(entityType string, id string, data map[string]interface{}) error

	// Patch partially updates an entity
	Patch(entityType string, id string, data map[string]interface{}) error

	// Delete removes an entity
	Delete(entityType string, id string) error

	// Initialize sets up storage for entity types
	Initialize(entityTypes []string) error

	// Seed loads initial data into storage
	Seed(entityType string, entities []map[string]interface{}) error
}

// InMemoryStore implements Store using in-memory storage
type InMemoryStore struct {
	mu      sync.RWMutex
	data    map[string]map[string]map[string]interface{} // entityType -> id -> entity
	counter map[string]int                               // entityType -> counter for ID generation
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data:    make(map[string]map[string]map[string]interface{}),
		counter: make(map[string]int),
	}
}
