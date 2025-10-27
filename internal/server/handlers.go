package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/ticktockbent/ape_my/internal/storage"
)

// handleCollection handles requests to collection endpoints (e.g., /users)
func (s *Server) handleCollection(entityName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if this is exactly the collection path (not an item path)
		expectedPath := fmt.Sprintf("/%s", entityName)
		if r.URL.Path != expectedPath {
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
func (s *Server) handleItem(entityName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		prefix := fmt.Sprintf("/%s/", entityName)
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
	s.respondJSON(w, http.StatusCreated, entity)
}

// handleList handles GET /entities - List all entities
func (s *Server) handleList(entityName string, w http.ResponseWriter, r *http.Request) {
	entities, err := s.store.List(entityName)
	if err != nil {
		if err == storage.ErrEntityTypeNotFound {
			s.respondError(w, http.StatusNotFound, "Entity type not found")
		} else {
			log.Printf("Error listing entities: %v", err)
			s.respondError(w, http.StatusInternalServerError, "Failed to list entities")
		}
		return
	}

	// Return 200 OK with the list
	s.respondJSON(w, http.StatusOK, entities)
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
	s.respondJSON(w, http.StatusOK, entity)
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
	s.respondJSON(w, http.StatusOK, entity)
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
	s.respondJSON(w, http.StatusOK, entity)
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
