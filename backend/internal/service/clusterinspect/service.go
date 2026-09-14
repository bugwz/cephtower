package clusterinspect

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
)

type Service struct {
	clusters *clusterservice.Service
	executor executor.Executor
}

func New(clusters *clusterservice.Service, runner executor.Executor) *Service {
	return &Service{clusters, runner}
}

type LogEntry struct {
	Name     string          `json:"name"`
	Rank     string          `json:"rank"`
	Stamp    string          `json:"stamp"`
	Seq      logSequence     `json:"seq"`
	Channel  string          `json:"channel"`
	Priority string          `json:"priority"`
	Message  string          `json:"message"`
	Addrs    json.RawMessage `json:"addrs"`
}

// Log sequences can exceed JavaScript's safe integer range. Emit them as strings.
type logSequence string

func (s *logSequence) UnmarshalJSON(data []byte) error {
	value := string(data)
	if _, err := strconv.ParseUint(value, 10, 64); err != nil {
		return err
	}
	*s = logSequence(value)
	return nil
}
func (s logSequence) String() string { return string(s) }

type Logs struct {
	Items      []LogEntry `json:"items"`
	ObservedAt time.Time  `json:"observed_at"`
}

func (s *Service) Logs(ctx context.Context, clusterID uint64, channel, level string, limit int) (Logs, error) {
	if channel == "" {
		channel = "cluster"
	}
	if level == "" {
		level = "debug"
	}
	if limit == 0 {
		limit = 100
	}
	if !oneOf(channel, "cluster", "audit", "cephadm", "*") || !oneOf(level, "debug", "info", "sec", "warn", "error") || limit < 1 || limit > 500 {
		return Logs{}, invalid("invalid log channel, level or limit (1–500)")
	}
	var rows []LogEntry
	if err := s.read(ctx, clusterID, "cluster.logs", []string{"log", "last", strconv.Itoa(limit), level, channel, "--format", "json"}, &rows); err != nil {
		return Logs{}, err
	}
	if rows == nil {
		return Logs{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph log response must be an array"}
	}
	for i := range rows {
		rows[i].Message = security.Redact(rows[i].Message)
	}
	// Ceph returns oldest first; Dashboard renders the latest record first.
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return Logs{Items: rows, ObservedAt: time.Now().UTC()}, nil
}

var optionName = regexp.MustCompile(`^(mgr/[A-Za-z][A-Za-z0-9_]*/)?[A-Za-z][A-Za-z0-9_]{0,255}$`)

func (s *Service) OSDInspection(ctx context.Context, clusterID uint64, id, section string) (map[string]any, error) {
	if !regexp.MustCompile(`^(0|[1-9][0-9]{0,9})$`).MatchString(id) {
		return nil, invalid("invalid OSD id")
	}
	var args []string
	switch section {
	case "metadata":
		args = []string{"osd", "metadata", id, "--format", "json"}
	case "histogram":
		args = []string{"tell", "osd." + id, "perf", "histogram", "dump", "--format", "json"}
	default:
		return nil, invalid("invalid OSD inspection section")
	}
	var result map[string]any
	if err := s.read(ctx, clusterID, "osd."+section, args, &result); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned empty OSD diagnostics"}
	}
	return result, nil
}

func (s *Service) ConfigurationOption(ctx context.Context, clusterID uint64, name string) (map[string]any, error) {
	if !optionName.MatchString(name) {
		return nil, invalid("invalid configuration option name")
	}
	var option map[string]any
	if err := s.read(ctx, clusterID, "configuration.help", []string{"config", "help", name, "--format", "json"}, &option); err != nil {
		return nil, err
	}
	if option["name"] != name {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an unexpected configuration option"}
	}
	return option, nil
}
func (s *Service) read(ctx context.Context, clusterID uint64, id string, args []string, out any) error {
	if clusterID == 0 {
		return invalid("cluster_id is required")
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return err
	}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: id, Binary: executor.BinaryCeph, Args: args, Timeout: 20 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		return &cephdomain.ActionError{Code: "ceph_command_failed", Message: security.Redact(err.Error()), Retryable: true}
	}
	decoder := json.NewDecoder(bytes.NewReader(result.Stdout))
	decoder.UseNumber()
	if err := decoder.Decode(out); err != nil {
		return &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned invalid JSON"}
	}
	return nil
}
func oneOf(value string, values ...string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
func invalid(message string) error {
	return &cephdomain.ActionError{Code: "invalid_request", Message: strings.TrimSpace(message)}
}

func (s *Service) SnapshotSchedules(ctx context.Context, clusterID uint64, fs, path, subvol, group string) ([]map[string]any, error) {
	if fs == "" || !strings.HasPrefix(path, "/") {
		return nil, invalid("filesystem and absolute path are required")
	}
	args := []string{"fs", "snap-schedule", "status", path, "--fs=" + fs, "--format=json"}
	for _, value := range []string{fs, path, subvol, group} {
		if strings.ContainsAny(value, "\x00\r\n") {
			return nil, invalid("invalid schedule scope")
		}
	}
	if group != "" && subvol == "" {
		return nil, invalid("subvolume is required with group")
	}
	if subvol != "" {
		args = append(args, "--subvol="+subvol)
	}
	if group != "" {
		args = append(args, "--group="+group)
	}
	var rows []map[string]any
	if err := s.read(ctx, clusterID, "snapshot_schedule.status", args, &rows); err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned null schedules"}
	}
	for _, row := range rows {
		path, pathOK := row["path"].(string)
		schedule, scheduleOK := row["schedule"].(string)
		start, startOK := row["start"].(string)
		_, activeOK := row["active"].(bool)
		if !pathOK || path == "" || !scheduleOK || schedule == "" || !startOK || start == "" || !activeOK {
			return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned incomplete schedule identity or state"}
		}
	}
	return rows, nil
}
