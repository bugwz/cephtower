package mutation

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

// Verify the explicit unmasked entity entry, not only an inherited effective value.
func rgwEncryptionStored(body []byte, entity, option, value string) bool {
	var rows []map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if decoder.Decode(&rows) != nil || decoder.Decode(new(any)) != io.EOF || rows == nil {
		return false
	}
	count := 0
	for _, row := range rows {
		section, sectionOK := row["section"].(string)
		name, nameOK := row["name"].(string)
		stored, valueOK := row["value"].(string)
		if !sectionOK || !nameOK || !valueOK || section == "" || name == "" {
			return false
		}
		if section != entity || name != option {
			continue
		}
		masked := false
		for _, key := range []string{"mask", "location_type", "location_value", "device_class"} {
			if raw, present := row[key]; present {
				mask, ok := raw.(string)
				if !ok {
					return false
				}
				masked = masked || mask != ""
			}
		}
		if masked {
			continue
		}
		count++
		if stored != value {
			return false
		}
	}
	return count == 1
}

func (s *Service) executeRGWEncryption(ctx context.Context, access executor.ClusterAccess, request Request) (cephdomain.ActionResult, error) {
	plan, err := planRGWEncryption(request.Parameters)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if request.ResourceKey != "rgw/encryption/"+plan.entity {
		return cephdomain.ActionResult{}, invalid("RGW encryption resource does not match the requested entity")
	}
	fail := func(code string) (cephdomain.ActionResult, error) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: code, Message: "RGW encryption configuration could not be verified; changes may be partially applied; inspect configuration before any manual retry", Retryable: false}
	}
	run := func(stage string, args []string, write bool, sensitive map[int]struct{}, inspect func([]byte) bool) bool {
		if ctx.Err() != nil {
			return false
		}
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: executor.BinaryCeph, Args: args, Mutating: write, SensitiveArgs: sensitive, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		defer func() { clear(result.Stdout); clear(result.Stderr) }()
		return err == nil && result.ExitCode == 0 && ctx.Err() == nil && inspect(result.Stdout)
	}
	read := func(stage string) (map[string]string, bool) {
		values := map[string]string{}
		for _, option := range plan.options {
			ok := run(stage+"."+option, []string{"config", "get", plan.entity, option}, false, nil, func(body []byte) bool {
				if len(body) > 16384 || !utf8.Valid(body) || !bytes.HasSuffix(body, []byte("\n")) {
					return false
				}
				value := strings.TrimSuffix(string(body), "\n")
				if strings.ContainsFunc(value, unicode.IsControl) {
					return false
				}
				values[option] = value
				return true
			})
			if !ok {
				return nil, false
			}
		}
		return values, true
	}
	baseline, ok := read("before")
	if !ok || baseline[plan.backendKey] != plan.expectedBackend {
		return fail("pre_check_failed")
	}
	for _, change := range plan.changes {
		current, ok := read("recheck_" + change.field)
		if !ok || !reflect.DeepEqual(current, baseline) {
			return fail("pre_check_failed")
		}
		if !run("set_"+change.field, change.command.args, true, change.command.sensitive, func([]byte) bool { return true }) {
			return fail("command_failed")
		}
		if !run("stored_"+change.field, []string{"config", "dump", "--format", "json"}, false, nil, func(body []byte) bool { return rgwEncryptionStored(body, plan.entity, change.option, change.value) }) {
			return fail("post_check_failed")
		}
		baseline[change.option] = change.value
		actual, ok := read("after_" + change.field)
		if !ok || !reflect.DeepEqual(actual, baseline) {
			return fail("post_check_failed")
		}
	}
	return cephdomain.ActionResult{Details: map[string]any{"entity": plan.entity, "encryption_type": plan.encryptionType, "provider": plan.provider, "fields_updated": len(plan.changes), "configuration_verified": true, "runtime_verified": false, "remote_connection_tested": false}}, nil
}
