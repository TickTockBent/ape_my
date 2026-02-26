package storage

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/ticktockbent/ape_my/pkg/types"
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

	// ListQuery retrieves entities with filtering, pagination, and cursor support
	ListQuery(entityType string, opts types.QueryOpts) (*types.QueryResult, error)

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

// ListQuery retrieves entities with filtering, pagination, and cursor support
func (s *InMemoryStore) ListQuery(entityType string, opts types.QueryOpts) (*types.QueryResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data[entityType] == nil {
		return nil, ErrEntityTypeNotFound
	}

	// Collect all entities sorted by ID for deterministic ordering
	allIDs := make([]string, 0, len(s.data[entityType]))
	for id := range s.data[entityType] {
		allIDs = append(allIDs, id)
	}
	sort.Strings(allIDs)

	// Apply filters
	var filtered []map[string]interface{}
	for _, id := range allIDs {
		entity := s.data[entityType][id]
		if matchesFilters(entity, opts.Filters) {
			filtered = append(filtered, copyMap(entity))
		}
	}

	totalCount := len(filtered)

	// Apply cursor-based pagination: skip to after the cursor ID
	if opts.Cursor != "" {
		cursorIndex := -1
		for i, item := range filtered {
			if idVal, ok := item["id"].(string); ok && idVal == opts.Cursor {
				cursorIndex = i
				break
			}
		}
		if cursorIndex >= 0 && cursorIndex+1 < len(filtered) {
			filtered = filtered[cursorIndex+1:]
		} else {
			filtered = nil
		}
	} else if opts.Offset > 0 {
		// Apply offset-based pagination
		if opts.Offset >= len(filtered) {
			filtered = nil
		} else {
			filtered = filtered[opts.Offset:]
		}
	}

	// Apply limit
	var nextCursor string
	if opts.Limit > 0 && len(filtered) > opts.Limit {
		// There are more results; set next cursor to last returned item's ID
		filtered = filtered[:opts.Limit]
		if lastItem := filtered[len(filtered)-1]; lastItem != nil {
			if id, ok := lastItem["id"].(string); ok {
				nextCursor = id
			}
		}
	}

	if filtered == nil {
		filtered = []map[string]interface{}{}
	}

	return &types.QueryResult{
		Items:      filtered,
		TotalCount: totalCount,
		NextCursor: nextCursor,
	}, nil
}

// matchesFilters checks if an entity matches all filter criteria (AND logic)
func matchesFilters(entity map[string]interface{}, filters map[string]string) bool {
	for key, filterValue := range filters {
		entityValue, exists := entity[key]
		if !exists {
			return false
		}

		// Type-coerced comparison
		switch typedValue := entityValue.(type) {
		case string:
			if typedValue != filterValue {
				return false
			}
		case float64:
			filterNum, err := strconv.ParseFloat(filterValue, 64)
			if err != nil || typedValue != filterNum {
				return false
			}
		case bool:
			filterBool, err := strconv.ParseBool(filterValue)
			if err != nil || typedValue != filterBool {
				return false
			}
		default:
			// For non-primitive types, compare string representation
			if fmt.Sprintf("%v", entityValue) != filterValue {
				return false
			}
		}
	}
	return true
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
