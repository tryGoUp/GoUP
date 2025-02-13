package api

import (
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mirkobrombin/goup/internal/config"
)

func getLogsHandler(w http.ResponseWriter, r *http.Request) {
	logFile := os.Getenv("GOUP_LOG_FILE")
	if logFile == "" {
		http.Error(w, "Log file not set", http.StatusNotFound)
		return
	}
	data, err := ioutil.ReadFile(logFile)
	if err != nil {
		http.Error(w, "Unable to read log file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}

func getMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"requests_total": 1234,
		"latency_avg_ms": 45.6,
		"cpu_usage":      23.4,
		"ram_usage_mb":   512,
		"active_sites":   len(config.SiteConfigs),
		"active_plugins": len(config.GlobalConf.EnabledPlugins),
	}
	jsonResponse(w, metrics)
}

func getStatusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"uptime":   "72h",
		"sites":    len(config.SiteConfigs),
		"plugins":  config.GlobalConf.EnabledPlugins,
		"apiAlive": true,
	}
	jsonResponse(w, status)
}

func getLogWeightHandler(w http.ResponseWriter, r *http.Request) {
	logDir := config.GetLogDir()
	var totalSize int64 = 0
	err := filepath.Walk(logDir, func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		http.Error(w, "Error calculating log weight", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, map[string]interface{}{
		"log_weight_bytes": totalSize,
	})
}

func getPluginUsageHandler(w http.ResponseWriter, r *http.Request) {
	usage := make(map[string]int)
	for _, site := range config.SiteConfigs {
		for pluginName := range site.PluginConfigs {
			usage[pluginName]++
		}
	}
	jsonResponse(w, usage)
}
