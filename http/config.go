package http

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/futureharmony/storagebrowser/v2/auth"
	"github.com/futureharmony/storagebrowser/v2/version"
)

var configHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	auther, err := d.store.Auth.Get(d.settings.AuthMethod)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	config := map[string]interface{}{
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
		_, err := os.Stat(fPath)
		if err == nil {
			config["CSS"] = true
		}
	}

	if d.settings.AuthMethod == auth.MethodJSONAuth {
		raw, err := d.store.Auth.Get(d.settings.AuthMethod)
		if err == nil {
			jsonAuther := raw.(*auth.JSONAuth)
			if jsonAuther.ReCaptcha != nil {
				config["ReCaptcha"] = jsonAuther.ReCaptcha.Key != "" && jsonAuther.ReCaptcha.Secret != ""
				config["ReCaptchaHost"] = jsonAuther.ReCaptcha.Host
				config["ReCaptchaKey"] = jsonAuther.ReCaptcha.Key
			}
		}
	}

	return renderJSON(w, r, config)
})
