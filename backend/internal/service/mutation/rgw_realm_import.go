package mutation

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

var realmImportName = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

// The mgr CLI accepts a single RGWSpec on stdin. Archive imports use a separate
// native chain because the mgr CLI cannot forward tier_type.
func buildRealmImport(p map[string]any) (command, error) {
	if tier, present := p["tier_type"]; present && tier != "archive" {
		return command{}, invalid("unsupported import tier type")
	}
	if _, present := p["unmanaged"]; present {
		return command{}, invalid("unmanaged import is not supported")
	}
	token, err := parseRGWRealmToken(rawText(p, "realm_token"))
	if err != nil {
		return command{}, err
	}
	name := rawText(p, "name")
	if !realmImportName.MatchString(name) || !realmImportName.MatchString(token.RealmName) || p["confirm_import"] != true {
		return command{}, invalid("valid realm and zone names and explicit import confirmation are required")
	}
	portJSON, err := json.Marshal(p["port"])
	var port int
	if err != nil || json.Unmarshal(portJSON, &port) != nil || port < 1 || port > 65535 {
		return command{}, invalid("port must be an integer between 1 and 65535")
	}
	placement, ok := p["placement"].(map[string]any)
	if !ok || placement == nil {
		return command{}, invalid("placement must be an object")
	}
	for key := range placement {
		if key != "hosts" && key != "label" && key != "count" {
			return command{}, invalid("unsupported import placement field")
		}
	}
	if err := validateServicePlacement(placement); err != nil {
		return command{}, err
	}
	body, err := json.Marshal(map[string]any{
		"service_type": "rgw", "service_id": token.RealmName + "." + name,
		"rgw_realm_token": p["realm_token"], "rgw_zone": name,
		"rgw_frontend_port": port, "placement": placement,
	})
	if err != nil {
		return command{}, invalid("invalid import spec")
	}
	return command{binary: executor.BinaryCeph, args: []string{"rgw", "zone", "create", "-i", "-"}, stdin: body, timeout: 10 * time.Minute}, nil
}

func (s *Service) executeRealmImport(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	defer clear(spec.stdin)
	token, err := parseRGWRealmToken(rawText(request.Parameters, "realm_token"))
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	name := rawText(request.Parameters, "name")
	fail := func(after bool) (cephdomain.ActionResult, error) {
		if after {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "import_unverified", Message: "realm import may have partially changed realm, zone, period or RGW deployment; inspect both clusters before any manual retry"}
		}
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "pre_check_failed", Message: "realm identity, new zone absence or deployment absence could not be verified; no import was started"}
	}
	run := func(stage string, binary executor.Binary, args []string) ([]byte, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: binary, Args: args, Timeout: time.Minute, MaxOutput: executor.DefaultMaxOutput})
		defer clear(result.Stderr)
		if err != nil || result.ExitCode != 0 {
			clear(result.Stdout)
			return nil, false
		}
		return result.Stdout, true
	}
	readAdmin := func(stage string, args ...string) (map[string]any, bool) {
		raw, ok := run(stage, executor.BinaryRGWAdmin, append(args, "--format", "json"))
		defer clear(raw)
		doc := periodDocument(raw)
		return doc, ok && doc != nil
	}
	zones, ok := readAdmin("zones", "zone", "list")
	if !ok || !realmImportList(zones["zones"], name, false) {
		return fail(false)
	}
	realms, ok := readAdmin("realms", "realm", "list")
	if !ok {
		return fail(false)
	}
	if !realmImportList(realms["realms"], token.RealmName, false) {
		if !realmImportList(realms["realms"], token.RealmName, true) {
			return fail(false)
		}
		local, ok := readAdmin("existing_realm", "realm", "get", "--rgw-realm", token.RealmName)
		if !ok || local["id"] != token.RealmID || local["name"] != token.RealmName {
			return fail(false)
		}
	}
	serviceName := "rgw." + token.RealmName + "." + name
	serviceArgs := []string{"orch", "ls", "--service-name", serviceName, "--export", "--format", "json"}
	raw, ok := run("service_absence", executor.BinaryCeph, serviceArgs)
	absent := ok && serviceSpecAbsent(raw)
	clear(raw)
	if !absent {
		return fail(false)
	}
	archive := request.Parameters["tier_type"] == "archive"
	if archive {
		if !s.importArchiveZone(ctx, access, request, token, spec.stdin) {
			return fail(true)
		}
	} else {
		written, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action, Binary: spec.binary, Args: spec.args, Stdin: spec.stdin, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: true})
		clear(written.Stdout)
		clear(written.Stderr)
		if err != nil || written.ExitCode != 0 {
			return fail(true)
		}
	}
	realm, ok := readAdmin("realm_post_check", "realm", "get", "--rgw-realm", token.RealmName)
	if !ok || realm["id"] != token.RealmID || realm["name"] != token.RealmName || !syncFlowToken(rawText(realm, "current_period")) {
		return fail(true)
	}
	zone, ok := readAdmin("zone_post_check", "zone", "get", "--rgw-zone", name)
	if !ok || zone["name"] != name || zone["realm_id"] != token.RealmID || !syncFlowToken(rawText(zone, "id")) {
		return fail(true)
	}
	period, ok := readAdmin("period_post_check", "period", "get", "--realm-id", token.RealmID)
	group, verified := realmImportPublishedZone(period, token.RealmID, rawText(realm, "current_period"), rawText(zone, "id"), name)
	if !ok || !verified {
		return fail(true)
	}
	if archive && !archiveZonePublished(period, token.RealmID, rawText(realm, "current_period"), rawText(zone, "id"), name) {
		return fail(true)
	}
	raw, ok = run("service_post_check", executor.BinaryCeph, serviceArgs)
	defer clear(raw)
	var services []map[string]any
	if !ok || json.Unmarshal(raw, &services) != nil || len(services) != 1 || services[0]["service_name"] != serviceName || services[0]["service_type"] != "rgw" {
		return fail(true)
	}
	settings, ok := services[0]["spec"].(map[string]any)
	if !ok || settings["rgw_realm"] != token.RealmName || settings["rgw_zone"] != name || settings["rgw_zonegroup"] != group {
		return fail(true)
	}
	wantPort, _ := json.Marshal(request.Parameters["port"])
	actualPort, _ := json.Marshal(settings["rgw_frontend_port"])
	if !bytes.Equal(wantPort, actualPort) || services[0]["service_id"] != token.RealmName+"."+name || services[0]["unmanaged"] == true {
		return fail(true)
	}
	actualPlacement, _ := services[0]["placement"].(map[string]any)
	for key, expected := range request.Parameters["placement"].(map[string]any) {
		want, _ := json.Marshal(expected)
		got, _ := json.Marshal(actualPlacement[key])
		if !bytes.Equal(want, got) {
			return fail(true)
		}
	}
	if !s.verifyRealmDeployment(ctx, access, request, services[0]) {
		return fail(true)
	}
	return cephdomain.ActionResult{Details: map[string]any{"realm_id": token.RealmID, "zone_id": zone["id"], "service_name": serviceName, "deployment_submitted": true, "daemons_verified": true, "replication_verified": false}}, nil
}

func realmImportList(value any, name string, present bool) bool {
	list, ok := value.([]any)
	if !ok || list == nil {
		return false
	}
	seen := map[string]bool{}
	for _, item := range list {
		text, ok := item.(string)
		if !ok || text == "" || seen[text] {
			return false
		}
		seen[text] = true
	}
	return seen[name] == present
}

func realmImportPublishedZone(period map[string]any, realm, current, id, name string) (string, bool) {
	if period["realm_id"] != realm || period["id"] != current || !syncFlowToken(rawText(period, "master_zonegroup")) {
		return "", false
	}
	pm, _ := period["period_map"].(map[string]any)
	groups, ok := pm["zonegroups"].([]any)
	if !ok {
		return "", false
	}
	found := ""
	for _, item := range groups {
		group, ok := item.(map[string]any)
		if !ok {
			return "", false
		}
		zones, ok := group["zones"].([]any)
		if !ok {
			return "", false
		}
		for _, item := range zones {
			zone, ok := item.(map[string]any)
			if !ok {
				return "", false
			}
			if zone["id"] == id || zone["name"] == name {
				if found != "" || zone["id"] != id || zone["name"] != name || group["id"] != period["master_zonegroup"] || group["master_zone"] == id || !syncFlowToken(rawText(group, "master_zone")) || !syncFlowToken(rawText(group, "name")) {
					return "", false
				}
				found = rawText(group, "name")
			}
		}
	}
	return found, found != ""
}
