package mutation

import (
	"context"
	"encoding/json"
	"time"
)

type setupDaemon struct {
	ID      string    `json:"daemon_id"`
	Type    string    `json:"daemon_type"`
	Service string    `json:"service_name"`
	Host    string    `json:"hostname"`
	Status  *int      `json:"status"`
	Started time.Time `json:"started"`
	Refresh time.Time `json:"last_refresh"`
}

func setupDaemonSnapshot(raw []byte, service string) (map[string]setupDaemon, bool) {
	var rows []setupDaemon
	if json.Unmarshal(raw, &rows) != nil || rows == nil {
		return nil, false
	}
	out := map[string]setupDaemon{}
	for _, d := range rows {
		if d.ID == "" || d.Type != "rgw" || d.Service != service || d.Host == "" || d.Status == nil || d.Started.IsZero() || d.Refresh.IsZero() {
			return nil, false
		}
		if _, exists := out[d.ID]; exists {
			return nil, false
		}
		out[d.ID] = d
	}
	return out, true
}

// A refreshed running state alone can predate the asynchronous restart. Require
// the same daemon identities and a newer native start time for each process.
func setupRestartReady(before, after map[string]setupDaemon) bool {
	if len(before) == 0 || len(before) != len(after) {
		return false
	}
	for id, old := range before {
		d, exists := after[id]
		if !exists || d.Host != old.Host || d.Service != old.Service || d.Status == nil || *d.Status != 1 || !d.Started.After(old.Started) || !d.Refresh.After(old.Refresh) || d.Refresh.Before(d.Started) {
			return false
		}
	}
	return true
}

func waitSetupRestart(ctx context.Context, interval time.Duration, before map[string]setupDaemon, read func() (map[string]setupDaemon, bool)) bool {
	for {
		if ctx.Err() != nil {
			return false
		}
		after, ok := read()
		if !ok || ctx.Err() != nil {
			return false
		}
		if setupRestartReady(before, after) {
			return true
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
	}
}
