package mutation

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
)

func nfsClusterIngressArgs(p map[string]any) ([]string, error) {
	enabled := false
	if value, exists := p["ingress"]; exists {
		var ok bool
		enabled, ok = value.(bool)
		if !ok {
			return nil, invalid("ingress must be a boolean")
		}
	}
	args := []string{}
	vip, mode := "", ""
	for _, key := range []string{"virtual_ip", "ingress_mode"} {
		if value, exists := p[key]; exists {
			text, ok := value.(string)
			if !ok || text == "" || !enabled {
				return nil, invalid(key + " requires ingress and a non-empty string")
			}
			if key == "virtual_ip" {
				vip = text
			} else {
				mode = text
			}
		}
	}
	if enabled {
		if vip == "" {
			return nil, invalid("virtual_ip is required with ingress")
		}
		if strings.Contains(vip, "/") {
			if _, _, err := net.ParseCIDR(vip); err != nil {
				return nil, invalid("virtual_ip must be an IP address or CIDR")
			}
		} else if net.ParseIP(vip) == nil {
			return nil, invalid("virtual_ip must be an IP address or CIDR")
		}
		args = append(args, "--ingress", "--virtual-ip="+vip)
		if mode != "" {
			switch mode {
			case "default", "keepalive-only", "haproxy-standard", "haproxy-protocol":
			default:
				return nil, invalid("invalid ingress_mode")
			}
			args = append(args, "--ingress-mode="+mode)
		}
	}
	if value, exists := p["nfs_port"]; exists {
		data, err := json.Marshal(value)
		var port int
		if err != nil || json.Unmarshal(data, &port) != nil || port < 1 || port > 65535 {
			return nil, invalid("nfs_port must be an integer from 1 to 65535")
		}
		limit := 65535
		if enabled {
			limit = 55535
			if mode == "keepalive-only" {
				limit = 58535
			}
		}
		if port > limit {
			return nil, invalid("nfs_port exceeds the range for derived ingress ports")
		}
		args = append(args, "--port="+strconv.Itoa(port))
	}
	return args, nil
}
