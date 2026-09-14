package ceph

import (
	"context"
	"fmt"
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
	rows := make([]Observation, 0, len(payload.Users))
	for _, user := range payload.Users {
		if user.Entity == "" || user.Caps == nil {
			return nil, fmt.Errorf("parse auth ls response: entity and caps are required")
		}
		kind, _, _ := strings.Cut(user.Entity, ".")
		rows = append(rows, observation("ceph_user", user.Entity, user.Entity, "ceph_cli", map[string]any{"entity": user.Entity, "entity_type": kind, "caps": user.Caps}, now))
	}
	return rows, nil
}
