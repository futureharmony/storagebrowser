package minio

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// S3PermissionWrapper wraps an S3 filesystem to enforce bucket and scope permissions
type S3PermissionWrapper struct {
	fs     afero.Fs
	user   *User
	bucket string
	prefix string
}

// User represents a user with bucket and scope permissions
type User struct {
	Bucket string
	Scope  string
}

// NewS3PermissionWrapper creates a new S3PermissionWrapper
func NewS3PermissionWrapper(fs afero.Fs, user *User) *S3PermissionWrapper {
	var bucket, prefix string

	if user.Bucket != "" {
		bucket = user.Bucket
	} else {
		bucket = Cfg.Bucket // Use default bucket if none specified
	}

	if user.Scope != "" {
		prefix = strings.TrimPrefix(user.Scope, "/")
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
	}

	return &S3PermissionWrapper{
		fs:     fs,
		user:   user,
		bucket: bucket,
		prefix: prefix,
	}
}

// EnforcePathPermission checks if the given path is within the user's permissions
func (w *S3PermissionWrapper) EnforcePathPermission(p string) error {
	// Normalize the path
	p = strings.TrimPrefix(p, "/")

	// If user has no specific bucket restriction, allow access
	if w.user.Bucket == "" {
		return nil
	}

	// Check if the path is within the allowed bucket and prefix
	if w.prefix != "" {
		// If the path doesn't start with the allowed prefix, deny access
		if !strings.HasPrefix(p, w.prefix) {
			return os.ErrPermission
		}
	}

	return nil
}

// Stat retrieves a file info
func (w *S3PermissionWrapper) Stat(name string) (os.FileInfo, error) {
	if err := w.EnforcePathPermission(name); err != nil {
		return nil, err
	}
	return w.fs.Stat(name)
}

// Rename renames a file
func (w *S3PermissionWrapper) Rename(oldname, newname string) error {
	if err := w.EnforcePathPermission(oldname); err != nil {
		return err
	}
	if err := w.EnforcePathPermission(newname); err != nil {
		return err
	}
	return w.fs.Rename(oldname, newname)
}

// Remove removes a file
func (w *S3PermissionWrapper) Remove(name string) error {
	if err := w.EnforcePathPermission(name); err != nil {
		return err
	}
	return w.fs.Remove(name)
}

// RemoveAll removes all files in a directory
func (w *S3PermissionWrapper) RemoveAll(path string) error {
	if err := w.EnforcePathPermission(path); err != nil {
		return err
	}
	return w.fs.RemoveAll(path)
}

// Open opens a file
func (w *S3PermissionWrapper) Open(name string) (afero.File, error) {
	if err := w.EnforcePathPermission(name); err != nil {
		return nil, err
	}
	return w.fs.Open(name)
}

// OpenFile opens a file with flags
func (w *S3PermissionWrapper) OpenFile(name string, flag int, perm fs.FileMode) (afero.File, error) {
	if err := w.EnforcePathPermission(name); err != nil {
		return nil, err
	}
	return w.fs.OpenFile(name, flag, perm)
}

// Mkdir creates a directory
func (w *S3PermissionWrapper) Mkdir(name string, perm fs.FileMode) error {
	if err := w.EnforcePathPermission(name); err != nil {
		return err
	}
	return w.fs.Mkdir(name, perm)
}

// MkdirAll creates all directories
func (w *S3PermissionWrapper) MkdirAll(path string, perm fs.FileMode) error {
	if err := w.EnforcePathPermission(path); err != nil {
		return err
	}
	return w.fs.MkdirAll(path, perm)
}

// Create creates a file
func (w *S3PermissionWrapper) Create(name string) (afero.File, error) {
	if err := w.EnforcePathPermission(name); err != nil {
		return nil, err
	}
	return w.fs.Create(name)
}

// Name returns the name of the filesystem
func (w *S3PermissionWrapper) Name() string {
	return "S3PermissionWrapper"
}

// Chmod changes file permissions
func (w *S3PermissionWrapper) Chmod(name string, mode fs.FileMode) error {
	if err := w.EnforcePathPermission(name); err != nil {
		return err
	}
	return w.fs.Chmod(name, mode)
}

// Chtimes changes file times
func (w *S3PermissionWrapper) Chtimes(name string, atime, mtime time.Time) error {
	if err := w.EnforcePathPermission(name); err != nil {
		return err
	}
	return w.fs.Chtimes(name, atime, mtime)
}

// PathJoin joins file paths
func (w *S3PermissionWrapper) PathJoin(elem ...string) string {
	return path.Join(elem...)
}

// Chroot changes the root directory (not supported in this wrapper)
func (w *S3PermissionWrapper) Chroot(p string) (afero.Fs, error) {
	return nil, fmt.Errorf("Chroot not supported in S3PermissionWrapper")
}

// Root returns the root directory (not applicable for S3)
func (w *S3PermissionWrapper) Root() string {
	return "/"
}

// Dir returns the directory (not applicable for S3)
func (w *S3PermissionWrapper) Dir() string {
	return "/"
}

// As returns the underlying filesystem
func (w *S3PermissionWrapper) As(name string) (afero.File, error) {
	if err := w.EnforcePathPermission(name); err != nil {
		return nil, err
	}
	return w.fs.Open(name)
}

// Chown changes file ownership (not supported in S3)
func (w *S3PermissionWrapper) Chown(name string, uid, gid int) error {
	return fmt.Errorf("Chown not supported in S3PermissionWrapper")
}
