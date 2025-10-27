package server

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// respondJSON writes a JSON response
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// If we can't encode the response, log it
			// but don't try to send another response
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

// respondError writes a JSON error response
func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, ErrorResponse{Error: message})
}
