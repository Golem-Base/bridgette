package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
)

//go:embed static
var staticFiles embed.FS

// createStaticHandler creates and returns an HTTP handler for serving static files
func createStaticHandler() http.Handler {
	// Create a sub-filesystem to trim the static prefix
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set appropriate content type based on file extension
		ext := path.Ext(r.URL.Path)
		switch ext {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		}

		// Use http.FileServer to serve the embedded files
		http.FileServer(http.FS(staticFS)).ServeHTTP(w, r)
	})
}
