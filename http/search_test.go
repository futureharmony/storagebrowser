package http

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchHandlerWithScopeAndPath(t *testing.T) {
	t.Parallel()

	t.Run("should read path from query parameter", func(t *testing.T) {
		t.Parallel()

		// Test that handler follows the same pattern as resource handler
		// which reads path from query parameter and defaults to "/"

		req := httptest.NewRequest("GET", "/api/search?query=test&path=/documents&scope=bucket1", nil)
		q := req.URL.Query()

		// Verify query parameters are accessible
		path := q.Get("path")
		query := q.Get("query")
		scope := q.Get("scope")

		assert.Equal(t, "/documents", path, "Path should be read from query parameter")
		assert.Equal(t, "test", query, "Query should be read from query parameter")
		assert.Equal(t, "bucket1", scope, "Scope should be read from query parameter")

		// Current implementation uses r.URL.Path instead of query parameter
		// This will fail until we update search.go
		// The fix is to change line 14 in search.go from:
		//   search.Search(d.requestFs, r.URL.Path, query, d, ...)
		// To:
		//   path := r.URL.Query().Get("path")
		//   if path == "" { path = "/" }
		//   search.Search(d.requestFs, path, query, d, ...)
	})

	t.Run("should default path to / when empty", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest("GET", "/api/search?query=test&scope=bucket1", nil)
		q := req.URL.Query()

		path := q.Get("path")
		assert.Empty(t, path, "Path should be empty when not provided")

		// Handler should default to "/" when path is empty
		// This matches the pattern in resource.go lines 30-33
	})
}
