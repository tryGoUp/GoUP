package api

import (
	"encoding/json"
	"net/http"

	"github.com/mirkobrombin/goup/internal/config"
)

func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	if config.GlobalConf == nil {
		http.Error(w, "Global config not loaded", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, config.GlobalConf)
}

func updateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if config.GlobalConf == nil {
		http.Error(w, "Global config not loaded", http.StatusInternalServerError)
		return
	}
	var newConf config.GlobalConfig
	if err := json.NewDecoder(r.Body).Decode(&newConf); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	config.GlobalConf = &newConf
	if err := config.SaveGlobalConfig(); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, config.GlobalConf)
}
