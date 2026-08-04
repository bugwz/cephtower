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
	clusterservice "cephtower/backend/internal/service/cluster"
)

var identifier = regexp.MustCompile(`^[A-Za-z0-9_.:@/+\-=]{1,512}$`)

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
	case "cluster.refresh", "health.mute", "health.unmute",
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
		"rbd_trash.purge", "rbd_group.create", "rbd_mirroring.update",
		"filesystem.create", "filesystem.update", "filesystem.delete",
		"subvolume_group.create", "subvolume_group.update", "subvolume_group.delete",
		"subvolume.create", "subvolume.update", "subvolume.delete",
		"cephfs_snapshot.create", "cephfs_snapshot.clone", "snapshot_schedule.create",
		"cephfs_authorization.create", "cephfs_client.evict", "cephfs_entry.quota",
		"rgw_user.create", "rgw_user.update", "rgw_user.delete",
		"rgw_account.create", "rgw_role.create", "rgw_key.create", "rgw_key.delete",
		"rgw_realm.create", "rgw_zonegroup.create", "rgw_zone.create", "rgw_period.commit",
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
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action, Binary: spec.binary, Args: spec.args, Stdin: spec.stdin, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: true, SensitiveArgs: spec.sensitive})
	if err != nil {
		return cephdomain.ActionResult{}, normalize(err)
	}
	for index, followup := range spec.followups {
		stepID := fmt.Sprintf("%s.step%d", request.Action, index+2)
		result, err = s.executor.Run(ctx, access, executor.CommandSpec{ID: stepID, Binary: followup.binary, Args: followup.args, Stdin: followup.stdin, Timeout: followup.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: true, SensitiveArgs: followup.sensitive})
		if err != nil {
			return cephdomain.ActionResult{}, normalize(err)
		}
	}
	if len(spec.check) > 0 {
		if _, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + ".post_check", Binary: spec.binary, Args: spec.check, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput}); err != nil {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "command was accepted but the expected state could not be verified", Retryable: true}
		}
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
		return command{binary: executor.BinaryRBD, args: append(args, "--format", "json"), check: append(check, "--format", "json"), timeout: 5 * time.Minute}
	}
	rgw := func(args, check []string) command {
		return command{binary: executor.BinaryRGWAdmin, args: append(args, "--format", "json"), check: append(check, "--format", "json"), timeout: 2 * time.Minute}
	}
	cephfsShell := func(args []string) command {
		return command{binary: executor.BinaryCephFSShell, args: args, timeout: 2 * time.Minute}
	}
	switch action {
	case "cluster.refresh":
		return ceph([]string{"status", "--format", "json"}, []string{"status", "--format", "json"}), nil
	case "health.mute":
		return ceph([]string{"health", "mute", last(tail)}, []string{"health", "detail", "--format", "json"}), nil
	case "health.unmute":
		return ceph([]string{"health", "unmute", last(tail)}, []string{"health", "detail", "--format", "json"}), nil
	case "host.create":
		name, err := required(p, "hostname")
		if err != nil {
			return command{}, err
		}
		args := []string{"orch", "host", "add", name}
		if addr := optional(p, "address"); addr != "" {
			args = append(args, addr)
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
		for _, key := range []string{"plugin", "k", "m", "crush-failure-domain", "crush-device-class"} {
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
		result.followups, err = poolCreateFollowups(p, name, poolType, ceph)
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
			return ceph([]string{"osd", "pool", "application", verb, name, application}, []string{"osd", "pool", "application", "get", name, "--format", "json"}), nil
		}
		if operation == "rename" {
			newName, err := required(p, "name")
			if err != nil {
				return command{}, err
			}
			return ceph([]string{"osd", "pool", "rename", name, newName}, []string{"osd", "pool", "ls", "detail", "--format", "json"}), nil
		}
		field, err := enum(p, "field", "size", "min_size", "pg_num", "pgp_num", "pg_autoscale_mode", "crush_rule", "compression_mode")
		if err != nil {
			return command{}, err
		}
		value, err := required(p, "value")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"osd", "pool", "set", name, field, value}, []string{"osd", "pool", "ls", "detail", "--format", "json"}), nil
	case "pool.delete":
		name := last(tail)
		return ceph([]string{"osd", "pool", "rm", name, name, "--yes-i-really-really-mean-it"}, []string{"osd", "pool", "ls", "detail", "--format", "json"}), nil
	case "rbd_image.create":
		spec, err := required(p, "image_spec")
		if err != nil {
			return command{}, err
		}
		size, err := required(p, "size")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"create", spec, "--size", size}, []string{"info", spec}), nil
	case "rbd_image.update":
		spec, err := decodeImageSpec(last(tail))
		if err != nil {
			return command{}, err
		}
		if size := optional(p, "size"); size != "" {
			return rbd([]string{"resize", spec, "--size", size}, []string{"info", spec}), nil
		}
		if name := optional(p, "name"); name != "" {
			destination := poolOf(spec) + "/" + name
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
		return rbd([]string{"rm", spec}, []string{"ls", poolOf(spec)}), nil
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
		verb, err := enum(p, "action", "flatten", "sparsify", "copy", "deep-copy", "move-to-trash")
		if err != nil {
			return command{}, err
		}
		if verb == "move-to-trash" {
			return rbd([]string{"trash", "mv", spec}, []string{"trash", "ls", poolOf(spec)}), nil
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
			args = append(args, destination)
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
		return rbd([]string{"trash", "remove", pool + "/" + imageID, "--force"}, []string{"trash", "ls", pool}), nil
	case "rbd_trash.purge":
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"trash", "purge", pool}, []string{"trash", "ls", pool}), nil
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
	case "rbd_mirroring.update":
		pool, err := required(p, "pool")
		if err != nil {
			return command{}, err
		}
		mode, err := enum(p, "mode", "disabled", "image", "pool")
		if err != nil {
			return command{}, err
		}
		return rbd([]string{"mirror", "pool", "enable", pool, mode}, []string{"mirror", "pool", "info", pool}), nil
	case "filesystem.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"fs", "volume", "create", name}, []string{"fs", "volume", "ls", "--format", "json"}), nil
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
		return ceph([]string{"fs", "subvolumegroup", "create", fs, name}, []string{"fs", "subvolumegroup", "info", fs, name, "--format", "json"}), nil
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
		return ceph([]string{"fs", "subvolume", "create", fs, name}, []string{"fs", "subvolume", "info", fs, name, "--format", "json"}), nil
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
	case "cephfs_snapshot.clone":
		fs := pathValue(tail, "filesystem")
		subvolume := pathValue(tail, "subvolume")
		snapshot := pathValue(tail, "snapshot")
		target, err := required(p, "target")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"fs", "subvolume", "snapshot", "clone", fs, subvolume, snapshot, target}, []string{"fs", "clone", "status", fs, target, "--format", "json"}), nil
	case "snapshot_schedule.create":
		fs := pathValue(tail, "filesystem")
		path, err := required(p, "path")
		if err != nil {
			return command{}, err
		}
		schedule, err := required(p, "schedule")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"fs", "snap-schedule", "add", path, schedule, fs}, []string{"fs", "snap-schedule", "list", path, "--format", "json"}), nil
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
		args := []string{"user", "create", "--uid", uid}
		for field, flag := range map[string]string{"display_name": "--display-name", "email": "--email"} {
			if value := rawText(p, field); value != "" {
				args = append(args, flag, value)
			}
		}
		return rgw(args, []string{"user", "info", "--uid", uid}), nil
	case "rgw_user.update":
		uid := last(tail)
		args := []string{"user", "modify", "--uid", uid}
		for field, flag := range map[string]string{"display_name": "--display-name", "email": "--email", "max_buckets": "--max-buckets"} {
			if value := rawText(p, field); value != "" {
				args = append(args, flag, value)
			}
		}
		for field, flag := range map[string]string{"suspended": "--suspended", "system": "--system"} {
			if value, ok := p[field].(bool); ok {
				args = append(args, flag, strconv.FormatBool(value))
			}
		}
		if len(args) == 4 {
			return command{}, invalid("at least one user field is required")
		}
		return rgw(args, []string{"user", "info", "--uid", uid}), nil
	case "rgw_user.delete":
		uid := last(tail)
		return rgw([]string{"user", "rm", "--uid", uid}, []string{"user", "list"}), nil
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
	case "rgw_role.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		args := []string{"role", "create", "--role-name", name}
		if path := rawText(p, "path"); path != "" {
			args = append(args, "--path", path)
		}
		if policy := rawText(p, "assume_role_policy"); policy != "" {
			args = append(args, "--assume-role-policy-doc", policy)
		}
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
	case "rgw_realm.create", "rgw_zonegroup.create", "rgw_zone.create":
		name, err := required(p, "name")
		if err != nil {
			return command{}, err
		}
		kind := strings.TrimPrefix(strings.TrimSuffix(action, ".create"), "rgw_")
		flag := "--rgw-" + kind
		return rgw([]string{kind, "create", flag, name}, []string{kind, "get", flag, name}), nil
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
	case "config_value.set":
		who := pathValue(tail, "value")
		name := last(tail)
		value, err := required(p, "value")
		if err != nil {
			return command{}, err
		}
		return ceph([]string{"config", "set", who, name, value}, []string{"config", "get", who, name, "--format", "json"}), nil
	case "config_value.delete":
		segments := strings.Split(tail, "/")
		if len(segments) < 3 {
			return command{}, invalid("invalid configuration resource")
		}
		return ceph([]string{"config", "rm", segments[len(segments)-2], segments[len(segments)-1]}, []string{"config", "dump", "--format", "json"}), nil
	default:
		return command{}, unsupported(action)
	}
}

func poolCreateFollowups(p map[string]any, name, poolType string, wrap func([]string, []string) command) ([]command, error) {
	var commands []command
	addSet := func(field string, value string) {
		if value != "" {
			commands = append(commands, wrap([]string{"osd", "pool", "set", name, field, value}, nil))
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
	if value := optional(p, "compression_mode"); value != "" {
		if _, err := enum(p, "compression_mode", "none", "passive", "aggressive", "force"); err != nil {
			return nil, err
		}
		addSet("compression_mode", value)
	}
	if applications, ok := stringSlice(p["applications"]); ok {
		for _, application := range applications {
			commands = append(commands, wrap([]string{"osd", "pool", "application", "enable", name, application}, nil))
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
			commands = append(commands, wrap([]string{"osd", "pool", "set-quota", name, quota.field, value}, nil))
		}
	}
	return commands, nil
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

func hostUpdate(p map[string]any, host string, wrap func([]string, []string) command) (command, error) {
	if address := optional(p, "address"); address != "" {
		return wrap([]string{"orch", "host", "set-addr", host, address}, []string{"orch", "host", "ls", "--detail", "--format", "json"}), nil
	}
	label, err := required(p, "label")
	if err != nil {
		return command{}, err
	}
	action, err := enum(p, "action", "add", "rm")
	if err != nil {
		return command{}, err
	}
	return wrap([]string{"orch", "host", "label", action, host, label}, []string{"orch", "host", "ls", "--detail", "--format", "json"}), nil
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
	return string(decoded), nil
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
