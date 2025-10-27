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

// Initialize sets up storage for entity types
func (s *InMemoryStore) Initialize(entityTypes []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entityType := range entityTypes {
		if s.data[entityType] == nil {
			s.data[entityType] = make(map[string]map[string]interface{})
		}
		if _, exists := s.counter[entityType]; !exists {
			s.counter[entityType] = 0
		}
	}

	return nil
}

// Create adds a new entity and returns its ID
func (s *InMemoryStore) Create(entityType string, data map[string]interface{}) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if entity type exists
	if s.data[entityType] == nil {
		return "", ErrEntityTypeNotFound
	}

	// Generate ID if not provided
	var id string
	if providedID, exists := data["id"]; exists && providedID != nil {
		id = providedID.(string)
	} else {
		s.counter[entityType]++
		id = formatID(s.counter[entityType])
		data["id"] = id
	}

	// Store the entity
	s.data[entityType][id] = copyMap(data)

	return id, nil
}

// Get retrieves a single entity by ID
func (s *InMemoryStore) Get(entityType, id string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if entity type exists
	if s.data[entityType] == nil {
		return nil, ErrEntityTypeNotFound
	}

	// Get the entity
	entity, exists := s.data[entityType][id]
	if !exists {
		return nil, ErrNotFound
	}

	return copyMap(entity), nil
}

// List retrieves all entities of a given type
func (s *InMemoryStore) List(entityType string) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if entity type exists
	if s.data[entityType] == nil {
		return nil, ErrEntityTypeNotFound
	}

	// Collect all entities
	entities := make([]map[string]interface{}, 0, len(s.data[entityType]))
	for _, entity := range s.data[entityType] {
		entities = append(entities, copyMap(entity))
	}

	return entities, nil
}

// Update replaces an entire entity
func (s *InMemoryStore) Update(entityType, id string, data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if entity type exists
	if s.data[entityType] == nil {
		return ErrEntityTypeNotFound
	}

	// Check if entity exists
	if _, exists := s.data[entityType][id]; !exists {
		return ErrNotFound
	}

	// Ensure ID is preserved
	data["id"] = id

	// Replace the entity
	s.data[entityType][id] = copyMap(data)

	return nil
}

// Patch partially updates an entity
func (s *InMemoryStore) Patch(entityType, id string, data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if entity type exists
	if s.data[entityType] == nil {
		return ErrEntityTypeNotFound
	}

	// Check if entity exists
	entity, exists := s.data[entityType][id]
	if !exists {
		return ErrNotFound
	}

	// Merge the data
	for key, value := range data {
		// Don't allow changing the ID
		if key != "id" {
			entity[key] = value
		}
	}

	return nil
}

// Delete removes an entity
func (s *InMemoryStore) Delete(entityType, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if entity type exists
	if s.data[entityType] == nil {
		return ErrEntityTypeNotFound
	}

	// Check if entity exists
	if _, exists := s.data[entityType][id]; !exists {
		return ErrNotFound
	}

	// Delete the entity
	delete(s.data[entityType], id)

	return nil
}

// Seed loads initial data into storage
func (s *InMemoryStore) Seed(entityType string, entities []map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if entity type exists
	if s.data[entityType] == nil {
		return ErrEntityTypeNotFound
	}

	// Load each entity
	for _, entity := range entities {
		// Get the ID
		idValue, exists := entity["id"]
		if !exists {
			// Skip entities without IDs in seed data
			continue
		}

		id, ok := idValue.(string)
		if !ok {
			// Skip if ID is not a string
			continue
		}

		// Store the entity
		s.data[entityType][id] = copyMap(entity)

		// Update counter to ensure we don't generate duplicate IDs
		if numID := parseIDNumber(id); numID > s.counter[entityType] {
			s.counter[entityType] = numID
		}
	}

	return nil
}

// Helper functions

// copyMap creates a deep copy of a map
func copyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

// formatID formats an integer counter into a string ID
func formatID(counter int) string {
	// Simple numeric string conversion
	if counter < 10 {
		return string(rune('0' + counter))
	}
	// For larger numbers, use string conversion
	var result []byte
	for counter > 0 {
		result = append([]byte{byte('0' + (counter % 10))}, result...)
		counter /= 10
	}
	return string(result)
}

// parseIDNumber attempts to parse a numeric ID from a string
func parseIDNumber(id string) int {
	// Simple parsing for numeric IDs (e.g., "1", "2", "3")
	if len(id) == 1 && id[0] >= '0' && id[0] <= '9' {
		return int(id[0] - '0')
	}
	// For multi-digit or non-numeric IDs, return 0
	var num int
	for _, ch := range id {
		if ch >= '0' && ch <= '9' {
			num = num*10 + int(ch-'0')
		} else {
			return 0
		}
	}
	return num
}
