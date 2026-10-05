package mutation

import (
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var setupHostName = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,252}$`)

// Replace only the URL authority. A hostname occurring in an URL path must not
// be replaced, and IPv6 addresses must be bracketed before restoring the port.
func resolveRealmSetupEndpoints(raw []byte, parameters map[string]any) (map[string]any, error) {
	fail := func() (map[string]any, error) {
		return nil, invalid("orchestrator host addresses or resolved gateway endpoints could not be verified")
	}
	var hosts []map[string]any
	if json.Unmarshal(raw, &hosts) != nil || hosts == nil {
		return fail()
	}
	addresses := map[string]string{}
	for _, host := range hosts {
		name, addr := rawText(host, "hostname"), rawText(host, "addr")
		name = strings.ToLower(name)
		if !setupHostName.MatchString(name) || addresses[name] != "" {
			return fail()
		}
		// HostSpec addresses may be IPv4, IPv6 or DNS names, but not host:port
		// pairs: the gateway port is taken from the submitted endpoint.
		if strings.HasPrefix(addr, "[") && strings.HasSuffix(addr, "]") {
			addr = strings.TrimSuffix(strings.TrimPrefix(addr, "["), "]")
			if net.ParseIP(addr) == nil || !strings.Contains(addr, ":") {
				return fail()
			}
		}
		if net.ParseIP(addr) == nil && !setupHostName.MatchString(addr) {
			return fail()
		}
		addresses[name] = addr
	}
	resolved := make(map[string]any, len(parameters))
	for key, value := range parameters {
		resolved[key] = value
	}
	for _, field := range []string{"zonegroup_endpoints", "zone_endpoints"} {
		values, ok := realmSetupStrings(parameters[field])
		if !ok {
			return fail()
		}
		seen := map[string]bool{}
		for i, value := range values {
			u, err := url.Parse(value)
			if err != nil || u.Hostname() == "" {
				return fail()
			}
			if addr, exists := addresses[strings.ToLower(u.Hostname())]; exists {
				port := u.Port()
				if port != "" {
					u.Host = net.JoinHostPort(addr, port)
				} else if strings.Contains(addr, ":") {
					u.Host = "[" + addr + "]"
				} else {
					u.Host = addr
				}
				values[i] = u.String()
			}
			if seen[values[i]] {
				return fail()
			}
			seen[values[i]] = true
		}
		resolved[field] = values
	}
	if _, err := buildRealmSetup(resolved); err != nil {
		return fail()
	}
	return resolved, nil
}
