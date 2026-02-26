package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ticktockbent/ape_my/pkg/types"
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

// respondSingle writes a single-entity response, applying wrapper if configured
func (s *Server) respondSingle(w http.ResponseWriter, status int, entity map[string]interface{}) {
	if s.schema != nil && s.schema.ResponseWrapper != nil && s.schema.ResponseWrapper.Single != nil {
		wrapped := applyTemplate(s.schema.ResponseWrapper.Single, map[string]interface{}{
			"$entity": entity,
		})
		s.respondJSON(w, status, wrapped)
		return
	}
	s.respondJSON(w, status, entity)
}

// respondList writes a list response with optional wrapping and pagination metadata
func (s *Server) respondList(w http.ResponseWriter, entityName string, result *types.QueryResult) {
	// Build metadata map for template substitution
	metadata := map[string]interface{}{
		"$entities":     result.Items,
		"$count":        len(result.Items),
		"$result_count": len(result.Items),
	}
	if result.NextCursor != "" {
		metadata["$next_token"] = result.NextCursor
	}

	if s.schema != nil && s.schema.ResponseWrapper != nil && s.schema.ResponseWrapper.List != nil {
		wrapped := applyTemplate(s.schema.ResponseWrapper.List, metadata)
		s.respondJSON(w, http.StatusOK, wrapped)
		return
	}

	// No wrapper configured — check if pagination metadata should be included
	if s.schema != nil && s.schema.Pagination != nil {
		meta := map[string]interface{}{
			"result_count": len(result.Items),
		}
		if s.schema.Pagination.Style == "cursor" && result.NextCursor != "" {
			meta["next_token"] = result.NextCursor
		}

		// Only include meta wrapper if there's meaningful pagination info
		if result.NextCursor != "" || result.TotalCount > len(result.Items) {
			response := map[string]interface{}{
				"data": result.Items,
				"meta": meta,
			}
			s.respondJSON(w, http.StatusOK, response)
			return
		}
	}

	s.respondJSON(w, http.StatusOK, result.Items)
}

// applyTemplate recursively processes a template structure, substituting variables
func applyTemplate(template interface{}, vars map[string]interface{}) interface{} {
	switch tmpl := template.(type) {
	case string:
		// Check if the entire string is a single variable reference
		if val, ok := vars[tmpl]; ok {
			return val
		}
		// Check for inline variable substitution in strings
		result := tmpl
		for key, val := range vars {
			if strings.Contains(result, key) {
				result = strings.ReplaceAll(result, key, fmt.Sprintf("%v", val))
			}
		}
		return result
	case map[string]interface{}:
		out := make(map[string]interface{}, len(tmpl))
		for k, v := range tmpl {
			out[k] = applyTemplate(v, vars)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(tmpl))
		for i, v := range tmpl {
			out[i] = applyTemplate(v, vars)
		}
		return out
	case float64, bool, nil:
		return tmpl
	default:
		return tmpl
	}
}
