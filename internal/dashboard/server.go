package dashboard

import (
	"fmt"
	"net/http"

	"github.com/mirkobrombin/goup/internal/config"
)

// StartDashboardServer starts a dedicated server for the dashboard.
func StartDashboardServer() {
	if config.GlobalConf == nil || config.GlobalConf.DashboardPort == 0 {
		return
	}
	port := config.GlobalConf.DashboardPort
	go func() {
		fmt.Printf("[Dashboard] Listening on :%d\n", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), Handler()); err != nil {
			fmt.Printf("[Dashboard] Error: %v\n", err)
		}
	}()
}
