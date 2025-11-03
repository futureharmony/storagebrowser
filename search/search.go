package search

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/afero"
	aferos3 "github.com/futureharmony/afero-aws-s3"

	"github.com/filebrowser/filebrowser/v2/rules"
)

type searchOptions struct {
	CaseSensitive bool
	Conditions    []condition
	Terms         []string
}

// Search searches for a query in a fs.
func Search(fs afero.Fs, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	// Check if the filesystem is an S3 filesystem and use optimized S3 search
	if s3fs, ok := fs.(*aferos3.Fs); ok {
		return s3Search(s3fs, scope, query, checker, found)
	}

	// Use the original implementation for other filesystem types
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

// s3Search searches for a query in an S3 filesystem using efficient ListObjectsV2 API calls.
func s3Search(fs *aferos3.Fs, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	search := parseSearch(query)

	scope = filepath.ToSlash(filepath.Clean(scope))
	scope = path.Join("/", scope)
	
	// Remove leading slash for S3 prefix
	s3Prefix := strings.TrimPrefix(scope, "/")
	if s3Prefix != "" && !strings.HasSuffix(s3Prefix, "/") {
		s3Prefix += "/"
	}
	
	// If scope is root, keep it as empty string for S3
	if scope == "/" {
		s3Prefix = ""
	}

	paginator := s3.NewListObjectsV2Paginator(fs.S3API(), &s3.ListObjectsV2Input{
		Bucket: aws.String(fs.Bucket()),
		Prefix: aws.String(s3Prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return err
		}

		for _, obj := range page.Contents {
			objectKey := *obj.Key

			// Skip if the object key is exactly the scope (directory placeholder)
			if objectKey == strings.TrimSuffix(s3Prefix, "/") {
				continue
			}

			// Convert S3 key to the file path format expected by the rest of the system
			fPath := path.Join("/", objectKey)
			
			// Get just the filename from the path for term matching
			_, fileName := path.Split(fPath)
			
			// For relative path calculation, remove the scope prefix
			relativePath := strings.TrimPrefix(fPath, scope)
			relativePath = strings.TrimPrefix(relativePath, "/")

			// Check permissions with the checker
			if !checker.Check(fPath) {
				continue
			}

			// Check conditions if any
			if len(search.Conditions) > 0 {
				match := false
				for _, t := range search.Conditions {
					if t(fPath) {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}

			// Check search terms if any
			if len(search.Terms) > 0 {
				termMatched := false
				for _, term := range search.Terms {
					fileNameToCheck := fileName
					termToCheck := term
					if !search.CaseSensitive {
						fileNameToCheck = strings.ToLower(fileNameToCheck)
						termToCheck = strings.ToLower(termToCheck)
					}
					if strings.Contains(fileNameToCheck, termToCheck) {
						termMatched = true
						break
					}
				}
				if !termMatched {
					continue
				}
			}

			// If we reach here, the file passed all checks, so call found
			fileInfo := aferos3.NewFileInfo(path.Base(objectKey), false, *obj.Size, *obj.LastModified)
			if err := found(relativePath, fileInfo); err != nil {
				return err
			}
		}
	}

	return nil
}
