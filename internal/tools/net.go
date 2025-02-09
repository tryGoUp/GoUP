package tools

import (
	"fmt"
	"net"
)

// getFreePort returns an available port from a high range.
func GetFreePort() (string, error) {
	const basePort = 30000
	const maxPort = 40000
	for port := basePort; port < maxPort; port++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return fmt.Sprintf("%d", port), nil
		}
	}
	return "", fmt.Errorf("no free port available in range %d-%d", basePort, maxPort)
}
