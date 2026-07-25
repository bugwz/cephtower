package webui

import (
	"net/http"

	"cephtower/backend/internal/webui/frontend"
)

func WebUIHandler() http.Handler {
	return frontend.Handler()
}
