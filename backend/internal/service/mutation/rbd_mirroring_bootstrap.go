package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

const (
	maxRBDMirrorBootstrapToken  = 1023
	maxRBDMirrorBootstrapOutput = 64 << 10
)

var rbdPoolNamePattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

// CreateRBDMirrorBootstrapToken returns the secret directly to the caller. It is
// intentionally not dispatched through the durable operation queue, which would
// persist the token in operation details.
func (s *Service) CreateRBDMirrorBootstrapToken(ctx context.Context, clusterID uint64, pool, siteName string) (string, error) {
	pool, siteName, err := validateRBDMirrorBootstrapTarget(clusterID, pool, siteName)
	if err != nil {
		return "", err
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return "", err
	}
	defer func() { access.ClientKey = "" }()
	if err := s.ensureRBDMirroringEnabled(ctx, access, pool); err != nil {
		return "", err
	}
	args := []string{"mirror", "pool", "peer", "bootstrap", "create", pool, "--site-name=" + siteName}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: "rbd_mirroring.bootstrap.create", Binary: executor.BinaryRBD, Args: args,
		Timeout: 2 * time.Minute, MaxOutput: maxRBDMirrorBootstrapOutput, Mutating: true,
	})
	if err != nil {
		return "", &cephdomain.ActionError{Code: "ceph_command_failed", Message: "RBD mirroring bootstrap token creation failed"}
	}
	token := strings.TrimSpace(string(result.Stdout))
	if err := validateRBDMirrorBootstrapToken(token); err != nil {
		return "", &cephdomain.ActionError{Code: "ceph_command_failed", Message: "RBD returned an invalid mirroring bootstrap token"}
	}
	return token, nil
}

// ImportRBDMirrorBootstrapToken passes the secret on stdin so it never appears
// in process arguments, logs, cached resource state, or durable operations.
func (s *Service) ImportRBDMirrorBootstrapToken(ctx context.Context, clusterID uint64, pool, siteName, direction, token string) error {
	pool, siteName, err := validateRBDMirrorBootstrapTarget(clusterID, pool, siteName)
	if err != nil {
		return err
	}
	direction = strings.TrimSpace(direction)
	if direction != "rx-only" && direction != "rx-tx" {
		return invalid("direction must be rx-only or rx-tx")
	}
	token = strings.TrimSpace(token)
	if err := validateRBDMirrorBootstrapToken(token); err != nil {
		return err
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return err
	}
	defer func() { access.ClientKey = "" }()
	if err := s.ensureRBDMirroringEnabled(ctx, access, pool); err != nil {
		return err
	}
	stdin := append([]byte(token), '\n')
	defer func() {
		for index := range stdin {
			stdin[index] = 0
		}
	}()
	args := []string{"mirror", "pool", "peer", "bootstrap", "import", pool, "-", "--site-name=" + siteName, "--direction=" + direction}
	if _, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: "rbd_mirroring.bootstrap.import", Binary: executor.BinaryRBD, Args: args, Stdin: stdin,
		Timeout: 5 * time.Minute, MaxOutput: executor.DefaultMaxOutput, Mutating: true,
	}); err != nil {
		return &cephdomain.ActionError{Code: "ceph_command_failed", Message: "RBD mirroring bootstrap token import failed"}
	}
	if _, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: "rbd_mirroring.bootstrap.import.post_check", Binary: executor.BinaryRBD,
		Args:    []string{"mirror", "pool", "info", pool, "--format", "json"},
		Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput,
	}); err != nil {
		return &cephdomain.ActionError{Code: "post_check_failed", Message: "bootstrap token was imported but the peer state could not be verified", Retryable: true}
	}
	return nil
}

func (s *Service) ensureRBDMirroringEnabled(ctx context.Context, access executor.ClusterAccess, pool string) error {
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: "rbd_mirroring.bootstrap.pool_info", Binary: executor.BinaryRBD,
		Args:    []string{"mirror", "pool", "info", pool, "--format", "json"},
		Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput,
	})
	if err != nil {
		return normalize(err)
	}
	var info struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(result.Stdout, &info); err != nil {
		return &cephdomain.ActionError{Code: "ceph_command_failed", Message: "RBD mirroring pool information is invalid"}
	}
	switch info.Mode {
	case "image", "pool", "init-only":
		return nil
	case "disabled":
		_, err = s.executor.Run(ctx, access, executor.CommandSpec{
			ID: "rbd_mirroring.bootstrap.enable", Binary: executor.BinaryRBD,
			Args:    []string{"mirror", "pool", "enable", pool, "image"},
			Timeout: 2 * time.Minute, MaxOutput: executor.DefaultMaxOutput, Mutating: true,
		})
		if err != nil {
			return normalize(err)
		}
		return nil
	default:
		return &cephdomain.ActionError{Code: "ceph_command_failed", Message: "RBD mirroring pool mode is missing or unsupported"}
	}
}

func validateRBDMirrorBootstrapTarget(clusterID uint64, pool, siteName string) (string, string, error) {
	pool = strings.TrimSpace(pool)
	siteName = strings.TrimSpace(siteName)
	if clusterID == 0 {
		return "", "", invalid("cluster is required")
	}
	if !rbdPoolNamePattern.MatchString(pool) {
		return "", "", invalid("pool is invalid")
	}
	if siteName == "" || len(siteName) > 128 || strings.ContainsAny(siteName, "\x00\r\n") {
		return "", "", invalid("site_name must be non-empty text of at most 128 characters")
	}
	return pool, siteName, nil
}

func validateRBDMirrorBootstrapToken(token string) error {
	if token == "" || len(token) > maxRBDMirrorBootstrapToken || strings.ContainsRune(token, 0) {
		return invalid("token must be a valid RBD mirroring bootstrap token")
	}
	var decoded []byte
	var err error
	for _, encoding := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding} {
		decoded, err = encoding.DecodeString(token)
		if err == nil {
			break
		}
	}
	if err != nil || len(decoded) == 0 {
		return invalid("token must be a valid RBD mirroring bootstrap token")
	}
	defer func() {
		for index := range decoded {
			decoded[index] = 0
		}
	}()
	var payload struct {
		FSID     string `json:"fsid"`
		ClientID string `json:"client_id"`
		Key      string `json:"key"`
		MonHost  string `json:"mon_host"`
	}
	if err := json.Unmarshal(decoded, &payload); err != nil || payload.FSID == "" || payload.ClientID == "" || payload.Key == "" || payload.MonHost == "" {
		return invalid("token must be a valid RBD mirroring bootstrap token")
	}
	return nil
}
