package middleware

import (
	"fmt"
)

func buildNameForConnection(ctx BuildContext, prefix string) (string, error) {
	portStr := ctx.PortString()
	if portStr == "" {
		return "", fmt.Errorf("port not found in middleware options for %s resolution", prefix)
	}
	conName := ctx.BuildConnectionName(prefix)
	return conName, nil
}

func ProxyPathName(prefix, port, path string) string {
	if path == "" {
		return ProxyName(prefix, port)
	}
	return fmt.Sprintf("%s-port-%s-path-%s", prefix, port, path)
}

func ProxyHostPathName(prefix, port, host, path string) string {
	if host == "" {
		return ProxyPathName(prefix, port, path)
	}
	if path == "" {
		return ProxyHostName(prefix, port, host)
	}
	return fmt.Sprintf("%s-port-%s-host-%s-path-%s", prefix, port, host, path)
}

func ProxyHostName(prefix, port, host string) string {
	if host == "" {
		return ProxyName(prefix, port)
	}
	return fmt.Sprintf("%s-port-%s-host-%s", prefix, port, host)
}

func ProxyName(prefix, port string) string {
	return fmt.Sprintf("%s-port-%s", prefix, port)
}
