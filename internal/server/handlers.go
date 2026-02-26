package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ticktockbent/ape_my/internal/storage"
	"github.com/ticktockbent/ape_my/pkg/types"
)

// handleCollection handles requests to collection endpoints (e.g., /users)
func (s *Server) handleCollection(entityName, collectionPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if this is exactly the collection path (not an item path)
		if r.URL.Path != collectionPath {
			// Let it fall through to item handler or 404
			s.handle404(w, r)
			return
		}

		switch r.Method {
		case http.MethodPost:
			s.handleCreate(entityName, w, r)
		case http.MethodGet:
			s.handleList(entityName, w, r)
		default:
			s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// handleItem handles requests to item endpoints (e.g., /users/123)
func (s *Server) handleItem(entityName, collectionPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		prefix := collectionPath + "/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			s.respondError(w, http.StatusNotFound, "Route not found")
			return
		}

		id := strings.TrimPrefix(r.URL.Path, prefix)
		if id == "" || strings.Contains(id, "/") {
			s.respondError(w, http.StatusNotFound, "Route not found")
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.handleGetOne(entityName, id, w, r)
		case http.MethodPut:
			s.handleUpdate(entityName, id, w, r)
		case http.MethodPatch:
			s.handlePatch(entityName, id, w, r)
		case http.MethodDelete:
			s.handleDelete(entityName, id, w, r)
		default:
			s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// handleCreate handles POST /entities - Create new entity
func (s *Server) handleCreate(entityName string, w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate against schema
	if err := s.validator.ValidateCreate(entityName, data); err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create entity in storage
	id, err := s.store.Create(entityName, data)
	if err != nil {
		if err == storage.ErrEntityTypeNotFound {
			s.respondError(w, http.StatusNotFound, "Entity type not found")
		} else {
			log.Printf("Error creating entity: %v", err)
			s.respondError(w, http.StatusInternalServerError, "Failed to create entity")
		}
		return
	}

	// Get the created entity to return it
	entity, err := s.store.Get(entityName, id)
	if err != nil {
		log.Printf("Error retrieving created entity: %v", err)
		s.respondError(w, http.StatusInternalServerError, "Entity created but failed to retrieve")
		return
	}

	// Return 201 Created with the entity
	s.respondSingle(w, http.StatusCreated, entity)
}

// handleList handles GET /entities - List all entities with optional filtering and pagination
func (s *Server) handleList(entityName string, w http.ResponseWriter, r *http.Request) {
	// Build query options from request query parameters
	opts := s.buildQueryOpts(entityName, r)

	result, err := s.store.ListQuery(entityName, opts)
	if err != nil {
		if err == storage.ErrEntityTypeNotFound {
			s.respondError(w, http.StatusNotFound, "Entity type not found")
		} else {
			log.Printf("Error listing entities: %v", err)
			s.respondError(w, http.StatusInternalServerError, "Failed to list entities")
		}
		return
	}

	// Build response using wrapper if configured, or return raw list
	s.respondList(w, entityName, result)
}

// buildQueryOpts extracts filtering and pagination parameters from the request
func (s *Server) buildQueryOpts(entityName string, r *http.Request) types.QueryOpts {
	opts := types.QueryOpts{
		Filters: make(map[string]string),
	}

	// Get valid field names for this entity to filter query params
	validFields := s.getEntityFieldNames(entityName)

	// Extract filter params — only use params that match entity field names
	for key, values := range r.URL.Query() {
		if validFields[key] && key != "limit" && key != "offset" && key != "cursor" {
			opts.Filters[key] = values[0]
		}
	}

	// Extract pagination params
	if s.schema != nil && s.schema.Pagination != nil {
		pagConfig := s.schema.Pagination

		// Set default limit
		opts.Limit = pagConfig.DefaultLimit
		if opts.Limit == 0 {
			opts.Limit = 20 // fallback default
		}

		// Parse limit from query
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				opts.Limit = limit
			}
		}

		// Cap at max limit
		if pagConfig.MaxLimit > 0 && opts.Limit > pagConfig.MaxLimit {
			opts.Limit = pagConfig.MaxLimit
		}

		// Parse style-specific params
		if pagConfig.Style == "cursor" {
			opts.Cursor = r.URL.Query().Get("cursor")
		} else {
			if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
				if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
					opts.Offset = offset
				}
			}
		}
	}

	return opts
}

// getEntityFieldNames returns a set of valid field names for an entity
func (s *Server) getEntityFieldNames(entityName string) map[string]bool {
	fields := make(map[string]bool)
	if s.schema == nil {
		return fields
	}
	entity, exists := s.schema.Entities[entityName]
	if !exists || entity == nil {
		return fields
	}
	for fieldName := range entity.Fields {
		fields[fieldName] = true
	}
	return fields
}

// handleGetOne handles GET /entities/{id} - Get single entity
func (s *Server) handleGetOne(entityName, id string, w http.ResponseWriter, r *http.Request) {
	entity, err := s.store.Get(entityName, id)
	if err != nil {
		if err == storage.ErrNotFound {
			s.respondError(w, http.StatusNotFound, "Entity not found")
		} else if err == storage.ErrEntityTypeNotFound {
			s.respondError(w, http.StatusNotFound, "Entity type not found")
		} else {
			log.Printf("Error getting entity: %v", err)
			s.respondError(w, http.StatusInternalServerError, "Failed to get entity")
		}
		return
	}

	// Return 200 OK with the entity
	s.respondSingle(w, http.StatusOK, entity)
}

// handleUpdate handles PUT /entities/{id} - Replace entire entity
func (s *Server) handleUpdate(entityName string, id string, w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate against schema
	if err := s.validator.ValidateUpdate(entityName, data); err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update entity in storage
	err = s.store.Update(entityName, id, data)
	if err != nil {
		if err == storage.ErrNotFound {
			s.respondError(w, http.StatusNotFound, "Entity not found")
		} else if err == storage.ErrEntityTypeNotFound {
			s.respondError(w, http.StatusNotFound, "Entity type not found")
		} else {
			log.Printf("Error updating entity: %v", err)
			s.respondError(w, http.StatusInternalServerError, "Failed to update entity")
		}
		return
	}

	// Get the updated entity to return it
	entity, err := s.store.Get(entityName, id)
	if err != nil {
		log.Printf("Error retrieving updated entity: %v", err)
		s.respondError(w, http.StatusInternalServerError, "Entity updated but failed to retrieve")
		return
	}

	// Return 200 OK with the updated entity
	s.respondSingle(w, http.StatusOK, entity)
}

// handlePatch handles PATCH /entities/{id} - Partially update entity
func (s *Server) handlePatch(entityName string, id string, w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate against schema (PATCH doesn't require all required fields)
	if err := s.validator.ValidatePatch(entityName, data); err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Patch entity in storage
	err = s.store.Patch(entityName, id, data)
	if err != nil {
		if err == storage.ErrNotFound {
			s.respondError(w, http.StatusNotFound, "Entity not found")
		} else if err == storage.ErrEntityTypeNotFound {
			s.respondError(w, http.StatusNotFound, "Entity type not found")
		} else {
			log.Printf("Error patching entity: %v", err)
			s.respondError(w, http.StatusInternalServerError, "Failed to patch entity")
		}
		return
	}

	// Get the patched entity to return it
	entity, err := s.store.Get(entityName, id)
	if err != nil {
		log.Printf("Error retrieving patched entity: %v", err)
		s.respondError(w, http.StatusInternalServerError, "Entity patched but failed to retrieve")
		return
	}

	// Return 200 OK with the patched entity
	s.respondSingle(w, http.StatusOK, entity)
}

// handleDelete handles DELETE /entities/{id} - Delete entity
func (s *Server) handleDelete(entityName, id string, w http.ResponseWriter, r *http.Request) {
	err := s.store.Delete(entityName, id)
	if err != nil {
		if err == storage.ErrNotFound {
			s.respondError(w, http.StatusNotFound, "Entity not found")
		} else if err == storage.ErrEntityTypeNotFound {
			s.respondError(w, http.StatusNotFound, "Entity type not found")
		} else {
			log.Printf("Error deleting entity: %v", err)
			s.respondError(w, http.StatusInternalServerError, "Failed to delete entity")
		}
		return
	}

	// Return 204 No Content (successful deletion)
	w.WriteHeader(http.StatusNoContent)
}

// handleCustomRoute handles custom route patterns with path parameter extraction
func (s *Server) handleCustomRoute(route *types.CustomRoute) http.HandlerFunc {
	// Extract parameter names from the original :param path pattern
	paramNames := extractParamNames(route.Path)
	paramSet := make(map[string]bool, len(paramNames))
	for _, name := range paramNames {
		paramSet[name] = true
	}

	return func(w http.ResponseWriter, r *http.Request) {
		filters := make(map[string]string)

		// Add static filters — entries in Filters whose keys are NOT path parameter names
		if route.Filters != nil {
			for key, value := range route.Filters {
				if !paramSet[key] {
					filters[key] = value
				}
			}
		}

		// Extract dynamic path parameters using Go 1.22's PathValue
		for _, paramName := range paramNames {
			paramValue := r.PathValue(paramName)
			if paramValue != "" {
				// Map param name to entity field name using route's Filters config
				filterKey := paramName
				if route.Filters != nil {
					if mappedField, ok := route.Filters[paramName]; ok {
						filterKey = mappedField
					}
				}
				filters[filterKey] = paramValue
			}
		}

		// Query storage with the extracted filters
		opts := types.QueryOpts{Filters: filters}
		result, err := s.store.ListQuery(route.Entity, opts)
		if err != nil {
			if err == storage.ErrEntityTypeNotFound {
				s.respondError(w, http.StatusNotFound, "Entity type not found")
			} else {
				log.Printf("Error querying entities: %v", err)
				s.respondError(w, http.StatusInternalServerError, "Failed to query entities")
			}
			return
		}

		// If filters would match a single entity, return single response
		if len(result.Items) == 1 && hasIDFilter(filters) {
			s.respondSingle(w, http.StatusOK, result.Items[0])
			return
		}

		s.respondList(w, route.Entity, result)
	}
}

// hasIDFilter checks if the filter set targets a specific entity by ID
func hasIDFilter(filters map[string]string) bool {
	_, hasID := filters["id"]
	return hasID
}
