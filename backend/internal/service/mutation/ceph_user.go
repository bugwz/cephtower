package mutation

import (
	"context"
	"regexp"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

var cephEntityPattern = regexp.MustCompile(`^(client|mon|mgr|osd|mds)\.[A-Za-z0-9_.:@+\-]{1,256}$`)

func cephUserCommand(request Request, p map[string]any) (command, error) {
	spec := command{binary: executor.BinaryCeph, timeout: 30 * time.Second}
	if request.Action == "ceph_user.import" {
		keyring, ok := p["keyring"].(string)
		if !ok || strings.TrimSpace(keyring) == "" || len(keyring) > 256<<10 || strings.ContainsRune(keyring, 0) {
			return command{}, invalid("keyring must be non-empty text of at most 256 KiB")
		}
		spec.args = []string{"auth", "import", "-i", "-"}
		spec.stdin = []byte(keyring)
		return spec, nil
	}
	entity, ok := p["entity"].(string)
	if !ok || !cephEntityPattern.MatchString(entity) {
		return command{}, invalid("a valid Ceph entity is required")
	}
	switch request.Action {
	case "ceph_user.create", "ceph_user.update":
		action := "add"
		if request.Action == "ceph_user.update" {
			action = "caps"
		}
		spec.args = []string{"auth", action, entity}
		raw, ok := p["caps"].(map[string]any)
		if !ok || len(raw) == 0 {
			return command{}, invalid("at least one capability is required")
		}
		for key := range raw {
			switch key {
			case "mon", "osd", "mds", "mgr":
			default:
				return command{}, invalid("unsupported capability subsystem")
			}
		}
		for _, subsystem := range []string{"mon", "osd", "mds", "mgr"} {
			value, exists := raw[subsystem]
			if !exists {
				continue
			}
			cap, ok := value.(string)
			if !ok || strings.TrimSpace(cap) == "" || strings.HasPrefix(strings.TrimSpace(cap), "-") || len(cap) > 8192 || strings.ContainsAny(cap, "\x00\r\n") {
				return command{}, invalid("invalid capability expression")
			}
			// A complete capability is one argument, including spaces and quoted paths.
			spec.args = append(spec.args, subsystem, cap)
		}
		spec.check = []string{"auth", "get", entity, "--format", "json"}
	case "ceph_user.delete":
		spec.args = []string{"auth", "rm", entity}
	default:
		return command{}, unsupported(request.Action)
	}
	return spec, nil
}

// ExportCephUsers reads only the explicitly selected entities. Keyrings are never
// added to the resource cache and the HTTP response is marked no-store.
func (s *Service) ExportCephUsers(ctx context.Context, clusterID uint64, entities []string) (string, error) {
	if clusterID == 0 || len(entities) == 0 || len(entities) > 100 {
		return "", invalid("select between 1 and 100 Ceph entities")
	}
	seen := map[string]bool{}
	for _, entity := range entities {
		if !cephEntityPattern.MatchString(entity) || seen[entity] {
			return "", invalid("invalid or duplicate Ceph entity")
		}
		seen[entity] = true
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return "", err
	}
	var output strings.Builder
	for _, entity := range entities {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "ceph_user.export", Binary: executor.BinaryCeph, Args: []string{"auth", "export", entity}, Timeout: 30 * time.Second, MaxOutput: 256 << 10})
		if err != nil {
			return "", &cephdomain.ActionError{Code: "ceph_command_failed", Message: "Ceph keyring export failed"}
		}
		if len(result.Stdout) == 0 || output.Len()+len(result.Stdout) > 256<<10 {
			return "", &cephdomain.ActionError{Code: "ceph_command_failed", Message: "Ceph keyring export is empty or exceeds 256 KiB"}
		}
		output.Write(result.Stdout)
		output.WriteByte('\n')
	}
	return output.String(), nil
}
