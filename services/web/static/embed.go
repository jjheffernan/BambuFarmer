package static

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed css/app.css manifest.webmanifest sw.js icons
var assets embed.FS

// Handler serves embedded static assets under /static/.
// Service worker responses include Service-Worker-Allowed: / so the shell SW
// can register with scope "/" while only caching /static/* assets.
func Handler() http.Handler {
	sub, err := fs.Sub(assets, ".")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/static/")
		if path == "sw.js" {
			w.Header().Set("Service-Worker-Allowed", "/")
			w.Header().Set("Cache-Control", "no-cache")
		}
		if path == "manifest.webmanifest" {
			w.Header().Set("Content-Type", "application/manifest+json")
		}
		http.StripPrefix("/static/", files).ServeHTTP(w, r)
	})
}
