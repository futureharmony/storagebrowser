package search

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"

	s3lib "github.com/futureharmony/afero-aws-s3"

	"github.com/futureharmony/storagebrowser/v2/rules"
)

type searchOptions struct {
	CaseSensitive bool
	Conditions    []condition
	Terms         []string
}

// Search searches for a query in a fs.
func Search(fs afero.Fs, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	search := parseSearch(query)

	scope = filepath.ToSlash(filepath.Clean(scope))
	scope = path.Join("/", scope)

	// Check if this is an S3 filesystem for efficient search
	if s3wrapper, ok := fs.(*s3lib.FsWrapper); ok {
		return s3SearchOptimized(s3wrapper, scope, query, checker, found)
	}

	// Use the original implementation for all filesystem types
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

// s3SearchOptimized performs efficient S3 search using library/afero-s3 SearchDeep method
func s3SearchOptimized(s3fs *s3lib.FsWrapper, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	prefix := scope
	if prefix == "/" {
		prefix = ""
	}

	return s3fs.SearchDeep(scope, query, func(relPath string, isDir bool) error {
		fullPath := path.Join("/", relPath)
		if !checker.Check(fullPath) {
			return nil
		}

		mockInfo := &s3FileInfo{
			name:    relPath,
			size:    0,
			modTime: time.Now(),
			isDir:   isDir,
		}

		return found(relPath, mockInfo)
	})
}

type s3FileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (i *s3FileInfo) Name() string { return i.name }
func (i *s3FileInfo) Size() int64  { return i.size }
func (i *s3FileInfo) Mode() os.FileMode {
	if i.isDir {
		return os.ModeDir | 0755
	}
	return 0644
}
func (i *s3FileInfo) ModTime() time.Time { return i.modTime }
func (i *s3FileInfo) IsDir() bool        { return i.isDir }
func (i *s3FileInfo) Sys() interface{}   { return nil }
