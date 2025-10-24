package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ticktockbent/ape_my/internal/schema"
	"github.com/ticktockbent/ape_my/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	port     int
	mux      *http.ServeMux
	store    storage.Store
	routeMap schema.RouteMap
	server   *http.Server
}

// New creates a new server instance
func New(port int, store storage.Store, routeMap schema.RouteMap) *Server {
	return &Server{
		port:     port,
		mux:      http.NewServeMux(),
		store:    store,
		routeMap: routeMap,
	}
}

// RegisterRoutes dynamically registers routes based on the schema
func (s *Server) RegisterRoutes() {
	// Register routes for each entity
	for _, route := range s.routeMap.GetRoutes() {
		entityName := route.EntityName
		collectionPath := route.CollectionPath

		// Collection routes: POST /entities, GET /entities
		s.mux.HandleFunc(collectionPath, s.withMiddleware(s.handleCollection(entityName)))

		// Item routes: GET /entities/123, PUT /entities/123, PATCH /entities/123, DELETE /entities/123
		// Use collection path with trailing slash to catch all sub-paths
		itemPattern := collectionPath + "/"
		s.mux.HandleFunc(itemPattern, s.withMiddleware(s.handleItem(entityName)))

		log.Printf("Registered routes: %s and %s", collectionPath, itemPattern)
	}

	// Handle 404 for all other routes
	s.mux.HandleFunc("/", s.withMiddleware(s.handle404))
}

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
			// Will be implemented in Phase 5
			s.respondJSON(w, http.StatusNotImplemented, map[string]string{
				"error": "POST not yet implemented",
			})
		case http.MethodGet:
			// Will be implemented in Phase 5
			s.respondJSON(w, http.StatusNotImplemented, map[string]string{
				"error": "GET not yet implemented",
			})
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
			// Will be implemented in Phase 5
			s.respondJSON(w, http.StatusNotImplemented, map[string]string{
				"error": "GET not yet implemented",
				"id":    id,
			})
		case http.MethodPut:
			// Will be implemented in Phase 5
			s.respondJSON(w, http.StatusNotImplemented, map[string]string{
				"error": "PUT not yet implemented",
				"id":    id,
			})
		case http.MethodPatch:
			// Will be implemented in Phase 5
			s.respondJSON(w, http.StatusNotImplemented, map[string]string{
				"error": "PATCH not yet implemented",
				"id":    id,
			})
		case http.MethodDelete:
			// Will be implemented in Phase 5
			s.respondJSON(w, http.StatusNotImplemented, map[string]string{
				"error": "DELETE not yet implemented",
				"id":    id,
			})
		default:
			s.respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// handle404 handles unknown routes
func (s *Server) handle404(w http.ResponseWriter, r *http.Request) {
	// Don't handle if it matches a registered route pattern
	for _, route := range s.routeMap.GetRoutes() {
		if r.URL.Path == route.CollectionPath {
			return
		}
		prefix := fmt.Sprintf("%s/", route.CollectionPath)
		if strings.HasPrefix(r.URL.Path, prefix) {
			return
		}
	}

	s.respondError(w, http.StatusNotFound, "Route not found")
}

// withMiddleware wraps a handler with logging and content-type checking
func (s *Server) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Logging middleware
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)

		// Content-Type validation for POST, PUT, PATCH
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") {
				s.respondError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}
		}

		// Set JSON response header
		w.Header().Set("Content-Type", "application/json")

		// Call the handler
		next(w, r)

		// Log completion
		duration := time.Since(start)
		log.Printf("%s %s completed in %v", r.Method, r.URL.Path, duration)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on http://localhost:%d", s.port)
	log.Printf("Press Ctrl+C to stop")

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
