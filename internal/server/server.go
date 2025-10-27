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
	port      int
	mux       *http.ServeMux
	store     storage.Store
	routeMap  schema.RouteMap
	validator *Validator
	server    *http.Server
}

// New creates a new server instance
func New(port int, store storage.Store, routeMap schema.RouteMap, loader *schema.Loader) *Server {
	return &Server{
		port:      port,
		mux:       http.NewServeMux(),
		store:     store,
		routeMap:  routeMap,
		validator: NewValidator(loader),
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
