package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
)

var identifier = regexp.MustCompile(`^[A-Za-z0-9_.:@/+\-=]{1,512}$`)

const allowECOverwritesPoolFlag = "allow_ec_overwrites"

var rbdPoolConfigurationFields = []string{
	"rbd_qos_bps_limit",
	"rbd_qos_iops_limit",
	"rbd_qos_read_bps_limit",
	"rbd_qos_read_iops_limit",
	"rbd_qos_write_bps_limit",
	"rbd_qos_write_iops_limit",
	"rbd_qos_bps_burst",
	"rbd_qos_iops_burst",
	"rbd_qos_read_bps_burst",
	"rbd_qos_read_iops_burst",
	"rbd_qos_write_bps_burst",
	"rbd_qos_write_iops_burst",
}

type Service struct {
	clusters *clusterservice.Service
	executor executor.Executor
}

type Request struct {
	ClusterID   uint64
	Action      string
	ResourceKey string
	Parameters  map[string]any
}

func New(clusters *clusterservice.Service, runner executor.Executor) *Service {
	return &Service{clusters: clusters, executor: runner}
}

type command struct {
	binary      executor.Binary
	args, check []string
	followups   []command
	stdin       []byte
	timeout     time.Duration
	sensitive   map[int]struct{}
}

func Supports(action string) bool {
	switch action {
	case "ceph_user.create", "ceph_user.update", "ceph_user.delete", "ceph_user.import",
		"cluster.refresh", "health.mute", "health.unmute",
		"host.create", "host.update", "host.delete", "host.action", "device.identify",
		"service.create", "service.update", "service.delete", "daemon.action",
		"upgrade.check", "upgrade.action", "manager.fail", "monitor.action", "manager_module.update",
		"osd.action", "osd_flag.update", "osd.removal_check", "osd.delete",
		"osd_deployment.preview", "osd_deployment.create", "device.zap",
		"crush_rule.create", "crush_rule.update", "crush_rule.delete",
		"erasure_code_profile.create", "erasure_code_profile.delete",
		"pool.create", "pool.update", "pool.delete",
		"rbd_image.create", "rbd_image.update", "rbd_image.delete", "rbd_image.action",
		"rbd_snapshot.create", "rbd_snapshot.update", "rbd_snapshot.delete", "rbd_snapshot.action",
		"rbd_namespace.create", "rbd_namespace.delete", "rbd_trash.restore", "rbd_trash.delete",
		"rbd_trash.purge", "rbd_group.create", "rbd_group.action", "rbd_group.member", "rbd_group.snapshot", "rbd_mirroring.update", "rbd_mirroring.peer",
		"filesystem.create", "filesystem.update", "filesystem.delete",
		"subvolume_group.create", "subvolume_group.update", "subvolume_group.delete",
		"subvolume.create", "subvolume.update", "subvolume.delete",
		"cephfs_snapshot.create", "cephfs_snapshot.delete", "cephfs_snapshot.clone", "snapshot_schedule.create", "snapshot_schedule.action", "snapshot_schedule.retention",
		"cephfs_authorization.create", "cephfs_client.evict", "cephfs_entry.quota",
		"rgw_user.create", "rgw_user.update", "rgw_user.delete", "rgw_user.quota", "rgw_user.caps", "rgw_user.ratelimit", "rgw_bucket.ratelimit", "rgw_bucket.quota",
		"rgw_account.create", "rgw_account.update", "rgw_account.quota", "rgw_account.delete", "rgw_role.create", "rgw_role.update", "rgw_role.delete", "rgw_role.policy", "rgw_key.create", "rgw_key.delete",
		"rgw_realm.create", "rgw_realm.update", "rgw_zonegroup.create", "rgw_zonegroup.update", "rgw_zone.create", "rgw_zone.update", "rgw_period.commit",
		"nfs_cluster.create", "nfs_cluster.delete", "nfs_export.create", "nfs_export.update", "nfs_export.delete",
		"smb_cluster.create", "smb_cluster.update", "smb_cluster.delete",
		"smb_share.create", "smb_share.update", "smb_share.delete",
		"config_value.set", "config_value.delete":
		return true
	default:
		return false
	}
}

func (s *Service) Execute(ctx context.Context, request Request) (cephdomain.ActionResult, error) {
	if request.ClusterID == 0 {
		return cephdomain.ActionResult{}, unsupported(request.Action)
	}
	access, err := s.clusters.Access(ctx, request.ClusterID)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	defer func() { access.ClientKey = "" }()
	if request.Parameters == nil {
		request.Parameters = map[string]any{}
	}
	spec, err := build(request, request.Parameters)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action, Binary: spec.binary, Args: spec.args, Stdin: spec.stdin, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: request.Action != "osd_deployment.preview", SensitiveArgs: spec.sensitive})
	if err != nil {
		if request.Action == "config_value.set" && len(spec.sensitive) > 0 {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "Ceph rejected the sensitive configuration update"}
		}
		if request.Action == "ceph_user.import" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "Ceph keyring import failed"}
		}
		return cephdomain.ActionResult{}, normalize(err)
	}
	checkSpec := spec
	for index, followup := range spec.followups {
		stepID := fmt.Sprintf("%s.step%d", request.Action, index+2)
		result, err = s.executor.Run(ctx, access, executor.CommandSpec{ID: stepID, Binary: followup.binary, Args: followup.args, Stdin: followup.stdin, Timeout: followup.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: true, SensitiveArgs: followup.sensitive})
		if err != nil {
			return cephdomain.ActionResult{}, normalize(err)
		}
		if len(followup.check) > 0 {
			checkSpec = followup
		}
	}
	if len(checkSpec.check) > 0 {
		checked, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + ".post_check", Binary: checkSpec.binary, Args: checkSpec.check, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		if err != nil || (request.Action == "rgw_bucket.quota" && !bucketQuotaMatches(request.Parameters, checked.Stdout)) {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "command was accepted but the expected state could not be verified", Retryable: true}
		}
	}
	if request.Action == "osd_deployment.preview" {
		return cephdomain.ActionResult{Details: map[string]any{"preview": security.Redact(string(result.Stdout))}}, nil
	}
	return cephdomain.ActionResult{Details: map[string]any{"exit_code": result.ExitCode, "duration_ms": result.Duration.Milliseconds()}}, nil
}

func build(request Request, p map[string]any) (command, error) {
	action := request.Action
	tail := resourceTail(request.ResourceKey)
	ceph := func(args, check []string) command {
		return command{binary: executor.BinaryCeph, args: args, check: check, timeout: 2 * time.Minute}
	}
	rbd := func(args, check []string) command {
		if len(check) > 0 {
			check = append(check, "--format", "json")
		}
		return command{binary: executor.BinaryRBD, args: args, check: check, timeout: 5 * time.Minute}
	}
	rgw := func(args, check []string) command {
		if strings.HasPrefix(action, "rgw_role.") {
			if account := optional(p, "account_id"); account != "" {
				args = append(args, "--account-id", account)
				if len(check) > 0 {
					check = append(check, "--account-id", account)
				}
			}
		}
		if len(check) > 0 {
			check = append(check, "--format", "json")
		}
		return command{binary: executor.BinaryRGWAdmin, args: append(args, "--format", "json"), check: check, timeout: 2 * time.Minute}
	}
	cephfsShell := func(args []string) command {
		return command{binary: executor.BinaryCephFSShell, args: args, timeout: 2 * time.Minute}
	}
	switch action {
	case "ceph_user.create", "ceph_user.update", "ceph_user.delete", "ceph_user.import":
		return cephUserCommand(request, p)
	case "cluster.refresh":
		return ceph([]string{"status", "--format", "json"}, []string{"status", "--format", "json"}), nil
	case "health.mute":
		code := last(tail)
		if !regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`).MatchString(code) {
			return command{}, invalid("health check code is invalid")
		}
		args := []string{"health", "mute", code}
		if value, exists := p["ttl"]; exists {
			ttl, ok := value.(string)
			if !ok || !regexp.MustCompile(`^[1-9][0-9]{0,8}[smhdw]?$`).MatchString(ttl) {
				return command{}, invalid("ttl must be a positive duration in seconds, minutes, hours, days or weeks")
			}
			args = append(args, ttl)
		}
		if value, exists := p["sticky"]; exists {
			sticky, ok := value.(bool)
			if !ok {
				return command{}, invalid("sticky must be a boolean")
			}
			if sticky {
				args = append(args, "--sticky")
			}
		}
		return ceph(args, []string{"health", "detail", "--format", "json"}), nil
	case "health.unmute":
		return ceph([]string{"health", "unmute", last(tail)}, []string{"health", "detail", "--format", "json"}), nil
	case "host.create":
		name, err := required(p, "hostname")
		if err != nil {
			return command{}, err
		}
		args := []string{"orch", "host", "add", name}
		if addr := optional(p, "address"); addr != "" {
			args = append(args, "--addr", addr)
		}
		if rawLabels, exists := p["labels"]; exists {
			labels, ok := stringSlice(rawLabels)
			if !ok {
				return command{}, invalid("labels are invalid")
			}
			if len(labels) > 0 {
				args = append(args, "--labels", strings.Join(labels, ","))
			}
		}
		if boolParameter(p, "maintenance") {
			args = append(args, "--maintenance")
		}
		return ceph(args, []string{"orch", "host", "ls", "--detail", "--format", "json"}), nil
	case "host.delete":
		return ceph([]string{"orch", "host", "rm", last(tail)}, []string{"orch", "host", "ls", "--detail", "--format", "json"}), nil
	case "host.update":
		return hostUpdate(p, last(tail), ceph)
	case "host.action":
		return hostAction(p, pathValue(tail, "host"), ceph)
	case "device.identify":
		host := pathValue(tail, "host")
		device, err := required(p, "device")
		if err != nil {
			return command{}, err
		}
		state, err := enum(p, "state", "on", "off")
		if err != nil {
			return command{}, err
		}
		light := optional(p, "light")
		if light == "" {
			light = "ident"
		}
		if light != "ident" && light != "fault" {
			return command{}, invalid("light is not supported")
		}
		return ceph([]string{"device", "light", state, device, light, "--force"}, []string{"orch", "device", "ls", "--host", host, "--format", "json"}), nil
	case "service.create", "service.update":
		serviceType, err := enum(p, "service_type", "mon", "mgr", "mds", "rgw", "nfs", "smb", "prometheus", "alertmanager", "grafana", "node-exporter", "crash")
		if err != nil {
			return command{}, err
		}
		serviceID := optional(p, "service_id")
		if action == "service.update" && serviceID == "" {
			serviceID = last(tail)
		}
		spec := map[string]any{"service_type": serviceType}
		if serviceID != "" {
			spec["service_id"] = serviceID
		}
		if placement, ok := p["placement"].(map[string]any); ok {
			allowed := map[string]any{}
			for _, key := range []string{"count", "host_pattern", "hosts", "label"} {
				if value, exists := placement[key]; exists {
					allowed[key] = value
				}
			}
			spec["placement"] = allowed
		}
		stdin, err := json.Marshal(spec)
		if err != nil {
			return command{}, invalid("service spec is invalid")
		}
		result := ceph([]string{"orch", "apply", "-i", "-"}, []string{"orch", "ls", "--export", "--format", "json"})
		result.stdin = stdin
		return result, nil
	case "service.delete":
		name := last(tail)
		return ceph([]string{"orch", "rm", name}, []string{"orch", "ls", "--export", "--format", "json"}), nil
	case "daemon.action":
		name := pathValue(tail, "daemon")
		verb, err := enum(p, "action", "start", "stop", "restart", "reconfig", "redeploy", "rotate-key")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"orch", "daemon", verb, name}, []string{"orch", "ps", "--daemon_name", name, "--refresh", "--format", "json"}), nil
	case "upgrade.check":
		version, err := required(p, "version")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"orch", "upgrade", "check", version, "--format", "json"}, nil), nil
	case "upgrade.action":
		verb, err := enum(p, "action", "start", "pause", "resume", "stop")
		if err != nil {
			return command{}, err
		}
		args := []string{"orch", "upgrade", verb}
		if verb == "start" {
			version, err := required(p, "version")
			if err != nil {
				return command{}, err
			}
			args = append(args, "--ceph-version", version)
		}
		return ceph(args, []string{"orch", "upgrade", "status", "--format", "json"}), nil
	case "manager.fail":
		return ceph([]string{"mgr", "fail", last(tail)}, []string{"mgr", "dump", "--format", "json"}), nil
	case "monitor.action":
		verb, err := enum(p, "action", "scrub", "ok-to-stop")
		if err != nil {
			return command{}, err
		}
		args := []string{"mon", verb}
		if verb == "ok-to-stop" {
			names, ok := stringSlice(p["names"])
			if !ok || len(names) == 0 {
				return command{}, invalid("names must be a non-empty string array")
			}
			args = append(args, names...)
		}
		return ceph(args, []string{"quorum_status", "--format", "json"}), nil
	case "manager_module.update":
		name := last(tail)
		enabled, ok := p["enabled"].(bool)
		if !ok {
			return command{}, invalid("enabled must be boolean")
		}
		verb := "disable"
		if enabled {
			verb = "enable"
		}
		return ceph([]string{"mgr", "module", verb, name}, []string{"mgr", "module", "ls", "--format", "json"}), nil
	case "osd.action":
		id := pathValue(tail, "osd")
		verb, err := enum(p, "action", "in", "out", "down", "reweight", "scrub", "deep-scrub")
		if err != nil {
			return command{}, err
		}
		args := []string{"osd", verb, id}
		if verb == "reweight" {
			weight, err := required(p, "weight")
			if err != nil {
				return command{}, err
			}
			args = append(args, weight)
		}
		return ceph(args, []string{"osd", "dump", "--format", "json"}), nil
	case "osd_flag.update":
		verb, err := enum(p, "action", "set", "unset")
		if err != nil {
			return command{}, err
		}
		flag, err := required(p, "flag")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"osd", verb, flag}, []string{"osd", "dump", "--format", "json"}), nil
	case "osd.removal_check":
		ids, ok := stringSlice(p["osd_ids"])
		if !ok || len(ids) == 0 {
			return command{}, invalid("osd_ids must be a non-empty array")
		}
		return ceph(append([]string{"osd", "safe-to-destroy"}, ids...), nil), nil
	case "osd.delete":
		id := pathValue(tail, "osd")
		args := []string{"orch", "osd", "rm", id}
		if zap, _ := p["zap"].(bool); zap {
			args = append(args, "--zap")
		}
		return ceph(args, []string{"orch", "osd", "rm", "status", "--format", "json"}), nil
	case "osd_deployment.preview", "osd_deployment.create":
		spec, err := osdSpec(p)
		if err != nil {
			return command{}, err
		}
		stdin, _ := json.Marshal(spec)
		args := []string{"orch", "apply", "osd", "-i", "-"}
		if action == "osd_deployment.preview" {
			args = append(args, "--dry-run")
		}
		result := ceph(args, []string{"orch", "ps", "--daemon-type", "osd", "--format", "json"})
		result.stdin = stdin
		if action == "osd_deployment.preview" {
			result.check = nil
		}
		return result, nil
	case "device.zap":
		host, device, err := decodePair(pathValue(tail, "device"))
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"orch", "device", "zap", host, device, "--force"}, []string{"orch", "device", "ls", "--host", host, "--refresh", "--format", "json"}), nil
	case "crush_rule.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		root, err := required(p, "root")
		if err != nil {
			return command{}, err
		}
		failure := optional(p, "failure_domain")
		if failure == "" {
			failure = "host"
		}
		class := optional(p, "device_class")
		args := []string{"osd", "crush", "rule", "create-replicated", name, root, failure}
		if class != "" {
			args = append(args, class)
		}
		return ceph(args, []string{"osd", "crush", "rule", "dump", name, "--format", "json"}), nil
	case "crush_rule.update":
		old := last(tail)
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"osd", "crush", "rule", "rename", old, name}, []string{"osd", "crush", "rule", "dump", name, "--format", "json"}), nil
	case "crush_rule.delete":
		return ceph([]string{"osd", "crush", "rule", "rm", last(tail)}, []string{"osd", "crush", "rule", "ls", "--format", "json"}), nil
	case "erasure_code_profile.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		args := []string{"osd", "erasure-code-profile", "set", name}
		for _, key := range []string{
			"plugin", "k", "m", "technique", "packetsize", "l", "crush-locality", "c", "d", "scalar_mds",
			"crush-failure-domain", "crush-num-failure-domains", "crush-osds-per-failure-domain",
			"crush-root", "crush-device-class", "directory",
		} {
			if value := optional(p, key); value != "" {
				args = append(args, key+"="+value)
			}
		}
		return ceph(args, []string{"osd", "erasure-code-profile", "get", name, "--format", "json"}), nil
	case "erasure_code_profile.delete":
		return ceph([]string{"osd", "erasure-code-profile", "rm", last(tail)}, []string{"osd", "erasure-code-profile", "ls", "--format", "json"}), nil
	case "pool.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		pg := optional(p, "pg_num")
		if pg == "" {
			pg = "32"
		}
		args := []string{"osd", "pool", "create", name, pg}
		poolType := optional(p, "pool_type")
		if poolType == "erasure" {
			profile := optional(p, "erasure_code_profile")
			if profile == "" {
				profile = "default"
			}
			args = append(args, pg, "erasure", profile)
		} else if poolType == "replicated" {
			args = append(args, pg, "replicated")
		}
		result := ceph(args, []string{"osd", "pool", "ls", "detail", "--format", "json"})
		result.followups, err = poolCreateFollowups(p, name, poolType, ceph, rbd)
		if err != nil {
			return command{}, err
		}
		return result, nil
	case "pool.update":
		name := last(tail)
		operation := optional(p, "operation")
		if operation == "quota" {
			field, err := enum(p, "field", "max_bytes", "max_objects")
			if err != nil {
				return command{}, err
			}
			value, err := required(p, "value")
			if err != nil {
				return command{}, err
			}
			return ceph([]string{"osd", "pool", "set-quota", name, field, value}, []string{"osd", "pool", "get-quota", name, "--format", "json"}), nil
		}
		if operation == "application" {
			verb, err := enum(p, "action", "enable", "disable")
			if err != nil {
				return command{}, err
			}
			application, err := required(p, "application")
			if err != nil {
				return command{}, err
			}
			args := []string{"osd", "pool", "application", verb, name, application}
			if verb == "disable" {
				args = append(args, "--yes-i-really-mean-it")
			}
			return ceph(args, []string{"osd", "pool", "application", "get", name, "--format", "json"}), nil
		}
		if operation == "rbd_configuration" {
			field, err := enum(p, "field", rbdPoolConfigurationFields...)
			if err != nil {
				return command{}, err
			}
			value, err := requiredNonNegativeInteger(p, "value")
			if err != nil {
				return command{}, err
			}
			return rbd(
				[]string{"config", "pool", "set", name, field, value},
				[]string{"config", "pool", "list", name},
			), nil
		}
		if operation == "rbd_mirroring" {
			mode, err := enum(p, "rbd_mirroring", "disabled", "pool")
			if err != nil {
				return command{}, err
			}
			return rbdMirrorPoolModeCommand(name, mode, rbd), nil
		}
		if operation == "rename" {
			newName, err := required(p, "name")
			if err != nil {
				return command{}, err
			}
			return ceph([]string{"osd", "pool", "rename", name, newName}, []string{"osd", "pool", "ls", "detail", "--format", "json"}), nil
		}
		field, err := enum(p, "field", poolSetFields()...)
		if err != nil {
			return command{}, err
		}
		value, err := required(p, "value")
		if err != nil {
			return command{}, err
		}
		if err := validatePoolSetValue(field, value); err != nil {
			return command{}, err
		}
		result := ceph([]string{"osd", "pool", "set", name, field, value}, []string{"osd", "pool", "ls", "detail", "--format", "json"})
		if field == "pg_num" {
			result.followups = []command{ceph([]string{"osd", "pool", "set", name, "pgp_num", value}, nil)}
		}
		return result, nil
	case "pool.delete":
		name := last(tail)
		return ceph([]string{"osd", "pool", "rm", name, name, "--yes-i-really-really-mean-it"}, []string{"osd", "pool", "ls", "detail", "--format", "json"}), nil
	case "rbd_image.create":
		spec, err := required(p, "image_spec")
		if err != nil {
			return command{}, err
		}
		if err := validateRBDImagePath(spec); err != nil {
			return command{}, err
		}
		size, err := required(p, "size")
		if err != nil {
			return command{}, err
		}
		args := []string{"create", spec, "--size", size + "B"}
		if pool := optional(p, "data_pool"); pool != "" {
			if !identifier.MatchString(pool) || strings.Contains(pool, "/") {
				return command{}, invalid("data_pool is invalid")
			}
			args = append(args, "--data-pool", pool)
		}
		for _, field := range []string{"object_size", "stripe_unit", "stripe_count"} {
			if value := optional(p, field); value != "" {
				n, err := strconv.ParseUint(value, 10, 64)
				if err != nil || n == 0 {
					return command{}, invalid(field + " must be a positive integer")
				}
				if field == "object_size" && (n < 4096 || n > 33554432 || n&(n-1) != 0) {
					return command{}, invalid("object_size must be a power of two from 4096 to 33554432 bytes")
				}
				if field == "stripe_unit" || field == "object_size" {
					value += "B"
				}
				args = append(args, "--"+strings.ReplaceAll(field, "_", "-"), value)
			}
		}
		return rbd(args, []string{"info", spec}), nil
	case "rbd_image.update":
		spec, err := decodeImageSpec(last(tail))
		if err != nil {
			return command{}, err
		}
		if size := optional(p, "size"); size != "" {
			args := []string{"resize", spec, "--size", size + "B"}
			if allow, _ := p["allow_shrink"].(bool); allow {
				args = append(args, "--allow-shrink")
			}
			return rbd(args, []string{"info", spec}), nil
		}
		if name := optional(p, "name"); name != "" {
			if !identifier.MatchString(name) || strings.ContainsAny(name, "/@") {
				return command{}, invalid("name must be an image name")
			}
			destination := spec[:strings.LastIndexByte(spec, '/')+1] + name
			return rbd([]string{"rename", spec, destination}, []string{"info", destination}), nil
		}
		featureAction := optional(p, "feature_action")
		if featureAction == "enable" || featureAction == "disable" {
			features, err := required(p, "features")
			if err != nil {
				return command{}, err
			}
			return rbd([]string{"feature", featureAction, spec, features}, []string{"info", spec}), nil
		}
		return command{}, invalid("size, name, or feature_action is required")
	case "rbd_image.delete":
		spec, err := decodeImageSpec(last(tail))
		if err != nil {
			return command{}, err
		}
		parts := strings.Split(spec, "/")
		check := []string{"ls", "--pool", parts[0]}
		if len(parts) == 3 {
			check = append(check, "--namespace", parts[1])
		}
		return rbd([]string{"rm", spec}, check), nil
	case "rbd_snapshot.create":
		spec, err := decodeImageSpec(pathValue(tail, "image"))
		if err != nil {
			return command{}, err
		}
		snap, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"snap", "create", spec + "@" + snap}, []string{"snap", "ls", spec}), nil
	case "rbd_snapshot.delete":
		spec, err := decodeImageSpec(pathValue(tail, "image"))
		if err != nil {
			return command{}, err
		}
		snap := last(tail)
		return rbd([]string{"snap", "rm", spec + "@" + snap}, []string{"snap", "ls", spec}), nil
	case "rbd_namespace.create":
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"namespace", "create", pool + "/" + name}, []string{"namespace", "list", pool}), nil
	case "rbd_namespace.delete":
		pool := pathValue(tail, "namespace")
		namespace := last(tail)
		return rbd([]string{"namespace", "remove", pool + "/" + namespace}, []string{"namespace", "list", pool}), nil
	case "rbd_image.action":
		spec, err := decodeImageSpec(pathValue(tail, "image"))
		if err != nil {
			return command{}, err
		}
		verb, err := enum(p, "action", "config-set", "config-remove", "snapshot-purge", "feature-enable", "feature-disable", "flatten", "sparsify", "copy", "deep-copy", "rename", "move-to-trash", "mirror-enable-journal", "mirror-enable-snapshot", "mirror-disable", "mirror-promote", "mirror-demote", "mirror-resync", "mirror-snapshot")
		if err != nil {
			return command{}, err
		}
		if verb == "config-set" || verb == "config-remove" {
			name, err := required(p, "config_name")
			if err != nil {
				return command{}, err
			}
			if !strings.HasPrefix(name, "rbd_") || strings.ContainsAny(name, "/@") || security.IsSensitiveName(name) {
				return command{}, invalid("config_name must be a non-secret RBD option")
			}
			args := []string{"config", "image", strings.TrimPrefix(verb, "config-"), spec, name}
			if verb == "config-set" {
				value, ok := p["config_value"].(string)
				if !ok || value == "" || strings.ContainsAny(value, "\x00\r\n") || strings.HasPrefix(value, "-") {
					return command{}, invalid("config_value must be a nonempty single-line value")
				}
				args = append(args, value)
			}
			return rbd(args, []string{"config", "image", "list", spec}), nil
		}
		if verb == "snapshot-purge" {
			return rbd([]string{"snap", "purge", spec}, []string{"snap", "ls", spec}), nil
		}
		if verb == "feature-enable" || verb == "feature-disable" {
			feature, err := enum(p, "feature", "exclusive-lock", "object-map", "fast-diff", "journaling", "deep-flatten")
			if err != nil {
				return command{}, err
			}
			if verb == "feature-enable" && feature == "deep-flatten" {
				return command{}, invalid("deep-flatten cannot be enabled on an existing image")
			}
			return rbd([]string{"feature", strings.TrimPrefix(verb, "feature-"), spec, feature}, []string{"info", spec}), nil
		}
		if verb == "rename" {
			name, err := required(p, "destination")
			if err != nil {
				return command{}, err
			}
			if strings.ContainsAny(name, "/@") {
				return command{}, invalid("destination must be a new image name within the same pool and namespace")
			}
			parts := strings.Split(spec, "/")
			destination := strings.Join(parts[:len(parts)-1], "/") + "/" + name
			return rbd([]string{"rename", spec, destination}, []string{"info", destination}), nil
		}
		if strings.HasPrefix(verb, "mirror-") {
			mirrorVerb := strings.TrimPrefix(verb, "mirror-")
			args := []string{"mirror", "image", mirrorVerb, spec}
			if strings.HasPrefix(mirrorVerb, "enable-") {
				args = []string{"mirror", "image", "enable", spec, strings.TrimPrefix(mirrorVerb, "enable-")}
			}
			check := []string{"mirror", "image", "status", spec}
			if mirrorVerb == "disable" {
				check = []string{"info", spec}
			}
			return rbd(args, check), nil
		}
		if verb == "move-to-trash" {
			scope := strings.Split(spec, "/")
			check := []string{"trash", "ls", "--pool", scope[0]}
			if len(scope) == 3 {
				check = append(check, "--namespace", scope[1])
			}
			args := []string{"trash", "mv", spec}
			if expiresAt := optional(p, "expires_at"); expiresAt != "" {
				expires, err := time.Parse(time.RFC3339, expiresAt)
				if err != nil {
					return command{}, invalid("expires_at must be an RFC3339 timestamp with timezone")
				}
				args = append(args, "--expires-at="+expires.UTC().Format(time.RFC3339))
			}
			return rbd(args, check), nil
		}
		commandVerb := verb
		if verb == "copy" {
			commandVerb = "cp"
		}
		args := []string{commandVerb, spec}
		if verb == "copy" || verb == "deep-copy" {
			destination, err := required(p, "destination")
			if err != nil {
				return command{}, err
			}
			if err := validateRBDImagePath(destination); err != nil {
				return command{}, err
			}
			args = append(args, destination)
			return rbd(args, []string{"info", destination}), nil
		}
		return rbd(args, []string{"info", spec}), nil
	case "rbd_snapshot.update":
		spec, err := decodeImageSpec(pathValue(tail, "image"))
		if err != nil {
			return command{}, err
		}
		old := last(tail)
		if action := optional(p, "action"); action == "protect" || action == "unprotect" {
			return rbd([]string{"snap", action, spec + "@" + old}, []string{"snap", "ls", spec}), nil
		}
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"snap", "rename", spec + "@" + old, spec + "@" + name}, []string{"snap", "ls", spec}), nil
	case "rbd_snapshot.action":
		spec, err := decodeImageSpec(pathValue(tail, "image"))
		if err != nil {
			return command{}, err
		}
		snap := last(strings.TrimSuffix(tail, "/action"))
		verb, err := enum(p, "action", "clone", "protect", "unprotect", "rollback")
		if err != nil {
			return command{}, err
		}
		if verb == "clone" {
			destination, err := required(p, "destination")
			if err != nil {
				return command{}, err
			}
			if err := validateRBDImagePath(destination); err != nil {
				return command{}, err
			}
			return rbd([]string{"clone", spec + "@" + snap, destination}, []string{"info", destination}), nil
		}
		return rbd([]string{"snap", verb, spec + "@" + snap}, []string{"snap", "ls", spec}), nil
	case "rbd_trash.restore":
		imageID := pathValue(tail, "trash")
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"trash", "restore", pool + "/" + imageID, "--image", name}, []string{"info", pool + "/" + name}), nil
	case "rbd_trash.delete":
		pool, imageID, err := decodePair(last(tail))
		if err != nil {
			return command{}, err
		}
		scope := strings.Split(pool, "/")
		check := []string{"trash", "ls", "--pool", scope[0]}
		if len(scope) == 2 {
			check = append(check, "--namespace", scope[1])
		}
		return rbd([]string{"trash", "remove", pool + "/" + imageID, "--force"}, check), nil
	case "rbd_trash.purge":
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		scope := strings.Split(pool, "/")
		if len(scope) > 2 || scope[0] == "" || (len(scope) == 2 && scope[1] == "") {
			return command{}, invalid("pool must be pool or pool/namespace")
		}
		flags := []string{"--pool", scope[0]}
		if len(scope) == 2 {
			flags = append(flags, "--namespace", scope[1])
		}
		args := append([]string{"trash", "purge"}, flags...)
		if cutoff := optional(p, "expired_before"); cutoff != "" {
			expires, err := time.Parse(time.RFC3339, cutoff)
			if err != nil {
				return command{}, invalid("expired_before must be an RFC3339 timestamp with timezone")
			}
			args = append(args, "--expired-before="+expires.UTC().Format(time.RFC3339))
		}
		return rbd(args, append([]string{"trash", "ls"}, flags...)), nil
	case "rbd_group.snapshot":
		group, err := required(p, "group_spec")
		if err != nil {
			return command{}, err
		}
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		verb, err := enum(p, "action", "create", "remove", "rollback", "rename")
		if err != nil {
			return command{}, err
		}
		args := []string{"group", "snap", verb, group + "@" + name}
		if verb == "rename" {
			destination, err := required(p, "new_name")
			if err != nil {
				return command{}, err
			}
			if strings.ContainsAny(destination, "/@") {
				return command{}, invalid("new_name must be a snapshot name")
			}
			args = append(args, destination)
		}
		return rbd(args, []string{"group", "snap", "list", group}), nil
	case "rbd_group.member":
		group, err := required(p, "group_spec")
		if err != nil {
			return command{}, err
		}
		image, err := required(p, "image")
		if err != nil {
			return command{}, err
		}
		verb, err := enum(p, "action", "add", "remove")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"group", "image", verb, group, image}, []string{"group", "image", "list", group}), nil
	case "rbd_group.action":
		group, err := required(p, "group_spec")
		if err != nil {
			return command{}, err
		}
		parts := strings.Split(group, "/")
		if len(parts) < 2 || len(parts) > 3 {
			return command{}, invalid("group_spec must include pool and group")
		}
		for _, part := range parts {
			if part == "" {
				return command{}, invalid("group_spec contains an empty component")
			}
		}
		verb, err := enum(p, "action", "rename", "remove")
		if err != nil {
			return command{}, err
		}
		args := []string{"group", verb, group}
		if verb == "rename" {
			name, err := required(p, "name")
			if err != nil {
				return command{}, err
			}
			if strings.Contains(name, "/") {
				return command{}, invalid("name must be a group name within the same pool and namespace")
			}
			args = append(args, strings.Join(parts[:len(parts)-1], "/")+"/"+name)
		}
		check := []string{"group", "list", "--pool", parts[0]}
		if len(parts) == 3 {
			check = append(check, "--namespace", parts[1])
		}
		return rbd(args, check), nil
	case "rbd_group.create":
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"group", "create", pool + "/" + name}, []string{"group", "list", pool}), nil
	case "rbd_mirroring.peer":
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		verb, err := enum(p, "action", "add", "remove", "set")
		if err != nil {
			return command{}, err
		}
		args := []string{"mirror", "pool", "peer", verb, pool}
		if verb == "remove" || verb == "set" {
			id, err := required(p, "uuid")
			if err != nil {
				return command{}, err
			}
			args = append(args, id)
			if verb == "set" {
				field, err := enum(p, "field", "site-name", "client", "mon-host", "direction")
				if err != nil {
					return command{}, err
				}
				value := optional(p, "value")
				if value == "" || strings.ContainsAny(value, "\x00\r\n") || strings.HasPrefix(value, "-") {
					return command{}, invalid("value is required or invalid")
				}
				if field == "direction" {
					value, err = enum(p, "value", "rx-only", "tx-only", "rx-tx")
					if err != nil {
						return command{}, err
					}
				}
				args = append(args, field, value)
			}
		} else {
			site, err := required(p, "remote_cluster")
			if err != nil {
				return command{}, err
			}
			client, err := required(p, "remote_client")
			if err != nil {
				return command{}, err
			}
			direction, err := enum(p, "direction", "rx-only", "rx-tx")
			if err != nil {
				return command{}, err
			}
			args = append(args, "--remote-cluster="+site, "--remote-client-name="+client, "--direction="+direction)
		}
		return rbd(args, []string{"mirror", "pool", "info", pool}), nil
	case "rbd_mirroring.update":
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		mode, err := enum(p, "mode", "disabled", "image", "pool")
		if err != nil {
			return command{}, err
		}
		return rbdMirrorPoolModeCommand(pool, mode, rbd), nil
	case "filesystem.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		placement := rawText(p, "placement")
		metadataPool := optional(p, "metadata_pool")
		dataPool := optional(p, "data_pool")
		if (metadataPool == "") != (dataPool == "") {
			return command{}, invalid("metadata_pool and data_pool must be provided together")
		}
		if metadataPool != "" {
			if !identifier.MatchString(metadataPool) || !identifier.MatchString(dataPool) {
				return command{}, invalid("metadata_pool or data_pool is invalid")
			}
			if placement == "" {
				placement = "1"
			}
		}
		if len(placement) > 1024 || strings.ContainsRune(placement, 0) {
			return command{}, invalid("placement is invalid")
		}
		args := []string{"fs", "volume", "create", name}
		if placement != "" {
			args = append(args, placement)
		}
		if metadataPool != "" {
			args = append(args, metadataPool, dataPool)
		}
		return ceph(args, []string{"fs", "volume", "ls", "--format", "json"}), nil
	case "filesystem.delete":
		name := last(tail)
		return ceph([]string{"fs", "volume", "rm", name, "--yes-i-really-mean-it"}, []string{"fs", "volume", "ls", "--format", "json"}), nil
	case "filesystem.update":
		fs := last(tail)
		maxMDS := optional(p, "max_mds")
		if maxMDS == "" {
			return command{}, invalid("max_mds is required")
		}
		return ceph([]string{"fs", "set", fs, "max_mds", maxMDS}, []string{"fs", "get", fs, "--format", "json"}), nil
	case "subvolume_group.create":
		fs := pathValue(tail, "filesystem")
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		args, err := subvolumeGroupCreateArgs(fs, name, p)
		if err != nil {
			return command{}, err
		}
		return ceph(args, []string{"fs", "subvolumegroup", "info", fs, name, "--format", "json"}), nil
	case "subvolume_group.update":
		fs := pathValue(tail, "filesystem")
		name := last(tail)
		size := optional(p, "size")
		if size == "" {
			return command{}, invalid("size is required")
		}
		return ceph([]string{"fs", "subvolumegroup", "resize", fs, name, size}, []string{"fs", "subvolumegroup", "info", fs, name, "--format", "json"}), nil
	case "subvolume_group.delete":
		fs := pathValue(tail, "filesystem")
		name := last(tail)
		return ceph([]string{"fs", "subvolumegroup", "rm", fs, name}, []string{"fs", "subvolumegroup", "ls", fs, "--format", "json"}), nil
	case "subvolume.create":
		fs := pathValue(tail, "filesystem")
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		args, err := subvolumeCreateArgs(fs, name, p)
		if err != nil {
			return command{}, err
		}
		return ceph(args, []string{"fs", "subvolume", "info", fs, name, optional(p, "group"), "--format", "json"}), nil
	case "subvolume.delete":
		fs := pathValue(tail, "filesystem")
		name := last(tail)
		return ceph([]string{"fs", "subvolume", "rm", fs, name}, []string{"fs", "subvolume", "ls", fs, "--format", "json"}), nil
	case "subvolume.update":
		fs := pathValue(tail, "filesystem")
		name := last(tail)
		size := optional(p, "size")
		if size == "" {
			return command{}, invalid("size is required")
		}
		return ceph([]string{"fs", "subvolume", "resize", fs, name, size}, []string{"fs", "subvolume", "info", fs, name, "--format", "json"}), nil
	case "cephfs_snapshot.create":
		fs := pathValue(tail, "filesystem")
		subvolume := pathValue(tail, "subvolume")
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"fs", "subvolume", "snapshot", "create", fs, subvolume, name}, []string{"fs", "subvolume", "snapshot", "ls", fs, subvolume, "--format", "json"}), nil
	case "cephfs_snapshot.delete":
		fs := pathValue(tail, "filesystem")
		subvolume := pathValue(tail, "subvolume")
		snapshot := last(tail)
		args := []string{"fs", "subvolume", "snapshot", "rm", fs, subvolume, snapshot}
		if group := optional(p, "group"); group != "" {
			args = append(args, group)
		}
		return ceph(args, []string{"fs", "subvolume", "snapshot", "ls", fs, subvolume, "--format", "json"}), nil
	case "cephfs_snapshot.clone":
		fs := pathValue(tail, "filesystem")
		subvolume := pathValue(tail, "subvolume")
		snapshot := pathValue(tail, "snapshot")
		target, err := required(p, "target")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"fs", "subvolume", "snapshot", "clone", fs, subvolume, snapshot, target}, []string{"fs", "clone", "status", fs, target, "--format", "json"}), nil
	case "snapshot_schedule.retention":
		return snapshotRetention(request, p)
	case "snapshot_schedule.create", "snapshot_schedule.action":
		fs := pathValue(tail, "filesystem")
		path, err := required(p, "path")
		if err != nil {
			return command{}, err
		}
		schedule, err := required(p, "schedule")
		if err != nil {
			return command{}, err
		}
		if fs == "" {
			return command{}, invalid("filesystem is required")
		}
		args := []string{"fs", "snap-schedule", "add", path, schedule, "--fs", fs}
		if action == "snapshot_schedule.action" {
			if _, err := required(p, "start"); err != nil {
				return command{}, err
			}
			verb, err := enum(p, "action", "activate", "deactivate", "remove")
			if err != nil {
				return command{}, err
			}
			args = []string{"fs", "snap-schedule", verb, path, "--repeat=" + schedule, "--fs=" + fs}
		}
		for _, option := range []string{"start", "subvol", "group"} {
			if value, ok := p[option].(string); ok && value != "" {
				if strings.ContainsAny(value, "\x00\r\n") {
					return command{}, invalid("invalid snapshot schedule option")
				}
				args = append(args, "--"+option+"="+value)
			}
		}

		check := []string{"fs", "snap-schedule", "status", path, "--fs", fs, "--format", "json"}
		for _, option := range []string{"subvol", "group"} {
			if value, ok := p[option].(string); ok && value != "" {
				check = append(check, "--"+option+"="+value)
			}
		}
		return ceph(args, check), nil
	case "cephfs_authorization.create":
		fs := pathValue(tail, "filesystem")
		client, err := required(p, "client")
		if err != nil {
			return command{}, err
		}
		path := optional(p, "path")
		if path == "" {
			path = "/"
		}
		access, err := enum(p, "access", "r", "rw")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"fs", "authorize", fs, client, path, access}, []string{"auth", "get", client, "--format", "json"}), nil
	case "cephfs_client.evict":
		return ceph([]string{"tell", "mds.*", "client", "evict", "id=" + last(tail)}, []string{"fs", "status", pathValue(tail, "filesystem"), "--format", "json"}), nil
	case "cephfs_entry.quota":
		path, err := required(p, "path")
		if err != nil {
			return command{}, err
		}
		bytes := optional(p, "max_bytes")
		if bytes == "" {
			return command{}, invalid("max_bytes is required")
		}
		return cephfsShell([]string{"setxattr", path, "ceph.quota.max_bytes", bytes}), nil
	case "rgw_user.create":
		uid, err := required(p, "uid")
		if err != nil {
			return command{}, err
		}
		if strings.TrimSpace(rawText(p, "display_name")) == "" {
			return command{}, invalid("display_name is required")
		}
		args := []string{"user", "create", "--uid", uid}
		if maximum := optional(p, "max_buckets"); maximum != "" {
			limit, err := strconv.ParseInt(maximum, 10, 32)
			if err != nil || limit < -1 {
				return command{}, invalid("max_buckets must be -1 or a nonnegative integer")
			}
			args = append(args, "--max-buckets", maximum)
		}
		for field, flag := range map[string]string{"display_name": "--display-name", "email": "--email"} {
			if value := rawText(p, field); value != "" {
				args = append(args, flag, value)
			}
		}
		return rgw(args, []string{"user", "info", "--uid", uid}), nil
	case "rgw_user.update":
		uid := last(tail)
		args := []string{"user", "modify", "--uid", uid}
		for field, flag := range map[string]string{"display_name": "--display-name", "max_buckets": "--max-buckets"} {
			if value := rawText(p, field); value != "" {
				args = append(args, flag, value)
			}
		}
		if email, ok := p["email"].(string); ok {
			if strings.ContainsAny(email, "\x00\r\n") {
				return command{}, invalid("email must be a single-line string")
			}
			args = append(args, "--email="+email)
		}
		for field, flag := range map[string]string{"system": "--system"} {
			if value, ok := p[field].(bool); ok {
				args = append(args, flag, strconv.FormatBool(value))
			}
		}
		if suspended, ok := p["suspended"].(bool); ok {
			verb := "enable"
			if suspended {
				verb = "suspend"
			}
			stateCommand := rgw([]string{"user", verb, "--uid", uid}, []string{"user", "info", "--uid", uid})
			if len(args) == 4 {
				return stateCommand, nil
			}
			result := rgw(args, nil)
			result.followups = []command{stateCommand}
			return result, nil
		}
		if len(args) == 4 {
			return command{}, invalid("at least one user field is required")
		}
		return rgw(args, []string{"user", "info", "--uid", uid}), nil
	case "rgw_bucket.quota":
		raw, err := base64.RawURLEncoding.DecodeString(rawText(p, "bucket_id"))
		pair := strings.SplitN(string(raw), "\x00", 2)
		if err != nil || len(pair) != 2 || pair[1] == "" || strings.HasPrefix(pair[1], "-") || strings.ContainsAny(pair[0]+pair[1], "\r\n\x00") {
			return command{}, invalid("bucket_id is invalid")
		}
		enabled, ok := p["enabled"].(bool)
		if !ok {
			return command{}, invalid("enabled must be a boolean")
		}
		target := []string{"--bucket", pair[1], "--quota-scope", "bucket"}
		if pair[0] != "" {
			target = append(target, "--tenant", pair[0])
		}
		verb := "set"
		if enabled {
			verb = "enable"
		}
		args := append([]string{"quota", verb}, target...)
		for _, field := range []string{"max_size", "max_objects"} {
			value, err := strconv.ParseInt(optional(p, field), 10, 64)
			if err != nil || value < -1 || value > 9007199254740991 {
				return command{}, invalid(field + " must be -1 or a nonnegative safe integer")
			}
			encoded := strconv.FormatInt(value, 10)
			if field == "max_size" && value >= 0 {
				encoded += "B"
			}
			args = append(args, "--"+strings.ReplaceAll(field, "_", "-"), encoded)
		}
		check := []string{"bucket", "stats", "--bucket", pair[1]}
		if pair[0] != "" {
			check = append(check, "--tenant", pair[0])
		}
		if enabled {
			return rgw(args, check), nil
		}
		result := rgw(args, nil)
		result.followups = []command{rgw(append([]string{"quota", "disable"}, target...), check)}
		return result, nil
	case "rgw_user.ratelimit", "rgw_bucket.ratelimit":
		uid := rawText(p, "uid")
		if action == "rgw_bucket.ratelimit" {
			uid = "bucket"
		}
		if !regexp.MustCompile(`^[A-Za-z0-9_.:@$-]+$`).MatchString(uid) || strings.HasPrefix(uid, "-") {
			return command{}, invalid("uid is invalid")
		}
		enabled, ok := p["enabled"].(bool)
		if !ok {
			return command{}, invalid("enabled must be a boolean")
		}
		target := []string{"--uid", uid, "--ratelimit-scope", "user"}
		if action == "rgw_bucket.ratelimit" {
			raw, err := base64.RawURLEncoding.DecodeString(rawText(p, "bucket_id"))
			pair := strings.SplitN(string(raw), "\x00", 2)
			if err != nil || len(pair) != 2 || pair[1] == "" || strings.HasPrefix(pair[1], "-") || strings.ContainsAny(pair[0]+pair[1], "\r\n\x00") {
				return command{}, invalid("bucket_id is invalid")
			}
			target = []string{"--bucket", pair[1], "--ratelimit-scope", "bucket"}
			if pair[0] != "" {
				target = append(target, "--tenant", pair[0])
			}
		}
		args := append([]string{"ratelimit", "set"}, target...)
		for _, field := range []string{"max_read_ops", "max_write_ops", "max_read_bytes", "max_write_bytes"} {
			value, err := strconv.ParseInt(optional(p, field), 10, 64)
			if err != nil || value < 0 || value > 9007199254740991 {
				return command{}, invalid(field + " must be a nonnegative safe integer")
			}
			args = append(args, "--"+strings.ReplaceAll(field, "_", "-"), strconv.FormatInt(value, 10))
		}
		verb := "disable"
		if enabled {
			verb = "enable"
		}
		result := rgw(args, nil)
		result.followups = []command{rgw(append([]string{"ratelimit", verb}, target...), append([]string{"ratelimit", "get"}, target...))}
		return result, nil
	case "rgw_user.caps":
		uid := rawText(p, "uid")
		if !regexp.MustCompile(`^[A-Za-z0-9_.:@$-]+$`).MatchString(uid) || strings.HasPrefix(uid, "-") {
			return command{}, invalid("uid is invalid")
		}
		verb, err := enum(p, "action", "add", "rm")
		if err != nil {
			return command{}, err
		}
		kind := rawText(p, "type")
		if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(kind) {
			return command{}, invalid("capability type is invalid")
		}
		permission, err := enum(p, "permission", "read", "write", "read,write", "*")
		if err != nil {
			return command{}, err
		}
		return rgw([]string{"caps", verb, "--uid", uid, "--caps", kind + "=" + permission}, []string{"user", "info", "--uid", uid}), nil
	case "rgw_user.delete":
		uid := last(tail)
		return rgw([]string{"user", "rm", "--uid", uid}, []string{"user", "list"}), nil
	case "rgw_account.quota", "rgw_user.quota":
		idField, idFlag, scopeName, noun, readVerb := "account_id", "--account-id", "account", "account", "get"
		if action == "rgw_user.quota" {
			idField, idFlag, scopeName, noun, readVerb = "uid", "--uid", "user", "user", "info"
		}
		id, err := required(p, idField)
		if action == "rgw_user.quota" {
			id = rawText(p, "uid")
			if regexp.MustCompile(`^[A-Za-z0-9_.:@$-]+$`).MatchString(id) && !strings.HasPrefix(id, "-") {
				err = nil
			} else {
				err = invalid("uid is invalid")
			}
		}
		if err != nil {
			return command{}, err
		}
		scope, err := enum(p, "scope", scopeName, "bucket")
		if err != nil {
			return command{}, err
		}
		enabled, ok := p["enabled"].(bool)
		if !ok {
			return command{}, invalid("enabled must be a boolean")
		}
		verb := "disable"
		if enabled {
			verb = "enable"
		}
		if action == "rgw_user.quota" && !enabled {
			verb = "set"
		}
		args := []string{"quota", verb, idFlag, id, "--quota-scope", scope}
		for _, field := range []string{"max_size", "max_objects"} {
			value, err := strconv.ParseInt(optional(p, field), 10, 64)
			if err != nil || value < -1 || value > 9007199254740991 {
				return command{}, invalid(field + " must be -1 or a nonnegative safe integer")
			}
			encoded := strconv.FormatInt(value, 10)
			if field == "max_size" {
				if value < 0 {
					encoded = "-1024"
				} else {
					encoded += "B"
				}
			}
			args = append(args, "--"+strings.ReplaceAll(field, "_", "-"), encoded)
		}
		check := []string{noun, readVerb, idFlag, id}
		if action == "rgw_user.quota" && !enabled {
			result := rgw(args, nil)
			result.followups = []command{rgw([]string{"quota", "disable", idFlag, id, "--quota-scope", scope}, check)}
			return result, nil
		}
		return rgw(args, check), nil
	case "rgw_account.update":
		id, err := required(p, "account_id")
		if err != nil {
			return command{}, err
		}
		args := []string{"account", "modify", "--account-id", id}
		for _, field := range []string{"account_name", "email"} {
			if _, exists := p[field]; exists {
				value := rawText(p, field)
				if value == "" {
					return command{}, invalid(field + " cannot be cleared by the native account command")
				}
				args = append(args, "--"+strings.ReplaceAll(field, "_", "-")+"="+value)
			}
		}
		for _, field := range []string{"max_users", "max_roles", "max_groups", "max_buckets", "max_access_keys"} {
			if _, exists := p[field]; exists {
				value, err := strconv.ParseInt(optional(p, field), 10, 32)
				if err != nil || value < -1 {
					return command{}, invalid(field + " must be -1 or a nonnegative integer")
				}
				args = append(args, "--"+strings.ReplaceAll(field, "_", "-"), strconv.FormatInt(value, 10))
			}
		}
		if len(args) == 4 {
			return command{}, invalid("at least one account field is required")
		}
		return rgw(args, []string{"account", "get", "--account-id", id}), nil
	case "rgw_account.delete":
		id, err := required(p, "account_id")
		if err != nil {
			return command{}, err
		}
		return rgw([]string{"account", "rm", "--account-id", id}, []string{"account", "list"}), nil
	case "rgw_account.create":
		accountID, err := required(p, "account_id")
		if err != nil {
			return command{}, err
		}
		args := []string{"account", "create", "--account-id", accountID}
		for field, flag := range map[string]string{"account_name": "--account-name", "email": "--email", "tenant": "--tenant"} {
			if value := rawText(p, field); value != "" {
				args = append(args, flag, value)
			}
		}
		return rgw(args, []string{"account", "get", "--account-id", accountID}), nil
	case "rgw_role.policy":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		policyName, err := required(p, "policy_name")
		if err != nil {
			return command{}, err
		}
		action, err := enum(p, "action", "put", "delete")
		if err != nil {
			return command{}, err
		}
		args := []string{"role-policy", action, "--role-name", name, "--policy-name", policyName}
		if action == "put" {
			policy := rawText(p, "policy_document")
			var doc map[string]any
			if json.Unmarshal([]byte(policy), &doc) != nil || doc == nil {
				return command{}, invalid("policy_document must be a JSON object")
			}
			args = append(args, "--perm-policy-doc", policy)
		}
		return rgw(args, []string{"role", "get", "--role-name", name}), nil
	case "rgw_role.update":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		var commands []command
		if _, exists := p["assume_role_policy"]; exists {
			policy := rawText(p, "assume_role_policy")
			var doc map[string]any
			if json.Unmarshal([]byte(policy), &doc) != nil || doc == nil {
				return command{}, invalid("assume_role_policy must be a JSON object")
			}
			commands = append(commands, rgw([]string{"role-trust-policy", "modify", "--role-name", name, "--assume-role-policy-doc", policy}, nil))
		}
		if _, exists := p["max_session_duration"]; exists {
			duration, err := strconv.ParseInt(optional(p, "max_session_duration"), 10, 64)
			if err != nil || duration < 3600 || duration > 43200 {
				return command{}, invalid("max_session_duration must be an integer from 3600 to 43200 seconds")
			}
			commands = append(commands, rgw([]string{"role", "update", "--role-name", name, "--max-session-duration", strconv.FormatInt(duration, 10)}, nil))
		}
		if len(commands) == 0 {
			return command{}, invalid("at least one role field is required")
		}
		commands[len(commands)-1].check = rgw(nil, []string{"role", "get", "--role-name", name}).check
		result := commands[0]
		result.followups = commands[1:]
		return result, nil

	case "rgw_role.delete":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return rgw([]string{"role", "delete", "--role-name", name}, []string{"role", "list"}), nil
	case "rgw_role.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		args := []string{"role", "create", "--role-name", name}
		if _, exists := p["max_session_duration"]; exists {
			duration, err := strconv.ParseInt(optional(p, "max_session_duration"), 10, 64)
			if err != nil || duration < 3600 || duration > 43200 {
				return command{}, invalid("max_session_duration must be an integer from 3600 to 43200 seconds")
			}
			args = append(args, "--max-session-duration", strconv.FormatInt(duration, 10))
		}
		if description := rawText(p, "description"); description != "" {
			args = append(args, "--description", description)
		}
		if path := rawText(p, "path"); path != "" {
			args = append(args, "--path", path)
		}
		policy := rawText(p, "assume_role_policy")
		var policyDocument map[string]any
		if json.Unmarshal([]byte(policy), &policyDocument) != nil || policyDocument == nil {
			return command{}, invalid("assume_role_policy must be a JSON object")
		}
		args = append(args, "--assume-role-policy-doc", policy)
		return rgw(args, []string{"role", "get", "--role-name", name}), nil
	case "rgw_key.create":
		uid := pathValue(tail, "user")
		accessKey, err := required(p, "access_key")
		if err != nil {
			return command{}, err
		}
		secretKey, err := required(p, "secret_key")
		if err != nil {
			return command{}, err
		}
		result := rgw([]string{"key", "create", "--uid", uid, "--key-type", "s3", "--access-key", accessKey, "--secret-key", secretKey}, []string{"user", "info", "--uid", uid})
		result.sensitive = map[int]struct{}{7: {}, 9: {}}
		return result, nil
	case "rgw_key.delete":
		uid := pathValue(tail, "user")
		accessKey, err := required(p, "access_key")
		if err != nil {
			return command{}, err
		}
		result := rgw([]string{"key", "rm", "--uid", uid, "--key-type", "s3", "--access-key", accessKey}, []string{"user", "info", "--uid", uid})
		result.sensitive = map[int]struct{}{7: {}}
		return result, nil
	case "rgw_zone.update":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		next, err := required(p, "new_name")
		if err != nil {
			return command{}, err
		}
		endpoints := ""
		if value, present := p["endpoints"]; present {
			var ok bool
			endpoints, ok = value.(string)
			if !ok || strings.TrimSpace(endpoints) == "" || strings.ContainsAny(endpoints, "\x00\r\n") {
				return command{}, invalid("endpoints must be a nonempty string")
			}
		}
		var syncArgs []string
		syncAll := false
		if value, present := p["sync_from_all"]; present {
			var ok bool
			syncAll, ok = value.(bool)
			if !ok {
				return command{}, invalid("sync_from_all must be a boolean")
			}
			syncArgs = append(syncArgs, "--sync-from-all="+strconv.FormatBool(syncAll))
		}
		if value, present := p["sync_from"]; present {
			sources, ok := value.(string)
			if !ok || strings.TrimSpace(sources) == "" || strings.ContainsAny(sources, "\x00\r\n") {
				return command{}, invalid("sync_from must be a nonempty string")
			}
			flag := "--sync-from"
			if syncAll {
				flag = "--sync-from-rm"
			}
			syncArgs = append(syncArgs, flag, sources)
		}
		if name == next && endpoints == "" && len(syncArgs) == 0 {
			return command{}, invalid("select a zone change")
		}
		args := []string{"zone", "rename", "--rgw-zone", name, "--zone-new-name", next}
		for _, key := range []string{"zonegroup", "realm_id"} {
			if value, present := p[key]; present {
				text, ok := value.(string)
				if !ok || strings.ContainsAny(text, "\x00\r\n") {
					return command{}, invalid(key + " must be a string")
				}
			}
		}
		if group := optional(p, "zonegroup"); group != "" {
			args = append(args, "--rgw-zonegroup", group)
		}
		check := []string{"zone", "get", "--rgw-zone", next}
		result := rgw(args, check)
		if endpoints != "" || len(syncArgs) > 0 {
			modify := []string{"zone", "modify", "--rgw-zone", next}
			if endpoints != "" {
				modify = append(modify, "--endpoints", endpoints)
			}
			modify = append(modify, syncArgs...)
			group := optional(p, "zonegroup")
			if group == "" {
				return command{}, invalid("zonegroup is required when updating zone configuration")
			}
			modify = append(modify, "--rgw-zonegroup", group)
			if name == next {
				result = rgw(modify, check)
			} else {
				result.followups = append(result.followups, rgw(modify, check))
			}
		}
		if realm := optional(p, "realm_id"); realm != "" {
			result.followups = append(result.followups, rgw([]string{"period", "update", "--commit", "--realm-id", realm}, check))
		}
		return result, nil
	case "rgw_zonegroup.update":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		newName, err := required(p, "new_name")
		if err != nil {
			return command{}, err
		}
		realmID := ""
		if value, present := p["realm_id"]; present {
			var ok bool
			realmID, ok = value.(string)
			if !ok || strings.ContainsAny(realmID, "\x00\r\n") {
				return command{}, invalid("realm_id must be a string")
			}
		}
		args := []string{"zonegroup", "modify", "--rgw-zonegroup", newName}
		changed := name != newName
		for _, key := range []string{"default", "master"} {
			if value, present := p[key]; present {
				enabled, ok := value.(bool)
				if !ok {
					return command{}, invalid(key + " must be a boolean")
				}
				if enabled {
					args = append(args, "--"+key)
					changed = true
				}
			}
		}
		if value, present := p["endpoints"]; present {
			endpoints, ok := value.(string)
			if !ok || strings.TrimSpace(endpoints) == "" || strings.ContainsAny(endpoints, "\x00\r\n") {
				return command{}, invalid("endpoints must be a nonempty string")
			}
			args = append(args, "--endpoints", endpoints)
			changed = true
		}
		var members []command
		seen := map[string]bool{}
		for _, field := range []struct{ key, action string }{{"add_zones", "add"}, {"remove_zones", "remove"}} {
			if value, present := p[field.key]; present {
				var zones []string
				switch list := value.(type) {
				case []string:
					zones = list
				case []any:
					for _, item := range list {
						zone, ok := item.(string)
						if !ok {
							return command{}, invalid(field.key + " must contain strings")
						}
						zones = append(zones, zone)
					}
				default:
					return command{}, invalid(field.key + " must be an array")
				}
				for _, zone := range zones {
					if zone == "" || !identifier.MatchString(zone) || seen[zone] {
						return command{}, invalid("zone names must be valid and unique across additions and removals")
					}
					seen[zone] = true
					members = append(members, rgw([]string{"zonegroup", field.action, "--rgw-zonegroup", newName, "--rgw-zone", zone}, []string{"zonegroup", "get", "--rgw-zonegroup", newName}))
					changed = true
				}
			}
		}
		if !changed {
			return command{}, invalid("select a zonegroup change")
		}
		check := []string{"zonegroup", "get", "--rgw-zonegroup", newName}
		result := rgw(args, check)
		if name != newName {
			result = rgw([]string{"zonegroup", "rename", "--rgw-zonegroup", name, "--zonegroup-new-name", newName}, nil)
			result.followups = append(result.followups, rgw(args, check))
		}
		result.followups = append(result.followups, members...)
		if realmID != "" {
			result.followups = append(result.followups, rgw([]string{"period", "update", "--commit", "--realm-id", realmID}, check))
		}
		return result, nil
	case "rgw_realm.update":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		newName, err := required(p, "new_name")
		if err != nil {
			return command{}, err
		}
		makeDefault := false
		if value, present := p["default"]; present {
			var ok bool
			makeDefault, ok = value.(bool)
			if !ok {
				return command{}, invalid("default must be a boolean")
			}
		}
		check := []string{"realm", "get", "--rgw-realm", newName}
		if name == newName {
			if !makeDefault {
				return command{}, invalid("change the name or select default")
			}
			return rgw([]string{"realm", "default", "--rgw-realm", name}, check), nil
		}
		result := rgw([]string{"realm", "rename", "--rgw-realm", name, "--realm-new-name", newName}, check)
		if makeDefault {
			result.followups = append(result.followups, rgw([]string{"realm", "default", "--rgw-realm", newName}, check))
		}
		return result, nil
	case "rgw_realm.create", "rgw_zonegroup.create", "rgw_zone.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		kind := strings.TrimPrefix(strings.TrimSuffix(action, ".create"), "rgw_")
		flag := "--rgw-" + kind
		args := []string{kind, "create", flag, name}
		if action == "rgw_realm.create" || action == "rgw_zonegroup.create" || action == "rgw_zone.create" {
			if value, present := p["default"]; present {
				enabled, ok := value.(bool)
				if !ok {
					return command{}, invalid("default must be a boolean")
				}
				if enabled {
					args = append(args, "--default")
				}
			}
		}
		if action == "rgw_zonegroup.create" || action == "rgw_zone.create" {
			if value, present := p["master"]; present {
				enabled, ok := value.(bool)
				if !ok {
					return command{}, invalid("master must be a boolean")
				}
				if enabled {
					args = append(args, "--master")
				}
			}
			fields := []struct{ key, flag string }{{"realm", "--rgw-realm"}, {"endpoints", "--endpoints"}}
			if action == "rgw_zone.create" {
				fields[0] = struct{ key, flag string }{"zonegroup", "--rgw-zonegroup"}
			}
			for _, field := range fields {
				if value, present := p[field.key]; present {
					text, ok := value.(string)
					if !ok || strings.TrimSpace(text) == "" || strings.ContainsAny(text, "\x00\r\n") {
						return command{}, invalid(field.key + " must be a nonempty string")
					}
					args = append(args, field.flag, text)
				}
			}
		}
		if action == "rgw_zone.create" {
			if value, present := p["tier_type"]; present {
				tier, ok := value.(string)
				if !ok || tier != "archive" {
					return command{}, invalid("tier_type must be archive")
				}
				args = append(args, "--tier-type", tier)
			}
			if value, present := p["sync_from_all"]; present {
				enabled, ok := value.(bool)
				if !ok {
					return command{}, invalid("sync_from_all must be a boolean")
				}
				args = append(args, "--sync-from-all="+strconv.FormatBool(enabled))
			}
			if value, present := p["sync_from"]; present {
				sources, ok := value.(string)
				if !ok || strings.TrimSpace(sources) == "" || strings.ContainsAny(sources, "\x00\r\n") {
					return command{}, invalid("sync_from must be a nonempty string")
				}
				args = append(args, "--sync-from", sources)
			}
		}
		result := rgw(args, []string{kind, "get", flag, name})
		if action == "rgw_zone.create" {
			_, hasAccess := p["access_key"]
			_, hasSecret := p["secret_key"]
			if hasAccess != hasSecret {
				return command{}, invalid("access_key and secret_key must be provided together")
			}
			if hasAccess {
				result.sensitive = map[int]struct{}{}
				for _, field := range []struct{ key, flag string }{{"access_key", "--access-key"}, {"secret_key", "--secret"}} {
					value, ok := p[field.key].(string)
					if !ok || value == "" || strings.ContainsAny(value, "\x00\r\n") {
						return command{}, invalid(field.key + " must be a nonempty string")
					}
					result.args = append(result.args, field.flag, value)
					result.sensitive[len(result.args)-1] = struct{}{}
				}
			}
		}
		return result, nil
	case "rgw_period.commit":
		return rgw([]string{"period", "update", "--commit"}, []string{"period", "get"}), nil
	case "nfs_cluster.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"nfs", "cluster", "create", name}, []string{"nfs", "cluster", "ls", "--format", "json"}), nil
	case "nfs_cluster.delete":
		name := last(tail)
		return ceph([]string{"nfs", "cluster", "rm", name}, []string{"nfs", "cluster", "ls", "--format", "json"}), nil
	case "nfs_export.create", "nfs_export.update":
		cluster, err := required(p, "cluster")
		if err != nil {
			return command{}, err
		}
		pseudo, err := required(p, "pseudo")
		if err != nil {
			return command{}, err
		}
		path, err := required(p, "path")
		if err != nil {
			return command{}, err
		}
		filesystem, err := required(p, "filesystem")
		if err != nil {
			return command{}, err
		}
		export := map[string]any{"cluster_id": cluster, "pseudo": pseudo, "path": path, "fsal": map[string]any{"name": "CEPH", "fs_name": filesystem}}
		if readOnly, ok := p["read_only"].(bool); ok {
			export["access_type"] = map[bool]string{true: "RO", false: "RW"}[readOnly]
		}
		stdin, _ := json.Marshal(export)
		result := ceph([]string{"nfs", "export", "apply", cluster, "-i", "-"}, []string{"nfs", "export", "ls", cluster, "--format", "json"})
		result.stdin = stdin
		return result, nil
	case "nfs_export.delete":
		cluster, pseudo, err := decodePair(last(tail))
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"nfs", "export", "rm", cluster, pseudo}, []string{"nfs", "export", "ls", cluster, "--format", "json"}), nil
	case "smb_cluster.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		authMode := optional(p, "auth_mode")
		if authMode == "" {
			authMode = "user"
		}
		if authMode != "user" && authMode != "active-directory" {
			return command{}, invalid("auth_mode is not supported")
		}
		return ceph([]string{"smb", "cluster", "create", name, authMode}, []string{"smb", "cluster", "ls", "--format", "json"}), nil
	case "smb_cluster.update":
		name := last(tail)
		authMode, err := enum(p, "auth_mode", "user", "active-directory")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"smb", "cluster", "create", name, authMode}, []string{"smb", "cluster", "ls", "--format", "json"}), nil
	case "smb_cluster.delete":
		return ceph([]string{"smb", "cluster", "rm", last(tail)}, []string{"smb", "cluster", "ls", "--format", "json"}), nil
	case "smb_share.create", "smb_share.update":
		cluster, err := required(p, "cluster")
		if err != nil {
			return command{}, err
		}
		share := optional(p, "name")
		if action == "smb_share.update" && share == "" {
			share = last(tail)
		}
		if share == "" {
			return command{}, invalid("name is required")
		}
		filesystem, err := required(p, "filesystem")
		if err != nil {
			return command{}, err
		}
		path := optional(p, "path")
		if path == "" {
			path = "/"
		}
		return ceph([]string{"smb", "share", "create", cluster, share, filesystem, path}, []string{"smb", "share", "ls", cluster, "--format", "json"}), nil
	case "smb_share.delete":
		cluster, share, err := decodePair(last(tail))
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"smb", "share", "rm", cluster, share}, []string{"smb", "share", "ls", cluster, "--format", "json"}), nil
	case "config_value.set", "config_value.delete":
		return configurationCommand(request, p)
	default:
		return command{}, unsupported(action)
	}
}

func poolCreateFollowups(
	p map[string]any,
	name, poolType string,
	wrapCeph, wrapRBD func([]string, []string) command,
) ([]command, error) {
	var commands []command
	addSet := func(field string, value string) {
		if value != "" {
			commands = append(commands, wrapCeph([]string{"osd", "pool", "set", name, field, value}, nil))
		}
	}
	if value := optional(p, "pg_autoscale_mode"); value != "" {
		if _, err := enum(p, "pg_autoscale_mode", "on", "off", "warn"); err != nil {
			return nil, err
		}
		addSet("pg_autoscale_mode", value)
	}
	if poolType == "replicated" {
		addSet("size", optional(p, "size"))
		addSet("crush_rule", optional(p, "crush_rule"))
	}
	if rawFlags, exists := p["flags"]; exists {
		flags, ok := stringSlice(rawFlags)
		if !ok {
			return nil, invalid("flags is invalid")
		}
		if poolType != "erasure" && len(flags) > 0 {
			return nil, invalid("flags are only supported for erasure pools")
		}
		seen := make(map[string]struct{}, len(flags))
		for _, flag := range flags {
			if flag != allowECOverwritesPoolFlag {
				return nil, invalid("pool flag is not supported")
			}
			if _, exists := seen[flag]; exists {
				continue
			}
			seen[flag] = struct{}{}
			addSet(flag, "true")
		}
	}
	if value := optional(p, "compression_mode"); value != "" {
		if _, err := enum(p, "compression_mode", "none", "passive", "aggressive", "force"); err != nil {
			return nil, err
		}
		addSet("compression_mode", value)
	}
	if optional(p, "compression_mode") != "none" {
		if value := optional(p, "compression_algorithm"); value != "" {
			if _, err := enum(p, "compression_algorithm", "snappy", "zlib", "zstd", "lz4"); err != nil {
				return nil, err
			}
			addSet("compression_algorithm", value)
		}
		for _, field := range []string{"compression_min_blob_size", "compression_max_blob_size"} {
			value, ok, err := optionalNonNegativeInteger(p, field)
			if err != nil {
				return nil, err
			}
			if ok {
				addSet(field, value)
			}
		}
		if value, ok, err := optionalRatio(p, "compression_required_ratio"); err != nil {
			return nil, err
		} else if ok {
			addSet("compression_required_ratio", value)
		}
	}
	if applications, ok := stringSlice(p["applications"]); ok {
		for _, application := range applications {
			commands = append(commands, wrapCeph([]string{"osd", "pool", "application", "enable", name, application}, nil))
		}
	} else if _, exists := p["applications"]; exists {
		return nil, invalid("applications is invalid")
	}
	for _, quota := range []struct {
		input string
		field string
	}{
		{input: "quota_max_bytes", field: "max_bytes"},
		{input: "quota_max_objects", field: "max_objects"},
	} {
		value, err := optionalPositiveInteger(p, quota.input)
		if err != nil {
			return nil, err
		}
		if value != "" {
			commands = append(commands, wrapCeph([]string{"osd", "pool", "set-quota", name, quota.field, value}, nil))
		}
	}
	if rawConfiguration, exists := p["configuration"]; exists {
		configuration, ok := rawConfiguration.(map[string]any)
		if !ok {
			return nil, invalid("configuration is invalid")
		}
		for _, field := range rbdPoolConfigurationFields {
			if _, exists := configuration[field]; !exists {
				continue
			}
			value, err := requiredNonNegativeInteger(configuration, field)
			if err != nil {
				return nil, err
			}
			commands = append(commands, wrapRBD([]string{"config", "pool", "set", name, field, value}, nil))
		}
	}
	if rawMirroringMode := optional(p, "rbd_mirroring"); rawMirroringMode != "" {
		mirroringMode, err := enum(p, "rbd_mirroring", "disabled", "pool")
		if err != nil {
			return nil, err
		}
		if mirroringMode != "disabled" {
			commands = append(commands, rbdMirrorPoolModeCommand(name, mirroringMode, wrapRBD))
		}
	}
	return commands, nil
}

func rbdMirrorPoolModeCommand(pool, mode string, wrapRBD func([]string, []string) command) command {
	if mode == "disabled" {
		return wrapRBD([]string{"mirror", "pool", "disable", pool}, []string{"mirror", "pool", "info", pool})
	}
	return wrapRBD([]string{"mirror", "pool", "enable", pool, mode}, []string{"mirror", "pool", "info", pool})
}

func poolSetFields() []string {
	return []string{
		"size",
		"min_size",
		"pg_num",
		"pgp_num",
		"pg_autoscale_mode",
		"crush_rule",
		"compression_mode",
		"compression_algorithm",
		"compression_required_ratio",
		"compression_max_blob_size",
		"compression_min_blob_size",
		allowECOverwritesPoolFlag,
	}
}

func validatePoolSetValue(field, value string) error {
	switch field {
	case "compression_mode":
		for _, allowed := range []string{"none", "passive", "aggressive", "force", "unset"} {
			if value == allowed {
				return nil
			}
		}
		return invalid("compression_mode is not supported")
	case "compression_algorithm":
		for _, allowed := range []string{"snappy", "zlib", "zstd", "lz4", "unset"} {
			if value == allowed {
				return nil
			}
		}
		return invalid("compression_algorithm is not supported")
	case "compression_required_ratio":
		number, err := strconv.ParseFloat(value, 64)
		if err != nil || number < 0 || number > 1 {
			return invalid("compression_required_ratio is invalid")
		}
	case "compression_min_blob_size", "compression_max_blob_size":
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil || number < 0 {
			return invalid(field + " is invalid")
		}
	case allowECOverwritesPoolFlag:
		if value != "true" && value != "false" {
			return invalid("allow_ec_overwrites is invalid")
		}
	}
	return nil
}

func optionalNonNegativeInteger(p map[string]any, key string) (string, bool, error) {
	if _, exists := p[key]; !exists {
		return "", false, nil
	}
	value := optional(p, key)
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil || number < 0 {
		return "", false, invalid(key + " is invalid")
	}
	return value, true, nil
}

func optionalRatio(p map[string]any, key string) (string, bool, error) {
	if _, exists := p[key]; !exists {
		return "", false, nil
	}
	value := optional(p, key)
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || number < 0 || number > 1 {
		return "", false, invalid(key + " is invalid")
	}
	return value, true, nil
}

func optionalPositiveInteger(p map[string]any, key string) (string, error) {
	value := optional(p, key)
	if value == "" {
		return "", nil
	}
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil || number < 0 {
		return "", invalid(key + " is invalid")
	}
	if number == 0 {
		return "", nil
	}
	return value, nil
}

func requiredNonNegativeInteger(p map[string]any, key string) (string, error) {
	value := optional(p, key)
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil || number < 0 {
		return "", invalid(key + " is invalid")
	}
	return value, nil
}

func hostUpdate(p map[string]any, host string, wrap func([]string, []string) command) (command, error) {
	if address := optional(p, "address"); address != "" {
		return wrap([]string{"orch", "host", "set-addr", host, address}, []string{"orch", "host", "ls", "--detail", "--format", "json"}), nil
	}
	labelsAdd, addOK := stringSlice(p["labels_add"])
	labelsRemove, removeOK := stringSlice(p["labels_remove"])
	if p["labels_add"] != nil && !addOK {
		return command{}, invalid("labels_add is invalid")
	}
	if p["labels_remove"] != nil && !removeOK {
		return command{}, invalid("labels_remove is invalid")
	}
	commands := make([]command, 0, len(labelsAdd)+len(labelsRemove))
	for _, label := range labelsAdd {
		commands = append(commands, wrap([]string{"orch", "host", "label", "add", host, label}, nil))
	}
	for _, label := range labelsRemove {
		commands = append(commands, wrap([]string{"orch", "host", "label", "rm", host, label}, nil))
	}
	if len(commands) == 0 {
		return command{}, invalid("at least one host update is required")
	}
	result := commands[0]
	result.followups = commands[1:]
	result.check = []string{"orch", "host", "ls", "--detail", "--format", "json"}
	return result, nil
}
func hostAction(p map[string]any, host string, wrap func([]string, []string) command) (command, error) {
	action, err := enum(p, "action", "maintenance_enter", "maintenance_exit", "drain", "stop_drain", "rescan")
	if err != nil {
		return command{}, err
	}
	var args []string
	switch action {
	case "maintenance_enter":
		args = []string{"orch", "host", "maintenance", "enter", host}
		if boolParameter(p, "force") {
			args = append(args, "--force", "--yes-i-really-mean-it")
		}
	case "maintenance_exit":
		args = []string{"orch", "host", "maintenance", "exit", host}
	case "drain":
		args = []string{"orch", "host", "drain", host}
	case "stop_drain":
		args = []string{"orch", "host", "drain", "stop", host}
	case "rescan":
		args = []string{"orch", "host", "rescan", host}
	}
	return wrap(args, []string{"orch", "host", "ls", "--detail", "--format", "json"}), nil
}

func subvolumeGroupCreateArgs(fs, name string, p map[string]any) ([]string, error) {
	pool, err := required(p, "pool")
	if err != nil {
		return nil, err
	}
	size, uid, gid := optional(p, "size"), optional(p, "uid"), optional(p, "gid")
	if size == "" {
		size = "0"
	}
	if uid == "" {
		uid = "0"
	}
	if gid == "" {
		gid = "0"
	}
	mode := optional(p, "mode")
	if mode == "" {
		mode = "0755"
	}
	if !regexp.MustCompile(`^0?[0-7]{3,4}$`).MatchString(mode) {
		return nil, invalid("mode is invalid")
	}
	args := []string{"fs", "subvolumegroup", "create", fs, name, size, pool, uid, gid, mode}
	if normalization := optional(p, "normalization"); normalization != "" {
		value, enumErr := enum(p, "normalization", "nfd", "nfc", "nfkd", "nfkc")
		if enumErr != nil {
			return nil, enumErr
		}
		args = append(args, "--normalization", value)
	}
	if boolParameter(p, "case_sensitive") {
		args = append(args, "--casesensitive")
	}
	return args, nil
}

func subvolumeCreateArgs(fs, name string, p map[string]any) ([]string, error) {
	group, err := required(p, "group")
	if err != nil {
		return nil, err
	}
	pool, err := required(p, "pool")
	if err != nil {
		return nil, err
	}
	size, uid, gid := optional(p, "size"), optional(p, "uid"), optional(p, "gid")
	if size == "" {
		size = "0"
	}
	if uid == "" {
		uid = "0"
	}
	if gid == "" {
		gid = "0"
	}
	mode := optional(p, "mode")
	if mode == "" {
		mode = "0755"
	}
	if !regexp.MustCompile(`^0?[0-7]{3,4}$`).MatchString(mode) {
		return nil, invalid("mode is invalid")
	}
	args := []string{"fs", "subvolume", "create", fs, name, size, group, pool, uid, gid, mode}
	if boolParameter(p, "namespace_isolated") {
		args = append(args, "--namespace-isolated")
	}
	return args, nil
}

func required(p map[string]any, key string) (string, error) {
	value := optional(p, key)
	if value == "" || !identifier.MatchString(value) {
		return "", invalid(key + " is required or invalid")
	}
	return value, nil
}
func optional(p map[string]any, key string) string {
	switch value := p[key].(type) {
	case string:
		return strings.TrimSpace(value)
	case json.Number:
		return value.String()
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	default:
		return ""
	}
}
func rawText(p map[string]any, key string) string {
	value, _ := p[key].(string)
	value = strings.TrimSpace(value)
	if len(value) > 32<<10 || strings.ContainsRune(value, 0) {
		return ""
	}
	return value
}
func enum(p map[string]any, key string, values ...string) (string, error) {
	value := optional(p, key)
	for _, allowed := range values {
		if value == allowed {
			return value, nil
		}
	}
	return "", invalid(key + " is not supported")
}
func stringSlice(value any) ([]string, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok || !identifier.MatchString(text) {
			return nil, false
		}
		result = append(result, text)
	}
	return result, true
}
func boolParameter(p map[string]any, key string) bool {
	value, _ := p[key].(bool)
	return value
}
func osdSpec(p map[string]any) (map[string]any, error) {
	serviceID := optional(p, "service_id")
	if serviceID == "" {
		serviceID = "default"
	}
	spec := map[string]any{"service_type": "osd", "service_id": serviceID}
	if host := optional(p, "host_pattern"); host != "" {
		spec["placement"] = map[string]any{"host_pattern": host}
	}
	dataDevices, ok := p["data_devices"].(map[string]any)
	if !ok || len(dataDevices) == 0 {
		return nil, invalid("data_devices is required")
	}
	allowed := map[string]any{}
	for _, key := range []string{"all", "paths", "rotational", "model", "vendor", "size"} {
		if value, exists := dataDevices[key]; exists {
			allowed[key] = value
		}
	}
	if len(allowed) == 0 {
		return nil, invalid("data_devices has no supported fields")
	}
	spec["data_devices"] = allowed
	return spec, nil
}
func resourceTail(value string) string { return strings.Trim(value, "/") }
func last(value string) string         { parts := strings.Split(value, "/"); return parts[len(parts)-1] }
func pathValue(path, segment string) string {
	parts := strings.Split(path, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == segment {
			return parts[i+1]
		}
	}
	return ""
}
func decodeImageSpec(value string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || !identifier.Match(decoded) {
		return "", invalid("image_spec is invalid")
	}
	if err := validateRBDImagePath(string(decoded)); err != nil {
		return "", err
	}
	return string(decoded), nil
}

func validateRBDImagePath(spec string) error {
	parts := strings.Split(spec, "/")
	if len(parts) < 2 || len(parts) > 3 {
		return invalid("image path must include pool and image")
	}
	for _, part := range parts {
		if part == "" || strings.Contains(part, "@") || !identifier.MatchString(part) {
			return invalid("image path contains an invalid component")
		}
	}
	return nil
}

func decodePair(value string) (string, string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", "", invalid("resource id is invalid")
	}
	parts := strings.SplitN(string(decoded), "\x00", 2)
	if len(parts) != 2 || !identifier.MatchString(parts[0]) || !identifier.MatchString(parts[1]) {
		return "", "", invalid("resource id is invalid")
	}
	return parts[0], parts[1], nil
}
func poolOf(spec string) string {
	if index := strings.IndexByte(spec, '/'); index >= 0 {
		return spec[:index]
	}
	return spec
}
func invalid(message string) error {
	return &cephdomain.ActionError{Code: "invalid_request", Message: message}
}
func unsupported(action string) error {
	return &cephdomain.ActionError{Code: "capability_unavailable", Message: fmt.Sprintf("native adapter for %s is unavailable", action)}
}
func normalize(err error) error {
	return &cephdomain.ActionError{Code: "ceph_command_failed", Message: err.Error(), Retryable: true}
}
