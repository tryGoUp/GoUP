package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes returns a configured router.
func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Ping
	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, map[string]string{"message": "pong"})
	}).Methods("GET")

	// Plugins
	r.HandleFunc("/api/plugins", getPluginsHandler).Methods("GET")
	r.HandleFunc("/api/plugins/{pluginName}/toggle", togglePluginHandler).Methods("POST")

	// Global config
	r.HandleFunc("/api/config", getConfigHandler).Methods("GET")
	r.HandleFunc("/api/config", updateConfigHandler).Methods("PUT")

	// Logs, metrics, status
	r.HandleFunc("/api/logs", getLogsHandler).Methods("GET")
	r.HandleFunc("/api/metrics", getMetricsHandler).Methods("GET")
	r.HandleFunc("/api/status", getStatusHandler).Methods("GET")
	r.HandleFunc("/api/logweight", getLogWeightHandler).Methods("GET")
	r.HandleFunc("/api/pluginusage", getPluginUsageHandler).Methods("GET")

	// Log files
	r.HandleFunc("/api/logfiles", listLogFilesHandler).Methods("GET")
	r.HandleFunc("/api/logfiles/{fileName:.*}", getLogFileHandler).Methods("GET")

	// Restart
	r.HandleFunc("/api/restart", restartHandler).Methods("POST")

	// Sites
	r.HandleFunc("/api/sites", listSitesHandler).Methods("GET")
	r.HandleFunc("/api/sites", createSiteHandler).Methods("POST")
	r.HandleFunc("/api/sites/{domain}", getSiteHandler).Methods("GET")
	r.HandleFunc("/api/sites/{domain}", updateSiteHandler).Methods("PUT")
	r.HandleFunc("/api/sites/{domain}", deleteSiteHandler).Methods("DELETE")
	r.HandleFunc("/api/sites/{domain}/validate", validateSiteHandler).Methods("GET")

	return r
}
