package schema

import (
	"fmt"
)

// RouteInfo holds information about a generated route
type RouteInfo struct {
	EntityName     string
	CollectionPath string // e.g., "/users"
	ItemPath       string // e.g., "/users/{id}"
}

// RouteMap maps entity names to their route information
type RouteMap map[string]*RouteInfo

// BuildRouteMap creates a route map from the loaded schema
func (l *Loader) BuildRouteMap() (RouteMap, error) {
	if l.schema == nil {
		return nil, fmt.Errorf("no schema loaded")
	}

	routeMap := make(RouteMap)

	for entityName := range l.schema.Entities {
		routeInfo := &RouteInfo{
			EntityName:     entityName,
			CollectionPath: fmt.Sprintf("/%s", entityName),
			ItemPath:       fmt.Sprintf("/%s/{id}", entityName),
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
