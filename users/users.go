package users

import (
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/futureharmony/storagebrowser/v2/errors"
	"github.com/futureharmony/storagebrowser/v2/files"
	"github.com/futureharmony/storagebrowser/v2/minio"
	"github.com/futureharmony/storagebrowser/v2/rules"
)

// ViewMode describes a view mode.
type ViewMode string

const (
	ListViewMode   ViewMode = "list"
	MosaicViewMode ViewMode = "mosaic"
)

// User describes a user.
type User struct {
	ID              uint          `storm:"id,increment" json:"id"`
	Username        string        `storm:"unique" json:"username"`
	Password        string        `json:"password"`
	AvailableScopes []Scope       `json:"availableScopes"`
	CurrentScope    Scope         `json:"currentScope"`
	Scope           string        `json:"scope"`
	Locale          string        `json:"locale"`
	LockPassword    bool          `json:"lockPassword"`
	ViewMode        ViewMode      `json:"viewMode"`
	SingleClick     bool          `json:"singleClick"`
	Perm            Permissions   `json:"perm"`
	Commands        []string      `json:"commands"`
	Sorting         files.Sorting `json:"sorting"`
	Fs              afero.Fs      `json:"-" yaml:"-"`
	Rules           []rules.Rule  `json:"rules"`
	HideDotfiles    bool          `json:"hideDotfiles"`
	DateFormat      bool          `json:"dateFormat"`
}

type Scope struct {
	Name       string `json:"name"`
	RootPrefix string `json:"rootPrefix"`
}

// SetS3Scopes sets up available scopes for S3 storage type from an array of Scope objects
func (u *User) SetS3Scopes(scopes []Scope) {
	u.AvailableScopes = scopes
	if len(scopes) > 0 && u.CurrentScope.Name == "" {
		u.CurrentScope = scopes[0] // Set first scope as current if not set
	}
}

// GetRules implements rules.Provider.
func (u *User) GetRules() []rules.Rule {
	return u.Rules
}

var checkableFields = []string{
	"Username",
	"Password",
	"Bucket",
	"Scope",
	"ViewMode",
	"Commands",
	"Sorting",
	"Rules",
	"AvailableScopes",
	"CurrentScope",
}

// Clean cleans up a user and verifies if all its fields
// are alright to be saved.
//
//nolint:gocyclo
func (u *User) Clean(baseScope string, fields ...string) error {
	if len(fields) == 0 {
		fields = checkableFields
	}

	for _, field := range fields {
		switch field {
		case "Username":
			if u.Username == "" {
				return errors.ErrEmptyUsername
			}
		case "Password":
			if u.Password == "" {
				return errors.ErrEmptyPassword
			}
		case "Bucket":
			// Allow empty bucket (means all buckets)
			continue
		case "ViewMode":
			if u.ViewMode == "" {
				u.ViewMode = ListViewMode
			}
		case "Commands":
			if u.Commands == nil {
				u.Commands = []string{}
			}
		case "Sorting":
			if u.Sorting.By == "" {
				u.Sorting.By = "name"
			}
		case "Rules":
			if u.Rules == nil {
				u.Rules = []rules.Rule{}
			}
		case "AvailableScopes":
			if u.AvailableScopes == nil {
				u.AvailableScopes = []Scope{}
			}
		case "CurrentScope":
			// CurrentScope is a struct, ensure it has valid values
			if u.CurrentScope.RootPrefix == "" {
				u.CurrentScope.RootPrefix = "/"
			}
		}
	}

	if u.Fs == nil {
		scope := filepath.Join(baseScope, filepath.Join("/", u.CurrentScope.RootPrefix)) //nolint:gocritic
		// Create a user-specific filesystem wrapper if using S3 storage
		u.Fs = minio.CreateUserFs(u.CurrentScope.Name, scope)
	}

	return nil
}

// FullPath gets the full path for a user's relative path.
func (u *User) FullPath(path string) string {
	return minio.FullPath(u.Fs, path)
}
