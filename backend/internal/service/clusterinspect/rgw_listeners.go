package clusterinspect

import (
	"net"
	"sort"
	"strconv"
	"strings"
)

type RGWListener struct {
	Frontend string `json:"frontend"`
	TLS      bool   `json:"tls"`
	Port     uint16 `json:"port"`
}

// Only extract listener ports, never return the raw configuration or TLS paths.
func rgwListeners(metadata map[string]string) ([]RGWListener, bool) {
	result := []RGWListener{}
	complete := true
	keys := []string{}
	for key := range metadata {
		if strings.HasPrefix(key, "frontend_config#") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return result, false
	}
	for _, key := range keys {
		tokens := strings.Fields(metadata[key])
		if len(tokens) == 0 || tokens[0] != "beast" {
			complete = false
			continue
		}
		found := false
		for _, token := range tokens[1:] {
			name, value, ok := strings.Cut(token, "=")
			if !oneOf(name, "port", "ssl_port", "endpoint", "ssl_endpoint") {
				continue
			}
			found = true
			tls := strings.HasPrefix(name, "ssl_")
			if !ok {
				complete = false
				continue
			}
			if strings.HasSuffix(name, "endpoint") {
				host, port, err := net.SplitHostPort(value)
				if err == nil && net.ParseIP(host) != nil {
					value = port
				} else {
					address := strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
					if net.ParseIP(address) == nil {
						complete = false
						continue
					}
					// Native unbracketed IPv6 is not supported by parse_endpoint.
					if strings.Contains(address, ":") && (!strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]")) {
						complete = false
						continue
					}
					value = "80"
					if tls {
						value = "443"
					}
				}
			}
			port, err := strconv.ParseUint(value, 10, 16)
			if err != nil || port == 0 {
				complete = false
				continue
			}
			result = append(result, RGWListener{Frontend: key, TLS: tls, Port: uint16(port)})
		}
		if !found {
			complete = false
		}
	}
	return result, complete
}
