package http

import (
	"net/http"
	"os"

	"github.com/futureharmony/storagebrowser/v2/search"
)

var searchHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	response := []map[string]interface{}{}
	query := r.URL.Query().Get("query")

	// Get path from query parameter (default to "/")
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	err := search.Search(d.requestFs, path, query, d, func(path string, f os.FileInfo) error {
		if f == nil {
			return nil
		}
		response = append(response, map[string]interface{}{
			"dir":  f.IsDir(),
			"path": path,
		})

		return nil
	})

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, response)
})
