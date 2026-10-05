package clusterinspect

import (
	"net/netip"
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
				var valid bool
				value, valid = rgwEndpointPort(value, tls)
				if !valid {
					complete = false
					continue
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

func rgwEndpointPort(value string, tls bool) (string, bool) {
	fallback := "80"
	if tls {
		fallback = "443"
	}
	if strings.HasPrefix(value, "[") {
		end := strings.IndexByte(value, ']')
		if end < 0 {
			return "", false
		}
		address, err := netip.ParseAddr(value[1:end])
		if err != nil || !address.Is6() {
			return "", false
		}
		tail := value[end+1:]
		if tail == "" {
			return fallback, true
		}
		if !strings.HasPrefix(tail, ":") {
			return "", false
		}
		return tail[1:], true
	}
	host, port, explicit := strings.Cut(value, ":")
	address, err := netip.ParseAddr(host)
	if err != nil || !address.Is4() {
		return "", false
	}
	if explicit {
		return port, true
	}
	return fallback, true
}
