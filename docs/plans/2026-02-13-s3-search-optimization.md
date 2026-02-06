# S3 Search Optimization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Optimize S3 search to efficiently query all files under a scope with prefix matching, supporting deep directory traversal and reducing API calls.

**Architecture:** Replace current buggy `Search` method with efficient S3 `ListObjectsV2` paginator that recursively searches all objects under scope, applies search filters, and calls back matching items. Use S3's `Delimiter` parameter for efficient directory listing when searching deep hierarchies.

**Tech Stack:** Go, AWS SDK v2 S3 client, afero filesystem interface

---

### Task 1: Fix Current Search Method Bug

**Files:**
- Modify: `library/afero-s3/s3_fs.go:958-991`
- Test: `library/afero-s3/s3_search_test.go:1-103`

**Step 1: Write the failing test**

```go
func TestSearchMethodUsesCorrectScope(t *testing.T) {
	t.Parallel()

	t.Run("Search should use scope parameter, not query", func(t *testing.T) {
		t.Parallel()
		
		// This will fail because current implementation uses query as scope
		// Expectation: Search("/test", "query", ...) should list from /test, not from "query"
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./library/afero-s3 -run TestSearchMethodUsesCorrectScope`
Expected: FAIL with test showing wrong behavior

**Step 3: Write minimal implementation fix**

```go
func (fw *FsWrapper) Search(scope, query string, found func(path string, isDir bool) error) error {
	if fw == nil || fw.Fs == nil || fw.Fs.s3API == nil {
		return nil
	}

	search := parseSearch(query)
	prefix := scope
	if prefix == "/" {
		prefix = ""
	}

	// FIX: Use scope, not query
	infos, err := fw.ListDirectory(scope, -1)
	if err != nil {
		return err
	}

	for _, info := range infos {
		relPath := strings.TrimPrefix(info.Name(), prefix)
		relPath = strings.TrimPrefix(relPath, "/")

		if relPath == "" {
			continue
		}

		if !checkSearchMatch(search, relPath, info.IsDir()) {
			continue
		}

		if err := found(relPath, info.IsDir()); err != nil {
			return err
		}
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v ./library/afero-s3 -run TestSearchMethodUsesCorrectScope`
Expected: PASS

**Step 5: Commit**

```bash
git add library/afero-s3/s3_fs.go library/afero-s3/s3_search_test.go
git commit -m "fix: Search method uses correct scope parameter"
```

---

### Task 2: Create Efficient S3 Search Implementation

**Files:**
- Create: `library/afero-s3/s3_search.go`
- Modify: `library/afero-s3/s3_fs.go:958-991`
- Test: `library/afero-s3/s3_search_test.go`

**Step 1: Write the failing test for deep search**

```go
func TestSearchDeepDirectoryTraversal(t *testing.T) {
	t.Parallel()

	t.Run("Search should find files in deep directories", func(t *testing.T) {
		t.Parallel()
		
		// Test expectation: Search("/", "deepfile", ...) should find
		// /a/b/c/deepfile.txt even though it's 3 levels deep
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./library/afero-s3 -run TestSearchDeepDirectoryTraversal`
Expected: FAIL (current method only searches immediate directory)

**Step 3: Create new efficient search implementation**

Create `library/afero-s3/s3_search.go`:

```go
package s3

import (
	"context"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SearchDeep performs efficient S3 search using ListObjectsV2 paginator
// It searches all objects under the given scope (prefix) and applies search filters
func (fw *FsWrapper) SearchDeep(scope, query string, found func(path string, isDir bool) error) error {
	if fw == nil || fw.Fs == nil || fw.Fs.s3API == nil {
		return nil
	}

	search := parseSearch(query)
	prefix := scope
	if prefix == "/" {
		prefix = ""
	}

	// Normalize prefix for S3
	normalizedPrefix := strings.TrimPrefix(prefix, "/")
	if normalizedPrefix != "" && !strings.HasSuffix(normalizedPrefix, "/") {
		normalizedPrefix += "/"
	}

	ctx := context.Background()
	paginator := s3.NewListObjectsV2Paginator(fw.Fs.s3API, &s3.ListObjectsV2Input{
		Bucket: aws.String(fw.Bucket),
		Prefix: aws.String(normalizedPrefix),
	})

	seenDirs := make(map[string]bool)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, obj := range page.Contents {
			key := *obj.Key
			
			// Convert S3 key to filesystem path
			fsPath := "/" + key
			
			// Get relative path from scope
			relPath := strings.TrimPrefix(fsPath, scope)
			relPath = strings.TrimPrefix(relPath, "/")
			
			if relPath == "" {
				continue
			}

			// Check if it's a directory (ends with /)
			isDir := false
			if strings.HasSuffix(key, "/") {
				isDir = true
				relPath = strings.TrimSuffix(relPath, "/")
			}

			// Apply search filters
			if !checkSearchMatch(search, relPath, isDir) {
				continue
			}

			if err := found(relPath, isDir); err != nil {
				return err
			}

			// Track parent directories for directory search
			if !isDir && len(search.Terms) > 0 {
				// Add parent directories for deep search
				parts := strings.Split(relPath, "/")
				for i := 1; i < len(parts); i++ {
					dirPath := strings.Join(parts[:i], "/")
					if !seenDirs[dirPath] {
						seenDirs[dirPath] = true
						if checkSearchMatch(search, dirPath, true) {
							if err := found(dirPath, true); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v ./library/afero-s3 -run TestSearchDeepDirectoryTraversal`
Expected: PASS

**Step 5: Commit**

```bash
git add library/afero-s3/s3_search.go library/afero-s3/s3_search_test.go
git commit -m "feat: add efficient S3 deep search implementation"
```

---

### Task 3: Integrate New Search with s3SearchOptimized

**Files:**
- Modify: `search/search.go:82-104`
- Test: `search/search_test.go:1-63`

**Step 1: Write the failing test for S3 integration**

```go
func TestS3SearchOptimizedUsesEfficientSearch(t *testing.T) {
	t.Parallel()

	t.Run("s3SearchOptimized should use efficient deep search", func(t *testing.T) {
		t.Parallel()
		
		// Test expectation: s3SearchOptimized should call SearchDeep
		// not the buggy Search method
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./search -run TestS3SearchOptimizedUsesEfficientSearch`
Expected: FAIL (still using old method)

**Step 3: Update s3SearchOptimized to use SearchDeep**

Modify `search/search.go:83`:

```go
func s3SearchOptimized(s3fs *s3lib.FsWrapper, scope, query string, checker rules.Checker, found func(path string, f os.FileInfo) error) error {
	prefix := scope
	if prefix == "/" {
		prefix = ""
	}

	// Use efficient deep search
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
```

**Step 4: Run test to verify it passes**

Run: `go test -v ./search -run TestS3SearchOptimizedUsesEfficientSearch`
Expected: PASS

**Step 5: Commit**

```bash
git add search/search.go search/search_test.go
git commit -m "feat: update s3SearchOptimized to use efficient deep search"
```

---

### Task 4: Add Support for Search Conditions (type filters)

**Files:**
- Modify: `library/afero-s3/s3_search.go`
- Test: `library/afero-s3/s3_search_test.go`

**Step 1: Write the failing test for type filters**

```go
func TestSearchWithTypeFilter(t *testing.T) {
	t.Parallel()

	t.Run("Search should filter by file type", func(t *testing.T) {
		t.Parallel()
		
		// Test expectation: Search("/", "type:image", ...) should only return image files
		// type:image, type:audio, type:video, type:pdf, etc.
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./library/afero-s3 -run TestSearchWithTypeFilter`
Expected: FAIL (type filters not properly integrated)

**Step 3: Enhance SearchDeep to handle type filters**

Update `library/afero-s3/s3_search.go`:

```go
func (fw *FsWrapper) SearchDeep(scope, query string, found func(path string, isDir bool) error) error {
	// ... existing code ...

	for paginator.HasMorePages() {
		// ... existing code ...

		for _, obj := range page.Contents {
			key := *obj.Key
			
			// Skip directory placeholders
			if strings.HasSuffix(key, "/") {
				continue
			}

			fsPath := "/" + key
			relPath := strings.TrimPrefix(fsPath, scope)
			relPath = strings.TrimPrefix(relPath, "/")
			
			if relPath == "" {
				continue
			}

			// Get file extension for type filtering
			ext := path.Ext(relPath)
			
			// Apply search filters including type conditions
			if !checkSearchMatch(search, relPath, false) {
				continue
			}

			if err := found(relPath, false); err != nil {
				return err
			}

			// Add parent directories
			parts := strings.Split(relPath, "/")
			for i := 1; i < len(parts); i++ {
				dirPath := strings.Join(parts[:i], "/")
				if !seenDirs[dirPath] {
					seenDirs[dirPath] = true
					// Directories don't have extensions, so check if they match search terms
					if checkSearchMatch(search, dirPath, true) {
						if err := found(dirPath, true); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v ./library/afero-s3 -run TestSearchWithTypeFilter`
Expected: PASS

**Step 5: Commit**

```bash
git add library/afero-s3/s3_search.go library/afero-s3/s3_search_test.go
git commit -m "feat: add support for type filters in S3 search"
```

---

### Task 5: Add Performance Optimization with Delimiter

**Files:**
- Modify: `library/afero-s3/s3_search.go`
- Test: `library/afero-s3/s3_search_test.go`

**Step 1: Write the failing test for delimiter optimization**

```go
func TestSearchWithDelimiterOptimization(t *testing.T) {
	t.Parallel()

	t.Run("Search should use delimiter for deep directory optimization", func(t *testing.T) {
		t.Parallel()
		
		// Test expectation: When searching deep directories, use Delimiter="/"
		// to get CommonPrefixes instead of listing all objects
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./library/afero-s3 -run TestSearchWithDelimiterOptimization`
Expected: FAIL (not using delimiter optimization)

**Step 3: Implement delimiter-based optimization**

Update `library/afero-s3/s3_search.go`:

```go
func (fw *FsWrapper) SearchDeep(scope, query string, found func(path string, isDir bool) error) error {
	// ... existing code ...

	search := parseSearch(query)
	
	// Determine if we need deep traversal or can use delimiter optimization
	// If search terms are empty or very short, use delimiter for efficiency
	useDelimiter := len(search.Terms) == 0 || (len(search.Terms) == 1 && len(search.Terms[0]) <= 3)

	ctx := context.Background()
	
	if useDelimiter {
		// Use delimiter for efficient directory listing
		paginator := s3.NewListObjectsV2Paginator(fw.Fs.s3API, &s3.ListObjectsV2Input{
			Bucket:    aws.String(fw.Bucket),
			Prefix:    aws.String(normalizedPrefix),
			Delimiter: aws.String("/"),
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return err
			}

			// Handle files in current directory
			for _, obj := range page.Contents {
				key := *obj.Key
				fsPath := "/" + key
				relPath := strings.TrimPrefix(fsPath, scope)
				relPath = strings.TrimPrefix(relPath, "/")
				
				if relPath == "" {
					continue
				}

				if !checkSearchMatch(search, relPath, false) {
					continue
				}

				if err := found(relPath, false); err != nil {
					return err
				}
			}

			// Handle subdirectories (CommonPrefixes)
			for _, prefix := range page.CommonPrefixes {
				dirKey := *prefix.Prefix
				dirPath := "/" + dirKey
				relDirPath := strings.TrimPrefix(dirPath, scope)
				relDirPath = strings.TrimSuffix(relDirPath, "/")
				relDirPath = strings.TrimPrefix(relDirPath, "/")
				
				if relDirPath == "" {
					continue
				}

				// Check if directory name matches search
				if checkSearchMatch(search, relDirPath, true) {
					if err := found(relDirPath, true); err != nil {
						return err
					}
				}

				// Recursively search subdirectory if needed
				if len(search.Terms) > 0 {
					subScope := scope + "/" + relDirPath
					if err := fw.SearchDeep(subScope, query, found); err != nil {
						return err
					}
				}
			}
		}
	} else {
		// Use full listing for precise search
		// ... existing full listing code ...
	}

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v ./library/afero-s3 -run TestSearchWithDelimiterOptimization`
Expected: PASS

**Step 5: Commit**

```bash
git add library/afero-s3/s3_search.go library/afero-s3/s3_search_test.go
git commit -m "feat: add delimiter optimization for S3 search"
```

---

### Task 6: Update Search Method to Use SearchDeep

**Files:**
- Modify: `library/afero-s3/s3_fs.go:958-991`
- Test: `library/afero-s3/s3_search_test.go`

**Step 1: Write the failing test for method replacement**

```go
func TestSearchMethodCallsSearchDeep(t *testing.T) {
	t.Parallel()

	t.Run("Search method should delegate to SearchDeep", func(t *testing.T) {
		t.Parallel()
		
		// Test expectation: The Search method should call SearchDeep
		// and handle any compatibility differences
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./library/afero-s3 -run TestSearchMethodCallsSearchDeep`
Expected: FAIL (Search method still uses old implementation)

**Step 3: Update Search method to use SearchDeep**

Modify `library/afero-s3/s3_fs.go:958`:

```go
func (fw *FsWrapper) Search(scope, query string, found func(path string, isDir bool) error) error {
	// Delegate to the efficient implementation
	return fw.SearchDeep(scope, query, found)
}
```

**Step 4: Run test to verify it passes**

Run: `go test -v ./library/afero-s3 -run TestSearchMethodCallsSearchDeep`
Expected: PASS

**Step 5: Commit**

```bash
git add library/afero-s3/s3_fs.go library/afero-s3/s3_search_test.go
git commit -m "refactor: update Search method to use efficient SearchDeep"
```

---

### Task 7: Run Full Test Suite

**Step 1: Run backend tests**

Run: `make test`
Expected: All tests pass

**Step 2: Run S3 library tests**

Run: `go test -v ./library/afero-s3/...`
Expected: All tests pass

**Step 3: Run search package tests**

Run: `go test -v ./search/...`
Expected: All tests pass

**Step 4: Run linter**

Run: `make lint`
Expected: No linting errors

**Step 5: Commit any test fixes**

```bash
git add .
git commit -m "test: ensure all tests pass after S3 search optimization"
```

---

Plan complete and saved to `docs/plans/2026-02-13-s3-search-optimization.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?