package frontend

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestHandlerWithoutBuiltFrontend(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler(fstest.MapFS{
		"dist/.keep": &fstest.MapFile{},
	}).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("response status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestHandlerServesFrontend(t *testing.T) {
	staticFiles := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("index")},
		"dist/app.js":     &fstest.MapFile{Data: []byte("app")},
	}
	tests := []struct {
		name string
		path string
		body string
	}{
		{name: "index", path: "/", body: "index"},
		{name: "static asset", path: "/app.js", body: "app"},
		{name: "frontend route", path: "/clusters/1", body: "index"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			handler(staticFiles).ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("response status = %d, want %d", response.Code, http.StatusOK)
			}
			if response.Body.String() != test.body {
				t.Fatalf("response body = %q, want %q", response.Body.String(), test.body)
			}
		})
	}
}

func TestStaticFileExistsRejectsDirectories(t *testing.T) {
	dist := fstest.MapFS{
		"assets/app.js": &fstest.MapFile{Data: []byte("app")},
	}
	sub, err := fs.Sub(dist, "assets")
	if err != nil {
		t.Fatal(err)
	}

	if staticFileExists(sub, "/assets") {
		t.Fatal("staticFileExists returned true for a directory")
	}
}
