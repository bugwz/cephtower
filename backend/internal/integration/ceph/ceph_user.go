package ceph

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
)

func (p *NativeProvider) collectCephUsers(ctx context.Context, access ClusterAccess) ([]Observation, error) {
	// Decode only public metadata. auth ls also returns keys, which must never be
	// persisted in the resource store or returned by the list endpoint.
	var payload struct {
		Users []struct {
			Entity string            `json:"entity"`
			Caps   map[string]string `json:"caps"`
		} `json:"auth_dump"`
	}
	if err := p.runInto(ctx, access, "collect.ceph_user", []string{"auth", "ls", "--format", "json"}, &payload); err != nil {
		return nil, err
	}
	if payload.Users == nil {
		return nil, fmt.Errorf("parse auth ls response: auth_dump is required")
	}
	now := time.Now().UTC()
	rows := make([]Observation, 0, len(payload.Users)*2)
	for _, user := range payload.Users {
		if user.Entity == "" || user.Caps == nil {
			return nil, fmt.Errorf("parse auth ls response: entity and caps are required")
		}
		kind, _, _ := strings.Cut(user.Entity, ".")
		rows = append(rows, observation("ceph_user", user.Entity, user.Entity, "ceph_cli", map[string]any{"entity": user.Entity, "entity_type": kind, "caps": user.Caps}, now))
		if kind == "client" {
			rows = append(rows, cephFSAuthorizationObservations(user.Entity, user.Caps["mds"], now)...)
		}
	}
	return rows, nil
}

type cephFSAuthorization struct {
	Filesystem string
	Path       string
	Permission string
	RootSquash bool
	UID        string
	GIDs       string
	Network    string
	Raw        string
}

func cephFSAuthorizationObservations(client, capability string, now time.Time) []Observation {
	grants := parseMDSAuthorizationGrants(capability)
	rows := make([]Observation, 0, len(grants))
	for _, grant := range grants {
		access := "r"
		if grant.Permission == "*" || grant.Permission == "all" {
			access = "*"
		} else if strings.Contains(grant.Permission, "w") {
			access = "rw"
		}
		payload := map[string]any{
			"fs":          grant.Filesystem,
			"client":      client,
			"path":        grant.Path,
			"access":      access,
			"permissions": grant.Permission,
			"quota":       grant.Permission == "*" || grant.Permission == "all" || strings.Contains(grant.Permission, "p"),
			"snapshot":    grant.Permission == "*" || grant.Permission == "all" || strings.Contains(grant.Permission, "s"),
			"full":        grant.Permission == "*" || grant.Permission == "all" || strings.Contains(grant.Permission, "f"),
			"root_squash": grant.RootSquash,
			"mds_cap":     grant.Raw,
		}
		if grant.UID != "" {
			payload["uid"] = grant.UID
		}
		if grant.GIDs != "" {
			payload["gids"] = grant.GIDs
		}
		if grant.Network != "" {
			payload["network"] = grant.Network
		}
		key := opaquePair(grant.Filesystem, client+"\x00"+grant.Path)
		rows = append(rows, Observation{
			Kind:       "cephfs_authorization",
			NaturalKey: key,
			ParentKind: "filesystem",
			ParentKey:  grant.Filesystem,
			Name:       client + " " + grant.Path,
			Status:     "available",
			Source:     "ceph_cli",
			Payload:    payload,
			ObservedAt: now,
		})
	}
	return rows
}

func parseMDSAuthorizationGrants(capability string) []cephFSAuthorization {
	var grants []cephFSAuthorization
	for _, raw := range splitMDSAuthorizationGrants(capability) {
		fields, ok := capabilityFields(raw)
		if !ok || len(fields) < 2 || fields[0] != "allow" || !validMDSPermission(fields[1]) {
			continue
		}
		grant := cephFSAuthorization{Filesystem: "*", Path: "/", Permission: fields[1], Raw: strings.TrimSpace(raw)}
		for index := 2; index < len(fields); index++ {
			field := fields[index]
			switch {
			case strings.HasPrefix(field, "fsname="):
				grant.Filesystem = strings.TrimPrefix(field, "fsname=")
			case strings.HasPrefix(field, "path="):
				grant.Path = strings.TrimPrefix(field, "path=")
			case field == "root_squash":
				grant.RootSquash = true
			case strings.HasPrefix(field, "uid="):
				grant.UID = strings.TrimPrefix(field, "uid=")
			case strings.HasPrefix(field, "gids="):
				grant.GIDs = strings.TrimPrefix(field, "gids=")
			case field == "network" && index+1 < len(fields):
				index++
				grant.Network = fields[index]
			}
		}
		if grant.Filesystem == "" {
			grant.Filesystem = "*"
		}
		grant.Path = strings.TrimSpace(grant.Path)
		if grant.Path == "" {
			grant.Path = "/"
		} else if !strings.HasPrefix(grant.Path, "/") {
			grant.Path = "/" + grant.Path
		}
		grants = append(grants, grant)
	}
	return grants
}

func validMDSPermission(value string) bool {
	if value == "*" || value == "all" {
		return true
	}
	if !strings.HasPrefix(value, "r") {
		return false
	}
	for _, letter := range value[1:] {
		if !slices.Contains([]rune{'w', 'f', 'p', 's'}, letter) {
			return false
		}
	}
	return true
}

func splitMDSAuthorizationGrants(capability string) []string {
	var grants []string
	start := 0
	var quote byte
	for index := 0; index < len(capability); index++ {
		char := capability[index]
		if quote != 0 {
			if char == quote && (index == 0 || capability[index-1] != '\\') {
				quote = 0
			}
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			continue
		}
		if char != ',' && char != ';' {
			continue
		}
		next := strings.TrimLeft(capability[index+1:], " \t\r\n")
		if next != "allow" && !strings.HasPrefix(next, "allow ") && !strings.HasPrefix(next, "allow\t") && !strings.HasPrefix(next, "allow\n") {
			continue
		}
		if grant := strings.TrimSpace(capability[start:index]); grant != "" {
			grants = append(grants, grant)
		}
		start = index + 1
	}
	if grant := strings.TrimSpace(capability[start:]); grant != "" {
		grants = append(grants, grant)
	}
	return grants
}

func capabilityFields(value string) ([]string, bool) {
	var fields []string
	var current strings.Builder
	var quote byte
	escaped := false
	flush := func() {
		if current.Len() > 0 {
			fields = append(fields, current.String())
			current.Reset()
		}
	}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if escaped {
			current.WriteByte(char)
			escaped = false
			continue
		}
		if quote != 0 {
			if char == '\\' {
				escaped = true
			} else if char == quote {
				quote = 0
			} else {
				current.WriteByte(char)
			}
			continue
		}
		switch char {
		case '\'', '"':
			quote = char
		case ' ', '\t', '\r', '\n':
			flush()
		default:
			current.WriteByte(char)
		}
	}
	if quote != 0 || escaped {
		return nil, false
	}
	flush()
	return fields, true
}
