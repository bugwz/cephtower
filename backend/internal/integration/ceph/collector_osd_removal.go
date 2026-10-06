package ceph

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func osdRemovalRows(raw []byte, now time.Time) ([]Observation, bool) {
	// Orchestrator returns this text before applying the requested output format.
	if strings.TrimSpace(string(raw)) == "No OSD remove/replace operations reported" {
		return []Observation{}, true
	}
	var items []map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&items) != nil || items == nil {
		return nil, false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, false
	}
	rows := make([]Observation, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		id, ok := item["osd_id"].(json.Number)
		n, err := strconv.ParseUint(string(id), 10, 31)
		key := strconv.FormatUint(n, 10)
		if !ok || err != nil || key != string(id) || seen[key] {
			return nil, false
		}
		seen[key] = true
		rows = append(rows, observation("osd_removal", key, key, "ceph_cli", item, now))
	}
	return rows, true
}

func (p *NativeProvider) collectOSDRemovals(ctx context.Context, access ClusterAccess, now time.Time) []Observation {
	const id = "collect.osd_removal"
	result, err := p.Executor.Run(ctx, access, executor.CommandSpec{ID: id, Binary: executor.BinaryCeph, Args: []string{"orch", "osd", "rm", "status", "--format", "json"}, Timeout: 45 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	if err != nil || result.ExitCode != 0 {
		markCollectionUnavailable(ctx, id)
		return nil
	}
	rows, valid := osdRemovalRows(result.Stdout, now)
	if !valid {
		markCollectionUnavailable(ctx, id)
		return nil
	}
	return rows
}
