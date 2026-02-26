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
	"github.com/ticktockbent/ape_my/pkg/types"
)

// Server represents the HTTP server
type Server struct {
	port      int
	mux       *http.ServeMux
	store     storage.Store
	routeMap  schema.RouteMap
	validator *Validator
	schema    *types.Schema
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
		schema:    loader.GetSchema(),
	}
}

// RegisterRoutes dynamically registers routes based on the schema
func (s *Server) RegisterRoutes() {
	// Register routes for each entity
	for _, route := range s.routeMap.GetRoutes() {
		entityName := route.EntityName
		collectionPath := route.CollectionPath

		// Collection routes: POST /entities, GET /entities
		s.mux.HandleFunc(collectionPath, s.withMiddleware(s.handleCollection(entityName, collectionPath)))

		// Item routes: GET /entities/123, PUT /entities/123, PATCH /entities/123, DELETE /entities/123
		// Use collection path with trailing slash to catch all sub-paths
		itemPattern := collectionPath + "/"
		s.mux.HandleFunc(itemPattern, s.withMiddleware(s.handleItem(entityName, collectionPath)))

		log.Printf("Registered routes: %s and %s", collectionPath, itemPattern)
	}

	// Register custom routes if configured
	if s.schema != nil && s.schema.Routes != nil {
		prefix := schema.NormalizeBasePath(s.schema.BasePath)
		for _, route := range s.schema.Routes {
			customRoute := route // capture loop variable
			// Convert :param syntax to Go 1.22 {param} syntax for mux registration
			routePath := prefix + convertPathParams(customRoute.Path)
			// Use method prefix for Go 1.22 mux to avoid conflicts with CRUD routes
			muxPattern := strings.ToUpper(customRoute.Method) + " " + routePath
			s.mux.HandleFunc(muxPattern, s.withMiddleware(s.handleCustomRoute(customRoute)))
			log.Printf("Registered custom route: %s %s -> %s", customRoute.Method, routePath, customRoute.Entity)
		}
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

// protectedHeaders are headers that custom response headers cannot override
var protectedHeaders = map[string]bool{
	"content-type":   true,
	"content-length": true,
}

// withMiddleware wraps a handler with logging, auth, and content-type checking
func (s *Server) withMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Logging middleware
		start := time.Now()
		log.Printf("%s %s", r.Method, r.URL.Path)

		// Auth middleware — validate Bearer token if configured
		if s.schema != nil && s.schema.Auth != nil {
			authHeader := r.Header.Get("Authorization")
			expectedToken := "Bearer " + s.schema.Auth.Token
			if authHeader != expectedToken {
				w.Header().Set("Content-Type", "application/json")
				s.respondError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
		}

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

		// Set custom response headers if configured
		if s.schema != nil && s.schema.ResponseHeaders != nil {
			for key, value := range s.schema.ResponseHeaders {
				if !protectedHeaders[strings.ToLower(key)] {
					w.Header().Set(key, value)
				}
			}
		}

		// Call the handler
		next(w, r)

		// Log completion
		duration := time.Since(start)
		log.Printf("%s %s completed in %v", r.Method, r.URL.Path, duration)
	}
}

// convertPathParams converts :param syntax to Go 1.22 {param} syntax
func convertPathParams(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

// extractParamNames returns the parameter names from a route path using :param syntax
func extractParamNames(path string) []string {
	var names []string
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, ":") {
			names = append(names, part[1:])
		}
	}
	return names
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
