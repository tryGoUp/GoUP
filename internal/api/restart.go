package api

import (
	"net/http"

	"github.com/mirkobrombin/goup/internal/restart"
)

func restartHandler(w http.ResponseWriter, r *http.Request) {
	restart.RestartHandler(w, r)
}
