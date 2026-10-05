package mutation

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func buildRealmSetup(p map[string]any) (command, error) {
	for _, field := range []string{"name", "zonegroup", "zone", "username"} {
		if !realmImportName.MatchString(rawText(p, field)) {
			return command{}, invalid("invalid primary site identity")
		}
	}
	if p["confirm_setup"] != true {
		return command{}, invalid("primary site setup requires explicit confirmation")
	}
	if tier, exists := p["tier_type"]; exists && tier != "archive" {
		return command{}, invalid("invalid primary zone tier")
	}
	for _, field := range []string{"zonegroup_endpoints", "zone_endpoints"} {
		values, ok := realmSetupStrings(p[field])
		if !ok || len(values) == 0 {
			return command{}, invalid("gateway endpoints are required")
		}
		for _, value := range values {
			u, err := url.Parse(value)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || u.Opaque != "" || strings.ContainsAny(value, ";,= \t\r\n") {
				return command{}, invalid("invalid gateway endpoint")
			}
			if u.Port() != "" {
				port, err := strconv.Atoi(u.Port())
				if err != nil || port < 1 || port > 65535 {
					return command{}, invalid("invalid gateway endpoint port")
				}
			}
		}
	}
	services, ok := realmSetupStrings(p["expected_services"])
	if !ok {
		return command{}, invalid("expected RGW service list is required")
	}
	for _, name := range services {
		if !strings.HasPrefix(name, "rgw.") || !realmImportName.MatchString(strings.TrimPrefix(name, "rgw.")) {
			return command{}, invalid("invalid RGW service identity")
		}
	}
	return command{}, nil
}

func realmSetupStrings(value any) ([]string, bool) {
	raw, err := json.Marshal(value)
	var values []string
	if err != nil || json.Unmarshal(raw, &values) != nil || values == nil {
		return nil, false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if value == "" || seen[value] {
			return nil, false
		}
		seen[value] = true
	}
	return values, true
}

func (s *Service) executeRealmSetup(ctx context.Context, access executor.ClusterAccess, request Request) (cephdomain.ActionResult, error) {
	p := request.Parameters
	realmName, groupName, zoneName, uid := rawText(p, "name"), rawText(p, "zonegroup"), rawText(p, "zone"), rawText(p, "username")
	groupEP, _ := realmSetupStrings(p["zonegroup_endpoints"])
	zoneEP, _ := realmSetupStrings(p["zone_endpoints"])
	services, _ := realmSetupStrings(p["expected_services"])
	slices.Sort(services)
	changed := false
	fail := func() (cephdomain.ActionResult, error) {
		if changed {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "setup_unverified", Message: "primary setup may have changed defaults, topology, pools, credentials, periods or restarted gateways; inspect the cluster before any manual retry"}
		}
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "pre_check_failed", Message: "new topology or the complete RGW restart scope could not be verified; setup was not started"}
	}
	run := func(stage string, binary executor.Binary, args []string, mutating bool) ([]byte, bool) {
		sensitive := map[int]struct{}{}
		for i, arg := range args {
			if (arg == "--access-key" || arg == "--secret") && i+1 < len(args) {
				sensitive[i+1] = struct{}{}
			}
		}
		if mutating {
			changed = true
		}
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: binary, Args: args, Mutating: mutating, SensitiveArgs: sensitive, Timeout: 5 * time.Minute, MaxOutput: executor.DefaultMaxOutput})
		defer clear(result.Stderr)
		if err != nil || result.ExitCode != 0 {
			clear(result.Stdout)
			return nil, false
		}
		return result.Stdout, true
	}
	admin := func(stage string, mutating bool, args ...string) (map[string]any, bool) {
		raw, ok := run(stage, executor.BinaryRGWAdmin, append(args, "--format", "json"), mutating)
		defer clear(raw)
		doc := periodDocument(raw)
		return doc, ok && doc != nil
	}
	serviceScope := func(stage string) bool {
		raw, ok := run(stage, executor.BinaryCeph, []string{"orch", "ls", "--service-type", "rgw", "--export", "--format", "json"}, false)
		defer clear(raw)
		if ok && len(services) == 0 && serviceSpecAbsent(raw) {
			return true
		}
		var specs []map[string]any
		if !ok || json.Unmarshal(raw, &specs) != nil || specs == nil {
			return false
		}
		names := []string{}
		seen := map[string]bool{}
		for _, spec := range specs {
			name := rawText(spec, "service_name")
			if spec["service_type"] != "rgw" || name == "" || seen[name] {
				return false
			}
			seen[name] = true
			names = append(names, name)
		}
		slices.Sort(names)
		return reflect.DeepEqual(names, services)
	}
	for _, target := range []struct{ kind, name string }{{"realm", realmName}, {"zonegroup", groupName}, {"zone", zoneName}} {
		doc, ok := admin(target.kind+"_absence", false, target.kind, "list")
		if !ok || !realmImportList(doc[target.kind+"s"], target.name, false) {
			return fail()
		}
	}
	if !serviceScope("services_before") {
		return fail()
	}
	hosts, ok := run("hosts", executor.BinaryCeph, []string{"orch", "host", "ls", "--format", "json"}, false)
	resolved, resolveErr := resolveRealmSetupEndpoints(hosts, p)
	clear(hosts)
	if !ok || resolveErr != nil {
		return fail()
	}
	p = resolved
	groupEP, _ = realmSetupStrings(p["zonegroup_endpoints"])
	zoneEP, _ = realmSetupStrings(p["zone_endpoints"])
	realm, ok := admin("realm_create", true, "realm", "create", "--rgw-realm", realmName, "--default")
	realmID := rawText(realm, "id")
	if !ok || realm["name"] != realmName || !syncFlowToken(realmID) {
		return fail()
	}
	group, ok := admin("group_create", true, "zonegroup", "create", "--realm-id", realmID, "--rgw-zonegroup", groupName, "--master", "--default", "--endpoints", strings.Join(groupEP, ","))
	groupID := rawText(group, "id")
	if !ok || group["name"] != groupName || group["realm_id"] != realmID || !syncFlowToken(groupID) {
		return fail()
	}
	args := []string{"zone", "create", "--realm-id", realmID, "--zonegroup-id", groupID, "--rgw-zone", zoneName, "--master", "--default", "--endpoints", strings.Join(zoneEP, ",")}
	if p["tier_type"] == "archive" {
		args = append(args, "--tier-type", "archive")
	}
	zone, ok := admin("zone_create", true, args...)
	zoneID := rawText(zone, "id")
	if !ok || zone["name"] != zoneName || zone["realm_id"] != realmID || !syncFlowToken(zoneID) {
		return fail()
	}
	commitArgs := []string{"period", "update", "--commit", "--realm-id", realmID, "--zonegroup-id", groupID, "--zone-id", zoneID}
	initial, ok := admin("initial_commit", true, commitArgs...)
	if !ok || !realmSetupPeriod(initial, p, realmID, groupID, zoneID) {
		return fail()
	}
	// A new zone has its own user metadata; never reuse an existing system user.
	raw, ok := run("user_absence", executor.BinaryRGWAdmin, []string{"user", "list", "--zone-id", zoneID, "--format", "json"}, false)
	var users []any
	valid := ok && json.Unmarshal(raw, &users) == nil && realmImportList(users, uid, false)
	clear(raw)
	if !valid {
		return fail()
	}
	user, ok := admin("user_create", true, "user", "create", "--zone-id", zoneID, "--uid", uid, "--display-name", uid, "--system")
	accessKey, secret, valid := realmSetupUserKey(user, uid)
	if !ok || !valid {
		return fail()
	}
	if _, ok := admin("zone_credentials", true, "zone", "modify", "--realm-id", realmID, "--zonegroup-id", groupID, "--zone-id", zoneID, "--access-key", accessKey, "--secret", secret); !ok {
		return fail()
	}
	committed, ok := admin("commit", true, commitArgs...)
	if !ok || !realmSetupPeriod(committed, p, realmID, groupID, zoneID) {
		return fail()
	}
	realm, ok = admin("realm_check", false, "realm", "get", "--realm-id", realmID)
	if !ok || realm["id"] != realmID || realm["name"] != realmName || realm["current_period"] != committed["id"] {
		return fail()
	}
	period, ok := admin("period_check", false, "period", "get", "--realm-id", realmID)
	if !ok || !reflect.DeepEqual(period, committed) {
		return fail()
	}
	for _, target := range []struct{ kind, id string }{{"realm", realmID}, {"zonegroup", groupID}, {"zone", zoneID}} {
		doc, ok := admin(target.kind+"_default", false, target.kind, "list")
		if !ok || doc["default_info"] != target.id {
			return fail()
		}
	}
	zone, ok = admin("zone_check", false, "zone", "get", "--zone-id", zoneID)
	key, _ := zone["system_key"].(map[string]any)
	if !ok || zone["id"] != zoneID || zone["name"] != zoneName || zone["realm_id"] != realmID || key["access_key"] != accessKey || key["secret_key"] != secret {
		return fail()
	}
	user, ok = admin("user_check", false, "user", "info", "--zone-id", zoneID, "--uid", uid)
	a, b, valid := realmSetupUserKey(user, uid)
	if !ok || !valid || a != accessKey || b != secret {
		return fail()
	}
	if !serviceScope("services_check") {
		return fail()
	}
	for _, name := range services {
		restartCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		readDaemons := func(stage string) (map[string]setupDaemon, bool) {
			result, err := s.executor.Run(restartCtx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: executor.BinaryCeph, Args: []string{"orch", "ps", "--service-name", name, "--daemon-type", "rgw", "--refresh", "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
			defer clear(result.Stdout)
			defer clear(result.Stderr)
			if err != nil || result.ExitCode != 0 {
				return nil, false
			}
			return setupDaemonSnapshot(result.Stdout, name)
		}
		before, valid := readDaemons("daemons_before")
		if !valid || !setupRestartBaseline(before) {
			cancel()
			return fail()
		}
		restarted, restartErr := s.executor.Run(restartCtx, access, executor.CommandSpec{ID: request.Action + ".restart", Binary: executor.BinaryCeph, Args: []string{"orch", "restart", name}, Mutating: true, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		clear(restarted.Stdout)
		clear(restarted.Stderr)
		if restartErr != nil || restarted.ExitCode != 0 {
			cancel()
			return fail()
		}
		ready := waitSetupRestart(restartCtx, time.Second, before, func() (map[string]setupDaemon, bool) { return readDaemons("daemons_ready") })
		cancel()
		if !ready {
			return fail()
		}
	}
	return cephdomain.ActionResult{Details: map[string]any{"realm_id": realmID, "zonegroup_id": groupID, "zone_id": zoneID, "zonegroup_endpoints": groupEP, "zone_endpoints": zoneEP, "restart_submitted": services, "daemons_verified": len(services) > 0, "replication_verified": false}}, nil
}

func realmSetupUserKey(user map[string]any, uid string) (string, string, bool) {
	if user["user_id"] != uid || user["system"] != true {
		return "", "", false
	}
	keys, ok := user["keys"].([]any)
	if !ok || len(keys) != 1 {
		return "", "", false
	}
	key, ok := keys[0].(map[string]any)
	if !ok || key["user"] != uid {
		return "", "", false
	}
	a, b := rawText(key, "access_key"), rawText(key, "secret_key")
	return a, b, a != "" && b != ""
}

func realmSetupPeriod(period map[string]any, p map[string]any, realm, groupID, zoneID string) bool {
	epoch, ok := period["epoch"].(json.Number)
	n, err := epoch.Int64()
	if !ok || err != nil || n <= 0 || !syncFlowToken(rawText(period, "id")) || period["realm_id"] != realm || period["master_zone"] != zoneID {
		return false
	}
	group, zone, ok := archiveMaster(period)
	if !ok || group["id"] != groupID || group["name"] != p["zonegroup"] || zone["id"] != zoneID || zone["name"] != p["zone"] {
		return false
	}
	for _, pair := range []struct {
		actual any
		field  string
	}{{group["endpoints"], "zonegroup_endpoints"}, {zone["endpoints"], "zone_endpoints"}} {
		actual, ok := realmSetupStrings(pair.actual)
		expected, _ := realmSetupStrings(p[pair.field])
		if !ok {
			return false
		}
		slices.Sort(actual)
		slices.Sort(expected)
		if !reflect.DeepEqual(actual, expected) {
			return false
		}
	}
	if p["tier_type"] == "archive" {
		return zone["tier_type"] == "archive"
	}
	return zone["tier_type"] == ""
}
