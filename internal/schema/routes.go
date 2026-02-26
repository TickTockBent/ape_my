package schema

import (
	"fmt"
	"strings"
)

// RouteInfo holds information about a generated route
type RouteInfo struct {
	EntityName     string
	CollectionPath string // e.g., "/users"
	ItemPath       string // e.g., "/users/{id}"
}

// RouteMap maps entity names to their route information
type RouteMap map[string]*RouteInfo

// NormalizeBasePath normalizes a base path by ensuring it has a leading slash
// and no trailing slash. Empty strings are returned as-is (no prefix).
func NormalizeBasePath(basePath string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" {
		return ""
	}
	// Ensure leading slash
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	// Remove trailing slash
	basePath = strings.TrimRight(basePath, "/")
	return basePath
}

// BuildRouteMap creates a route map from the loaded schema
func (l *Loader) BuildRouteMap() (RouteMap, error) {
	if l.schema == nil {
		return nil, fmt.Errorf("no schema loaded")
	}

	routeMap := make(RouteMap)
	prefix := NormalizeBasePath(l.schema.BasePath)

	for entityName := range l.schema.Entities {
		routeInfo := &RouteInfo{
			EntityName:     entityName,
			CollectionPath: fmt.Sprintf("%s/%s", prefix, entityName),
			ItemPath:       fmt.Sprintf("%s/%s/{id}", prefix, entityName),
		}
		routeMap[entityName] = routeInfo
	}

	return routeMap, nil
}

// GetRoutes returns all route information as a slice
func (rm RouteMap) GetRoutes() []*RouteInfo {
	routes := make([]*RouteInfo, 0, len(rm))
	for _, route := range rm {
		routes = append(routes, route)
	}
	return routes
}

// GetRouteInfo returns route information for a specific entity
func (rm RouteMap) GetRouteInfo(entityName string) (*RouteInfo, bool) {
	route, exists := rm[entityName]
	return route, exists
}
