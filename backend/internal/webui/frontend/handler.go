package frontend

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed dist
var staticFiles embed.FS

func Handler() http.Handler {
	dist, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		return http.NotFoundHandler()
	}

	fileServer := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		if staticFileExists(dist, r.URL.Path) {
			fileServer.ServeHTTP(w, r)
			return
		}

		indexRequest := new(http.Request)
		*indexRequest = *r
		indexURL := *r.URL
		indexURL.Path = "/"
		indexRequest.URL = &indexURL
		fileServer.ServeHTTP(w, indexRequest)
	})
}

func staticFileExists(dist fs.FS, requestPath string) bool {
	name := strings.TrimPrefix(path.Clean(requestPath), "/")
	if name == "." || name == "" {
		return true
	}

	file, err := dist.Open(name)
	if err != nil {
		return false
	}
	defer file.Close()

	stat, err := file.Stat()
	return err == nil && !stat.IsDir()
}
