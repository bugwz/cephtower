package mutation

import (
	"encoding/json"
	"net/netip"
	"regexp"
)

func ingressCreateSpec(p map[string]any, serviceType, action string) (map[string]any, error) {
	fields := []string{"backend_service", "virtual_ip", "frontend_port", "monitor_port"}
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
	for _, key := range fields[2:] {
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
	return spec, nil
}
