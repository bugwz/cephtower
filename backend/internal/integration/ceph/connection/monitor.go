package connection

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const MaxMonitorAddressesBytes = 4096

type MonitorEndpoint struct {
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     uint16 `json:"port"`
	Nonce    uint64 `json:"nonce,omitempty"`
	Vector   int    `json:"vector"`
}

// ParseMonitorAddresses parses Ceph mon_host without splitting commas inside addrvecs.
func ParseMonitorAddresses(value string) ([]MonitorEndpoint, error) {
	if value == "" || strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("monitor addresses are required")
	}
	if value != strings.TrimSpace(value) {
		value = strings.TrimSpace(value)
	}
	if len(value) > MaxMonitorAddressesBytes {
		return nil, fmt.Errorf("monitor addresses exceed %d bytes", MaxMonitorAddressesBytes)
	}
	groups, err := splitGroups(value)
	if err != nil {
		return nil, err
	}
	var result []MonitorEndpoint
	for vector, group := range groups {
		items := []string{group}
		if strings.HasPrefix(group, "[") {
			items = strings.Split(group[1:len(group)-1], ",")
		}
		for _, item := range items {
			endpoint, err := parseEndpoint(strings.TrimSpace(item), vector)
			if err != nil {
				return nil, err
			}
			result = append(result, endpoint)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("monitor addresses contain no endpoints")
	}
	return result, nil
}

func splitGroups(value string) ([]string, error) {
	var groups []string
	for i := 0; i < len(value); {
		for i < len(value) && (value[i] == ' ' || value[i] == '\t' || value[i] == ',') {
			i++
		}
		if i == len(value) {
			break
		}
		start := i
		if value[i] == '[' {
			depth := 1
			i++
			for i < len(value) && depth > 0 {
				if value[i] == '[' {
					depth++
				}
				if value[i] == ']' {
					depth--
				}
				i++
			}
			if depth != 0 {
				return nil, fmt.Errorf("unbalanced monitor addrvec brackets")
			}
			groups = append(groups, value[start:i])
			continue
		}
		for i < len(value) && value[i] != ' ' && value[i] != '\t' && value[i] != ',' {
			i++
		}
		if strings.Contains(value[start:i], "]") {
			return nil, fmt.Errorf("unbalanced monitor addrvec brackets")
		}
		groups = append(groups, value[start:i])
	}
	return groups, nil
}

func parseEndpoint(value string, vector int) (MonitorEndpoint, error) {
	protocol := "v1"
	if strings.HasPrefix(value, "v1:") || strings.HasPrefix(value, "v2:") {
		protocol = value[:2]
		value = value[3:]
	}
	nonce := uint64(0)
	if slash := strings.LastIndex(value, "/"); slash >= 0 {
		parsed, err := strconv.ParseUint(value[slash+1:], 10, 64)
		if err != nil {
			return MonitorEndpoint{}, fmt.Errorf("invalid monitor nonce")
		}
		nonce = parsed
		value = value[:slash]
	}
	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		colon := strings.LastIndex(value, ":")
		if colon <= 0 || strings.Contains(value[:colon], ":") {
			return MonitorEndpoint{}, fmt.Errorf("invalid monitor endpoint %q", value)
		}
		host, portText = value[:colon], value[colon+1:]
	}
	host = strings.Trim(host, "[]")
	if host == "" || strings.ContainsAny(host, " \t/[]") {
		return MonitorEndpoint{}, fmt.Errorf("invalid monitor host")
	}
	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port == 0 {
		return MonitorEndpoint{}, fmt.Errorf("invalid monitor port")
	}
	return MonitorEndpoint{Protocol: protocol, Host: host, Port: uint16(port), Nonce: nonce, Vector: vector}, nil
}
