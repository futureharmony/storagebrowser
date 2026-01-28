package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/futureharmony/storagebrowser/v2/auth"
	"github.com/futureharmony/storagebrowser/v2/settings"
	"github.com/futureharmony/storagebrowser/v2/storage"
	"github.com/futureharmony/storagebrowser/v2/version"
)

func handleWithStaticData(w http.ResponseWriter, _ *http.Request, d *data, fSys fs.FS, file, contentType string) (int, error) {
	w.Header().Set("Content-Type", contentType)

	auther, err := d.store.Auth.Get(d.settings.AuthMethod)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	data := map[string]interface{}{
		"Name":                  d.settings.Branding.Name,
		"DisableExternal":       d.settings.Branding.DisableExternal,
		"DisableUsedPercentage": d.settings.Branding.DisableUsedPercentage,
		"Color":                 d.settings.Branding.Color,
		"BaseURL":               d.server.BaseURL,
		"Version":               version.Version,
		"StaticURL":             path.Join(d.server.BaseURL, "/static"),
		"Signup":                d.settings.Signup,
		"NoAuth":                d.settings.AuthMethod == auth.MethodNoAuth,
		"AuthMethod":            d.settings.AuthMethod,
		"LoginPage":             auther.LoginPage(),
		"CSS":                   false,
		"ReCaptcha":             false,
		"Theme":                 d.settings.Branding.Theme,
		"EnableThumbs":          d.server.EnableThumbnails,
		"ResizePreview":         d.server.ResizePreview,
		"EnableExec":            d.server.EnableExec,
		"TusSettings":           d.settings.Tus,
		"StorageType":           d.server.StorageType,
	}

	if d.settings.Branding.Files != "" {
		fPath := filepath.Join(d.settings.Branding.Files, "custom.css")
		_, err := os.Stat(fPath) //nolint:govet

		if err != nil && !os.IsNotExist(err) {
			log.Printf("couldn't load custom styles: %v", err)
		}

		if err == nil {
			data["CSS"] = true
		}
	}

	if d.settings.AuthMethod == auth.MethodJSONAuth {
		raw, err := d.store.Auth.Get(d.settings.AuthMethod) //nolint:govet
		if err != nil {
			return http.StatusInternalServerError, err
		}

		auther := raw.(*auth.JSONAuth)

		if auther.ReCaptcha != nil {
			data["ReCaptcha"] = auther.ReCaptcha.Key != "" && auther.ReCaptcha.Secret != ""
			data["ReCaptchaHost"] = auther.ReCaptcha.Host
			data["ReCaptchaKey"] = auther.ReCaptcha.Key
		}
	}

	b, err := json.Marshal(data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	data["Json"] = strings.ReplaceAll(string(b), `'`, `\'`)

	fileContents, err := fs.ReadFile(fSys, file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return http.StatusNotFound, err
		}
		return http.StatusInternalServerError, err
	}
	index := template.Must(template.New("index").Delims("[{[", "]}]").Parse(string(fileContents)))
	err = index.Execute(w, data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func getStaticHandlers(store *storage.Storage, server *settings.Server, assetsFs fs.FS) (index, static http.Handler) {
	index = handle(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Method != http.MethodGet {
			return http.StatusNotFound, nil
		}

		w.Header().Set("x-xss-protection", "1; mode=block")
		return handleWithStaticData(w, r, d, assetsFs, "public/index.html", "text/html; charset=utf-8")
	}, "", store, server)

	static = handle(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.Method != http.MethodGet {
			return http.StatusNotFound, nil
		}

		// stripPrefix removed /static/ prefix, path is relative
		path := strings.TrimPrefix(r.URL.Path, "/static/")

		if strings.HasSuffix(path, "/") {
			return http.StatusNotFound, nil
		}

		const maxAge = 86400 // 1 day
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%v", maxAge))

		if d.settings.Branding.Files != "" {
			if strings.HasPrefix(path, "img/") {
				fPath := filepath.Join(d.settings.Branding.Files, path)
				_, err := os.Stat(fPath) //nolint:govet
				if err != nil && !os.IsNotExist(err) {
					log.Printf("could not load branding file override: %v", err)
				} else if err == nil {
					http.ServeFile(w, r, fPath)
					return 0, nil
				}
			} else if path == "custom.css" && d.settings.Branding.Files != "" {
				http.ServeFile(w, r, filepath.Join(d.settings.Branding.Files, "custom.css"))
				return 0, nil
			}
		}

		// Map URL path to embed.FS path
		// - Most paths map directly (img/*, assets/*)
		// - /static/img/logo.svg maps to static/img/logo.svg (only logo is in static/)
		// - JS files are gzip-compressed, add .gz extension
		fsPath := path
		if path == "img/logo.svg" {
			fsPath = "static/img/logo.svg"
		}
		if strings.HasSuffix(path, ".js") {
			fsPath = path + ".gz"
		}

		fileContents, err := fs.ReadFile(assetsFs, fsPath)
		if err != nil {
			return http.StatusNotFound, err
		}

		// Set appropriate headers based on file type
		if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".js.gz") {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if strings.HasSuffix(path, ".svg") {
			w.Header().Set("Content-Type", "image/svg+xml")
		} else if strings.HasSuffix(path, ".woff") || strings.HasSuffix(path, ".woff2") {
			w.Header().Set("Content-Type", "font/woff2")
		}

		if _, err := w.Write(fileContents); err != nil {
			return http.StatusInternalServerError, err
		}
		return 0, nil
	}, "/static/", store, server)

	return index, static
}
