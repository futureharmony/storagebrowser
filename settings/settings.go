package settings

import (
	"crypto/rand"
	"errors"
	"io/fs"
	"log"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/rules"
)

const DefaultUsersHomeBasePath = "/users"
const DefaultMinimumPasswordLength = 12
const DefaultFileMode = 0640
const DefaultDirMode = 0750

// AuthMethod describes an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Key                   []byte              `json:"key"`
	Signup                bool                `json:"signup"`
	CreateUserDir         bool                `json:"createUserDir"`
	UserHomeBasePath      string              `json:"userHomeBasePath"`
	Defaults              UserDefaults        `json:"defaults"`
	AuthMethod            AuthMethod          `json:"authMethod"`
	Branding              Branding            `json:"branding"`
	Tus                   Tus                 `json:"tus"`
	Commands              map[string][]string `json:"commands"`
	Shell                 []string            `json:"shell"`
	Rules                 []rules.Rule        `json:"rules"`
	MinimumPasswordLength uint                `json:"minimumPasswordLength"`
	FileMode              fs.FileMode         `json:"fileMode"`
	DirMode               fs.FileMode         `json:"dirMode"`
}

// GetRules implements rules.Provider.
func (s *Settings) GetRules() []rules.Rule {
	return s.Rules
}

// Server specific settings.
type Server struct {
	Root                  string `json:"root"`
	BaseURL               string `json:"baseURL"`
	Socket                string `json:"socket"`
	TLSKey                string `json:"tlsKey"`
	TLSCert               string `json:"tlsCert"`
	Port                  string `json:"port"`
	Address               string `json:"address"`
	Log                   string `json:"log"`
	EnableThumbnails      bool   `json:"enableThumbnails"`
	ResizePreview         bool   `json:"resizePreview"`
	EnableExec            bool   `json:"enableExec"`
	TypeDetectionByHeader bool   `json:"typeDetectionByHeader"`
	AuthHook              string `json:"authHook"`
	TokenExpirationTime   string `json:"tokenExpirationTime"`
	StorageType           string `json:"storageType"`
	S3Bucket              string `json:"s3Bucket"`
	S3Endpoint            string `json:"s3Endpoint"`
	S3AccessKey           string `json:"s3AccessKey"`
	S3SecretKey           string `json:"s3SecretKey"`
	S3Region              string `json:"s3Region"`
}

// Clean cleans any variables that might need cleaning.
func (s *Server) Clean() {
	s.BaseURL = strings.TrimSuffix(s.BaseURL, "/")

	if s.StorageType == "" {
		s.StorageType = "s3"
	}

	if s.S3Bucket == "" {
		s.S3Bucket = ""
	}
	if s.S3Endpoint == "" {
		s.S3Endpoint = ""
	}
	if s.S3AccessKey == "" {
		s.S3AccessKey = ""
	}
	if s.S3SecretKey == "" {
		s.S3SecretKey = ""
	}
	if s.S3Region == "" {
		s.S3Region = "us-east-1"
	}
}

func (s *Server) GetTokenExpirationTime(fallback time.Duration) time.Duration {
	if s.TokenExpirationTime == "" {
		return fallback
	}

	duration, err := time.ParseDuration(s.TokenExpirationTime)
	if err != nil {
		log.Printf("[WARN] Failed to parse tokenExpirationTime: %v", err)
		return fallback
	}
	return duration
}

// Validate validates the server configuration.
func (s *Server) Validate() error {
	if s.StorageType == "s3" {
		if s.S3Endpoint == "" {
			return errors.New("s3Endpoint is required when storageType is 's3'")
		}
		if s.S3AccessKey == "" {
			return errors.New("s3AccessKey is required when storageType is 's3'")
		}
		if s.S3SecretKey == "" {
			return errors.New("s3SecretKey is required when storageType is 's3'")
		}
		if s.S3Bucket == "" {
			return errors.New("s3Bucket is required when storageType is 's3'")
		}
	}
	return nil
}

// GenerateKey generates a key of 512 bits.
func GenerateKey() ([]byte, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
