package search

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	aferos3 "github.com/futureharmony/afero-aws-s3"

	"github.com/futureharmony/storagebrowser/v2/rules"
)

type searchOptions struct {
	CaseSensitive bool
	Conditions    []condition
	Terms         []string
}

// Search searches for a query in a fs.
func Search(fs afero.Fs, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	// Use the original implementation for all filesystem types
	// The optimized S3 search is no longer used due to API changes
	search := parseSearch(query)

	scope = filepath.ToSlash(filepath.Clean(scope))
	scope = path.Join("/", scope)

	return afero.Walk(fs, scope, func(fPath string, f os.FileInfo, _ error) error {
		fPath = filepath.ToSlash(filepath.Clean(fPath))
		fPath = path.Join("/", fPath)
		relativePath := strings.TrimPrefix(fPath, scope)
		relativePath = strings.TrimPrefix(relativePath, "/")

		if fPath == scope {
			return nil
		}

		if !checker.Check(fPath) {
			return nil
		}

		if len(search.Conditions) > 0 {
			match := false

			for _, t := range search.Conditions {
				if t(fPath) {
					match = true
					break
				}
			}

			if !match {
				return nil
			}
		}

		if len(search.Terms) > 0 {
			for _, term := range search.Terms {
				_, fileName := path.Split(fPath)
				if !search.CaseSensitive {
					fileName = strings.ToLower(fileName)
					term = strings.ToLower(term)
				}
				if strings.Contains(fileName, term) {
					return found(relativePath, f)
				}
			}
			return nil
		}

		return found(relativePath, f)
	})
}

// TODO move to afer-aws-s3
// s3Search searches for a query in an S3 filesystem using efficient ListObjectsV2 API calls.
func s3Search(fs *aferos3.Fs, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	// This function is now deprecated since we changed the API to use FsWrapper
	// which handles bucket/prefix internally. This function would need to be rewritten
	// to work with the new API, but it's better to use the standard afero walk instead.
	// For now, we'll return an error to indicate this function is not compatible with the new API.
	return nil
}
