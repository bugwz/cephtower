package clusterinspect

import (
	"context"
	"errors"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
)

const maxCephFSDirectories = 500

var (
	cephFSNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.][A-Za-z0-9_.-]{0,254}$`)
	cephFSLongLine    = regexp.MustCompile(`^([dl-][rwxStTs-]{9})\s+([0-9]+)\s+([0-9]+)\s+([0-9]+)\s+(\S+)\s+(\S+)\s+(.+)$`)
)

type CephFSQuota struct {
	MaxBytes int64 `json:"max_bytes"`
	MaxFiles int64 `json:"max_files"`
}

type CephFSDirectory struct {
	Name       string       `json:"name"`
	Path       string       `json:"path"`
	Parent     *string      `json:"parent"`
	Mode       string       `json:"mode"`
	Size       int64        `json:"size"`
	UID        uint64       `json:"uid"`
	GID        uint64       `json:"gid"`
	ModifiedAt string       `json:"modified_at"`
	Quotas     *CephFSQuota `json:"quotas"`
}

type CephFSDirectoryList struct {
	Filesystem string            `json:"filesystem"`
	Path       string            `json:"path"`
	Items      []CephFSDirectory `json:"items"`
	ObservedAt time.Time         `json:"observed_at"`
}

func (s *Service) CephFSDirectories(ctx context.Context, clusterID uint64, filesystem, requestedPath string) (CephFSDirectoryList, error) {
	filesystem = strings.TrimSpace(filesystem)
	requestedPath = strings.TrimSpace(requestedPath)
	if !cephFSNamePattern.MatchString(filesystem) {
		return CephFSDirectoryList{}, invalid("filesystem is required or invalid")
	}
	if requestedPath == "" {
		requestedPath = "/"
	}
	if !strings.HasPrefix(requestedPath, "/") || len(requestedPath) > 32<<10 || strings.ContainsAny(requestedPath, "\x00\r\n,") {
		return CephFSDirectoryList{}, invalid("path must be absolute and cannot contain commas or newlines")
	}
	requestedPath = path.Clean(requestedPath)
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return CephFSDirectoryList{}, err
	}
	defer func() { access.ClientKey = "" }()
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: "cephfs.directory.list", Binary: executor.BinaryCephFSShell,
		Args:    []string{"--fs", filesystem, "ls", "-la", quoteCephFSShellToken(requestedPath)},
		Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput,
	})
	if err != nil {
		return CephFSDirectoryList{}, cephFSReadError(err)
	}
	children, err := parseCephFSDirectories(requestedPath, result.Stdout)
	if err != nil {
		return CephFSDirectoryList{}, err
	}
	items := make([]CephFSDirectory, 0, len(children)+1)
	current := CephFSDirectory{Name: path.Base(requestedPath), Path: requestedPath}
	if requestedPath == "/" {
		current.Name = "/"
	} else {
		parent := path.Dir(requestedPath)
		current.Parent = &parent
		quota, quotaErr := s.cephFSQuota(ctx, access, filesystem, requestedPath)
		if quotaErr != nil {
			return CephFSDirectoryList{}, quotaErr
		}
		current.Quotas = &quota
	}
	items = append(items, current)
	if err := s.populateCephFSQuotas(ctx, access, filesystem, children); err != nil {
		return CephFSDirectoryList{}, err
	}
	items = append(items, children...)
	return CephFSDirectoryList{Filesystem: filesystem, Path: requestedPath, Items: items, ObservedAt: time.Now().UTC()}, nil
}

func (s *Service) populateCephFSQuotas(ctx context.Context, access executor.ClusterAccess, filesystem string, directories []CephFSDirectory) error {
	if len(directories) == 0 {
		return nil
	}
	workerCount := min(8, len(directories))
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var next atomic.Int64
	var workers sync.WaitGroup
	errCh := make(chan error, 1)
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for {
				index := int(next.Add(1) - 1)
				if index >= len(directories) || workerCtx.Err() != nil {
					return
				}
				quota, err := s.cephFSQuota(workerCtx, access, filesystem, directories[index].Path)
				if err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
					return
				}
				directories[index].Quotas = &quota
			}
		}()
	}
	workers.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (s *Service) cephFSQuota(ctx context.Context, access executor.ClusterAccess, filesystem, directory string) (CephFSQuota, error) {
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: "cephfs.directory.quota", Binary: executor.BinaryCephFSShell,
		Args:    []string{"--fs", filesystem, "quota", "get", quoteCephFSShellToken(directory)},
		Timeout: 20 * time.Second, MaxOutput: executor.DefaultMaxOutput,
	})
	if err != nil {
		var commandErr *executor.Error
		if !errors.As(err, &commandErr) || commandErr.Kind != "exit" || commandErr.ExitCode != 9 {
			return CephFSQuota{}, cephFSReadError(err)
		}
	}
	quota, parseErr := parseCephFSQuota(result.Stdout)
	if parseErr != nil {
		return CephFSQuota{}, parseErr
	}
	return quota, nil
}

func parseCephFSDirectories(parent string, data []byte) ([]CephFSDirectory, error) {
	rows := make([]CephFSDirectory, 0)
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSuffix(raw, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		match := cephFSLongLine.FindStringSubmatch(line)
		if match == nil {
			return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "cephfs-shell returned an unrecognized directory listing"}
		}
		if match[1][0] != 'd' {
			continue
		}
		name := strings.TrimSuffix(match[7], "/")
		if name == "." || name == ".." || name == "" {
			continue
		}
		if len(rows) >= maxCephFSDirectories {
			return nil, &cephdomain.ActionError{Code: "response_too_large", Message: "directory contains more than 500 child directories"}
		}
		size, sizeErr := strconv.ParseInt(match[2], 10, 64)
		uid, uidErr := strconv.ParseUint(match[3], 10, 64)
		gid, gidErr := strconv.ParseUint(match[4], 10, 64)
		if sizeErr != nil || uidErr != nil || gidErr != nil {
			return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "cephfs-shell returned invalid directory metadata"}
		}
		parentCopy := parent
		rows = append(rows, CephFSDirectory{
			Name: name, Path: path.Join(parent, name), Parent: &parentCopy, Mode: match[1],
			Size: size, UID: uid, GID: gid, ModifiedAt: match[5] + " " + match[6],
		})
	}
	return rows, nil
}

func parseCephFSQuota(data []byte) (CephFSQuota, error) {
	quota := CephFSQuota{}
	for _, raw := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		parts := strings.SplitN(strings.TrimSpace(raw), ":", 2)
		if len(parts) != 2 || (parts[0] != "max_bytes" && parts[0] != "max_files") {
			if strings.TrimSpace(raw) == "" {
				continue
			}
			return CephFSQuota{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "cephfs-shell returned an unrecognized quota response"}
		}
		value, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil || value < 0 {
			return CephFSQuota{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "cephfs-shell returned an invalid quota value"}
		}
		if parts[0] == "max_bytes" {
			quota.MaxBytes = value
		} else {
			quota.MaxFiles = value
		}
	}
	return quota, nil
}

func quoteCephFSShellToken(value string) string {
	if strings.ContainsAny(value, " \t'\"\\") {
		return strconv.Quote(value)
	}
	return value
}

func cephFSReadError(err error) error {
	return &cephdomain.ActionError{Code: "ceph_command_failed", Message: security.Redact(err.Error()), Retryable: true}
}
