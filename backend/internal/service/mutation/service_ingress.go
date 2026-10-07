package mutation

import (
	"crypto/tls"
	"encoding/json"
	"net/netip"
	"regexp"
)

func ingressCreateSpec(p map[string]any, serviceType, action string) (map[string]any, error) {
	fields := []string{"backend_service", "virtual_ip", "frontend_port", "monitor_port", "virtual_interface_networks", "ssl", "ssl_cert", "ssl_key"}
	if serviceType != "ingress" || action != "service.create" {
		for _, key := range fields {
			if _, exists := p[key]; exists {
				return nil, invalid("ingress parameters are only supported when creating an ingress service")
			}
		}
		return nil, nil
	}
	backend, err := required(p, "backend_service")
	if err != nil || !regexp.MustCompile(`^(rgw|nfs)\.[a-zA-Z0-9_.-]+$`).MatchString(backend) {
		return nil, invalid("backend_service must name an RGW or NFS service")
	}
	vip, err := required(p, "virtual_ip")
	if err != nil {
		return nil, err
	}
	if _, err := netip.ParsePrefix(vip); err != nil {
		return nil, invalid("virtual_ip must be an IPv4 or IPv6 address with a prefix length")
	}
	spec := map[string]any{"backend_service": backend, "virtual_ip": vip}
	for _, key := range fields[2:4] {
		encoded, err := json.Marshal(p[key])
		var port *int
		if err != nil || json.Unmarshal(encoded, &port) != nil || port == nil || *port < 1 || *port > 65535 {
			return nil, invalid(key + " must be an integer between 1 and 65535")
		}
		spec[key] = *port
	}
	if spec["frontend_port"] == spec["monitor_port"] {
		return nil, invalid("frontend_port and monitor_port must differ")
	}
	if value, exists := p["virtual_interface_networks"]; exists {
		encoded, err := json.Marshal(value)
		var networks []string
		if err != nil || json.Unmarshal(encoded, &networks) != nil || networks == nil {
			return nil, invalid("virtual_interface_networks must be an array of CIDR strings")
		}
		for _, network := range networks {
			if _, err := netip.ParsePrefix(network); err != nil {
				return nil, invalid("virtual_interface_networks entries must be IPv4 or IPv6 CIDRs")
			}
		}
		spec["virtual_interface_networks"] = networks
	}
	if value, exists := p["ssl"]; exists {
		enabled, ok := value.(bool)
		if !ok {
			return nil, invalid("ssl must be a boolean")
		}
		spec["ssl"] = enabled
		if enabled {
			cert, certOK := p["ssl_cert"].(string)
			key, keyOK := p["ssl_key"].(string)
			if !certOK || !keyOK || len(cert) == 0 || len(key) == 0 || len(cert) > 64<<10 || len(key) > 64<<10 {
				return nil, invalid("TLS requires a PEM certificate and private key of at most 64 KiB each")
			}
			if _, err := tls.X509KeyPair([]byte(cert), []byte(key)); err != nil {
				return nil, invalid("TLS certificate and private key must be valid matching PEM data")
			}
			spec["ssl_cert"], spec["ssl_key"] = cert, key
			return spec, nil
		}
	}
	for _, field := range []string{"ssl_cert", "ssl_key"} {
		if _, exists := p[field]; exists {
			return nil, invalid("TLS material requires ssl=true")
		}
	}
	return spec, nil
}
