package schema

import (
	"testing"

	"github.com/ticktockbent/ape_my/pkg/types"
)

func TestNormalizeBasePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"just slash", "/", ""},
		{"with leading slash", "/2", "/2"},
		{"without leading slash", "2", "/2"},
		{"with trailing slash", "/2/", "/2"},
		{"both slashes", "/api/v1/", "/api/v1"},
		{"no slashes", "api/v1", "/api/v1"},
		{"whitespace", "  /2  ", "/2"},
		{"multiple trailing slashes", "/2///", "/2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeBasePath(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeBasePath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildRouteMapWithBasePath(t *testing.T) {
	tests := []struct {
		name               string
		basePath           string
		entityName         string
		wantCollectionPath string
		wantItemPath       string
	}{
		{
			name:               "no base path",
			basePath:           "",
			entityName:         "users",
			wantCollectionPath: "/users",
			wantItemPath:       "/users/{id}",
		},
		{
			name:               "versioned base path",
			basePath:           "/2",
			entityName:         "tweets",
			wantCollectionPath: "/2/tweets",
			wantItemPath:       "/2/tweets/{id}",
		},
		{
			name:               "api prefix",
			basePath:           "/api/v1",
			entityName:         "users",
			wantCollectionPath: "/api/v1/users",
			wantItemPath:       "/api/v1/users/{id}",
		},
		{
			name:               "base path without leading slash",
			basePath:           "2",
			entityName:         "tweets",
			wantCollectionPath: "/2/tweets",
			wantItemPath:       "/2/tweets/{id}",
		},
		{
			name:               "base path with trailing slash",
			basePath:           "/2/",
			entityName:         "tweets",
			wantCollectionPath: "/2/tweets",
			wantItemPath:       "/2/tweets/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &Loader{
				schema: &types.Schema{
					BasePath: tt.basePath,
					Entities: map[string]*types.Entity{
						tt.entityName: {
							Fields: map[string]*types.Field{
								"id":   {Type: "string", Required: true},
								"name": {Type: "string", Required: true},
							},
						},
					},
				},
			}

			routeMap, err := loader.BuildRouteMap()
			if err != nil {
				t.Fatalf("BuildRouteMap() error = %v", err)
			}

			route, exists := routeMap[tt.entityName]
			if !exists {
				t.Fatalf("route for %q not found", tt.entityName)
			}

			if route.CollectionPath != tt.wantCollectionPath {
				t.Errorf("CollectionPath = %q, want %q", route.CollectionPath, tt.wantCollectionPath)
			}
			if route.ItemPath != tt.wantItemPath {
				t.Errorf("ItemPath = %q, want %q", route.ItemPath, tt.wantItemPath)
			}
		})
	}
}
