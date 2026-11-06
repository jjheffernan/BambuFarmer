package static

import (
	"embed"
	"net/http"
)

//go:embed css/app.css
var assets embed.FS

// Handler serves embedded static assets under /static/.
func Handler() http.Handler {
	return http.StripPrefix("/static/", http.FileServer(http.FS(assets)))
}
