package handler

import (
	"fmt"
	"math"
	"sort"
)

// JSONField describes the stable JSON shape accepted by a mutation action.
// It is shared by runtime validation and the OpenAPI generator.
type JSONField struct {
	Type       string
	Required   bool
	WriteOnly  bool
	Enum       []string
	Properties map[string]JSONField
	Items      *JSONField
}

type RequestContract struct {
	Required bool
	Fields   map[string]JSONField
}

func stringField(required bool, values ...string) JSONField {
	return JSONField{Type: "string", Required: required, Enum: values}
}

func boolField(required bool) JSONField { return JSONField{Type: "boolean", Required: required} }

func numberField(required bool) JSONField { return JSONField{Type: "number", Required: required} }

func integerField(required bool) JSONField { return JSONField{Type: "integer", Required: required} }

func anyField(required bool) JSONField { return JSONField{Required: required} }

func stringsField(required bool) JSONField {
	item := stringField(true)
	return JSONField{Type: "array", Required: required, Items: &item}
}

func objectField(required bool, fields map[string]JSONField) JSONField {
	return JSONField{Type: "object", Required: required, Properties: fields}
}

func objectArrayField(required bool, fields map[string]JSONField) JSONField {
	item := objectField(true, fields)
	return JSONField{Type: "array", Required: required, Items: &item}
}

var mutationRequestContracts = buildMutationRequestContracts()

func buildMutationRequestContracts() map[string]RequestContract {
	contracts := map[string]RequestContract{}
	add := func(actions []string, required bool, fields map[string]JSONField) {
		for _, action := range actions {
			copyFields := make(map[string]JSONField, len(fields)+len(routeIdentifierFields())+2)
			for name, field := range fields {
				copyFields[name] = field
			}
			copyFields["cluster_id"] = integerField(true)
			for name, field := range routeIdentifierFields() {
				if _, exists := copyFields[name]; !exists {
					copyFields[name] = field
				}
			}
			contracts[action] = RequestContract{Required: required, Fields: copyFields}
		}
	}
	empty := map[string]JSONField{}
	add([]string{"health.mute", "health.unmute", "host.delete", "service.delete", "manager.fail", "device.zap", "crush_rule.delete", "erasure_code_profile.delete", "pool.delete", "rbd_image.delete", "rbd_snapshot.delete", "rbd_namespace.delete", "rbd_trash.delete", "filesystem.delete", "subvolume_group.delete", "subvolume.delete", "cephfs_client.evict", "rgw_user.delete", "rgw_period.commit", "nfs_cluster.delete", "nfs_export.delete", "smb_cluster.delete", "smb_share.delete", "config_value.delete", "silence.delete", "rgw_bucket.delete", "iscsi_target.delete", "nvmeof_subsystem.delete"}, false, empty)
	add([]string{"cluster.refresh"}, false, map[string]JSONField{"modules": stringsField(false)})
	add([]string{"osd.delete"}, false, map[string]JSONField{"zap": boolField(false)})
	add([]string{"host.create"}, true, map[string]JSONField{"hostname": stringField(true), "address": stringField(false)})
	add([]string{"host.update"}, true, map[string]JSONField{"address": stringField(false), "label": stringField(false), "action": stringField(false, "add", "rm")})
	add([]string{"host.action"}, true, map[string]JSONField{"action": stringField(true, "maintenance_enter", "maintenance_exit", "drain", "stop_drain", "rescan")})
	add([]string{"device.identify"}, true, map[string]JSONField{"device": stringField(true), "state": stringField(true, "on", "off"), "light": stringField(false, "ident", "fault")})
	placement := objectField(false, map[string]JSONField{"count": integerField(false), "host_pattern": stringField(false), "hosts": stringsField(false), "label": stringField(false)})
	add([]string{"service.create", "service.update"}, true, map[string]JSONField{"service_type": stringField(true, "mon", "mgr", "mds", "rgw", "nfs", "smb", "prometheus", "alertmanager", "grafana", "node-exporter", "crash"), "service_id": stringField(false), "placement": placement})
	add([]string{"daemon.action"}, true, map[string]JSONField{"action": stringField(true, "start", "stop", "restart", "reconfig", "redeploy", "rotate-key")})
	add([]string{"upgrade.check"}, true, map[string]JSONField{"version": stringField(true)})
	add([]string{"upgrade.action"}, true, map[string]JSONField{"action": stringField(true, "start", "pause", "resume", "stop"), "version": stringField(false)})
	add([]string{"monitor.action"}, true, map[string]JSONField{"action": stringField(true, "scrub", "ok-to-stop"), "names": stringsField(false)})
	add([]string{"manager_module.update"}, true, map[string]JSONField{"enabled": boolField(true)})
	add([]string{"osd.action"}, true, map[string]JSONField{"action": stringField(true, "in", "out", "down", "reweight", "scrub", "deep-scrub"), "weight": numberField(false)})
	add([]string{"osd_flag.update"}, true, map[string]JSONField{"action": stringField(true, "set", "unset"), "flag": stringField(true)})
	add([]string{"osd.removal_check"}, true, map[string]JSONField{"osd_ids": stringsField(true)})
	dataDevices := objectField(true, map[string]JSONField{"all": boolField(false), "paths": stringsField(false), "rotational": boolField(false), "model": stringField(false), "vendor": stringField(false), "size": stringField(false)})
	add([]string{"osd_deployment.preview", "osd_deployment.create"}, true, map[string]JSONField{"service_id": stringField(false), "host_pattern": stringField(false), "data_devices": dataDevices})
	add([]string{"crush_rule.create"}, true, map[string]JSONField{"name": stringField(true), "root": stringField(true), "failure_domain": stringField(false), "device_class": stringField(false)})
	add([]string{"crush_rule.update"}, true, map[string]JSONField{"name": stringField(true)})
	add([]string{"erasure_code_profile.create"}, true, map[string]JSONField{"name": stringField(true), "plugin": stringField(false), "k": integerField(false), "m": integerField(false), "crush-failure-domain": stringField(false), "crush-device-class": stringField(false)})
	add([]string{"pool.create"}, true, map[string]JSONField{"name": stringField(true), "pg_num": integerField(false)})
	add([]string{"pool.update"}, true, map[string]JSONField{"operation": stringField(false, "quota", "application", "rename"), "field": stringField(false), "value": anyField(false), "action": stringField(false, "enable", "disable"), "application": stringField(false), "name": stringField(false)})
	add([]string{"rbd_image.create"}, true, map[string]JSONField{"image_spec": stringField(true), "size": integerField(true)})
	add([]string{"rbd_image.update"}, true, map[string]JSONField{"size": integerField(false), "name": stringField(false), "feature_action": stringField(false, "enable", "disable"), "features": stringField(false)})
	add([]string{"rbd_image.action"}, true, map[string]JSONField{"action": stringField(true, "flatten", "sparsify", "copy", "deep-copy", "move-to-trash"), "destination": stringField(false)})
	add([]string{"rbd_snapshot.create"}, true, map[string]JSONField{"name": stringField(true)})
	add([]string{"rbd_snapshot.update"}, true, map[string]JSONField{"name": stringField(false), "action": stringField(false, "protect", "unprotect")})
	add([]string{"rbd_snapshot.action"}, true, map[string]JSONField{"action": stringField(true, "clone", "protect", "unprotect", "rollback"), "destination": stringField(false)})
	add([]string{"rbd_namespace.create", "rbd_group.create"}, true, map[string]JSONField{"pool": stringField(true), "name": stringField(true)})
	add([]string{"rbd_trash.restore"}, true, map[string]JSONField{"pool": stringField(true), "name": stringField(true)})
	add([]string{"rbd_trash.purge"}, true, map[string]JSONField{"pool": stringField(true)})
	add([]string{"rbd_mirroring.update"}, true, map[string]JSONField{"pool": stringField(true), "mode": stringField(true, "disabled", "image", "pool")})
	add([]string{"filesystem.create", "subvolume_group.create", "subvolume.create", "cephfs_snapshot.create", "nfs_cluster.create"}, true, map[string]JSONField{"name": stringField(true)})
	add([]string{"filesystem.update"}, true, map[string]JSONField{"max_mds": integerField(true)})
	add([]string{"subvolume_group.update", "subvolume.update"}, true, map[string]JSONField{"size": integerField(true)})
	add([]string{"cephfs_snapshot.clone"}, true, map[string]JSONField{"target": stringField(true)})
	add([]string{"snapshot_schedule.create"}, true, map[string]JSONField{"path": stringField(true), "schedule": stringField(true)})
	add([]string{"cephfs_authorization.create"}, true, map[string]JSONField{"client": stringField(true), "path": stringField(false), "access": stringField(true, "r", "rw")})
	add([]string{"cephfs_entry.quota"}, true, map[string]JSONField{"path": stringField(true), "max_bytes": integerField(true)})
	add([]string{"rgw_user.create"}, true, map[string]JSONField{"uid": stringField(true), "display_name": stringField(false), "email": stringField(false)})
	add([]string{"rgw_user.update"}, true, map[string]JSONField{"display_name": stringField(false), "email": stringField(false), "max_buckets": integerField(false), "suspended": boolField(false), "system": boolField(false)})
	add([]string{"rgw_account.create"}, true, map[string]JSONField{"account_id": stringField(true), "account_name": stringField(false), "email": stringField(false), "tenant": stringField(false)})
	add([]string{"rgw_role.create"}, true, map[string]JSONField{"name": stringField(true), "path": stringField(false), "assume_role_policy": stringField(false)})
	secret := stringField(true)
	secret.WriteOnly = true
	add([]string{"rgw_key.create"}, true, map[string]JSONField{"access_key": secret, "secret_key": secret})
	add([]string{"rgw_key.delete"}, true, map[string]JSONField{"access_key": secret})
	add([]string{"rgw_realm.create", "rgw_zonegroup.create", "rgw_zone.create"}, true, map[string]JSONField{"name": stringField(true)})
	add([]string{"nfs_export.create", "nfs_export.update"}, true, map[string]JSONField{"cluster": stringField(true), "pseudo": stringField(true), "path": stringField(true), "filesystem": stringField(true), "read_only": boolField(false)})
	add([]string{"smb_cluster.create"}, true, map[string]JSONField{"name": stringField(true), "auth_mode": stringField(false, "user", "active-directory")})
	add([]string{"smb_cluster.update"}, true, map[string]JSONField{"auth_mode": stringField(true, "user", "active-directory")})
	add([]string{"smb_share.create", "smb_share.update"}, true, map[string]JSONField{"cluster": stringField(true), "name": stringField(false), "filesystem": stringField(true), "path": stringField(false)})
	add([]string{"config_value.set"}, true, map[string]JSONField{"value": stringField(true)})
	matcher := map[string]JSONField{"name": stringField(true), "value": stringField(true), "isRegex": boolField(true), "isEqual": boolField(true)}
	add([]string{"silence.create"}, true, map[string]JSONField{"matchers": objectArrayField(true, matcher), "startsAt": stringField(true), "endsAt": stringField(true), "createdBy": stringField(true), "comment": stringField(true)})
	add([]string{"rgw_bucket.create"}, true, map[string]JSONField{"name": stringField(true)})
	add([]string{"rgw_bucket.update"}, true, map[string]JSONField{"versioning": stringField(true, "enabled", "suspended")})
	add([]string{"rgw_bucket_policy.update"}, true, map[string]JSONField{"kind": stringField(false, "policy", "cors", "lifecycle", "encryption"), "document": anyField(false), "policy": anyField(false), "cors": anyField(false), "lifecycle": anyField(false), "encryption": anyField(false)})
	portal := objectArrayField(false, map[string]JSONField{"host": stringField(true), "ip": stringField(true)})
	disk := objectArrayField(false, map[string]JSONField{"pool": stringField(true), "image": stringField(true), "backstore": stringField(false)})
	initiator := objectArrayField(false, map[string]JSONField{"iqn": stringField(true), "luns": stringsField(false)})
	group := objectArrayField(false, map[string]JSONField{"name": stringField(true), "members": stringsField(false), "disks": stringsField(false)})
	add([]string{"iscsi_target.create", "iscsi_target.update"}, true, map[string]JSONField{"iqn": stringField(false), "portals": portal, "disks": disk, "clients": initiator, "groups": group})
	add([]string{"nvmeof_subsystem.create"}, true, map[string]JSONField{"subsystem_nqn": stringField(true), "serial_number": stringField(false), "model_number": stringField(false), "max_namespaces": integerField(false), "enable_ha": boolField(false), "no_group_append": boolField(false)})
	add([]string{"nvmeof_subsystem.update"}, false, empty)
	add([]string{"nvmeof_namespace.create"}, true, map[string]JSONField{"rbd_pool_name": stringField(true), "rbd_image_name": stringField(true), "block_size": integerField(false), "create_image": boolField(false), "size": integerField(false), "force": boolField(false), "no_auto_visible": boolField(false), "disable_auto_resize": boolField(false)})
	add([]string{"nvmeof_namespace.update"}, true, map[string]JSONField{"new_size": integerField(true)})
	add([]string{"nvmeof_namespace.delete"}, false, map[string]JSONField{"force": boolField(false)})
	listenerFields := map[string]JSONField{"host_name": stringField(true), "traddr": stringField(true), "adrfam": integerField(false), "trsvcid": integerField(false), "secure": boolField(false), "verify_host_name": boolField(false), "force": boolField(false)}
	add([]string{"nvmeof_listener.create"}, true, listenerFields)
	add([]string{"nvmeof_listener.delete"}, true, listenerFields)
	add([]string{"nvmeof_host.create"}, true, map[string]JSONField{"host_nqn": stringField(true), "psk": stringField(false), "dhchap_key": stringField(false)})
	add([]string{"nvmeof_host.delete"}, false, empty)
	return contracts
}

func routeIdentifierFields() map[string]JSONField {
	return map[string]JSONField{
		"account_id":    stringField(false),
		"binding_id":    integerField(false),
		"bucket_id":     stringField(false),
		"client_id":     stringField(false),
		"code":          stringField(false),
		"device_id":     stringField(false),
		"endpoint_id":   integerField(false),
		"export_id":     stringField(false),
		"fs":            stringField(false),
		"group":         stringField(false),
		"host":          stringField(false),
		"host_nqn":      stringField(false),
		"id":            stringField(false),
		"image_id":      stringField(false),
		"image_spec":    stringField(false),
		"iqn":           stringField(false),
		"listener_id":   stringField(false),
		"name":          stringField(false),
		"namespace":     stringField(false),
		"nqn":           stringField(false),
		"nsid":          stringField(false),
		"osd_id":        stringField(false),
		"pool":          stringField(false),
		"service_id":    stringField(false),
		"share_id":      stringField(false),
		"silence_id":    stringField(false),
		"snap":          stringField(false),
		"subvolume":     stringField(false),
		"subsystem_nqn": stringField(false),
		"uid":           stringField(false),
		"who":           stringField(false),
	}
}

func MutationRequestContract(action string) (RequestContract, bool) {
	contract, ok := mutationRequestContracts[action]
	return contract, ok
}

func MutationContractActions() []string {
	actions := make([]string, 0, len(mutationRequestContracts))
	for action := range mutationRequestContracts {
		actions = append(actions, action)
	}
	sort.Strings(actions)
	return actions
}

func ValidateMutationRequest(action string, value map[string]any) error {
	contract, ok := MutationRequestContract(action)
	if !ok {
		return fmt.Errorf("request contract for action %q is not registered", action)
	}
	return validateObject("request", value, contract.Fields)
}

func validateObject(path string, value map[string]any, fields map[string]JSONField) error {
	for name := range value {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("%s contains unknown field %q", path, name)
		}
	}
	for name, field := range fields {
		item, exists := value[name]
		if !exists {
			if field.Required {
				return fmt.Errorf("%s.%s is required", path, name)
			}
			continue
		}
		if item == nil {
			return fmt.Errorf("%s.%s must not be null", path, name)
		}
		if err := validateField(path+"."+name, item, field); err != nil {
			return err
		}
	}
	return nil
}

func validateField(path string, value any, field JSONField) error {
	if field.Type == "" {
		return nil
	}
	switch field.Type {
	case "string":
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s must be a string", path)
		}
		if len(field.Enum) > 0 {
			for _, allowed := range field.Enum {
				if text == allowed {
					return nil
				}
			}
			return fmt.Errorf("%s has an unsupported value", path)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s must be a boolean", path)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%s must be a number", path)
		}
	case "integer":
		number, ok := value.(float64)
		if !ok || math.Trunc(number) != number {
			return fmt.Errorf("%s must be an integer", path)
		}
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s must be an object", path)
		}
		return validateObject(path, object, field.Properties)
	case "array":
		items, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%s must be an array", path)
		}
		if field.Items != nil {
			for index, item := range items {
				if err := validateField(fmt.Sprintf("%s[%d]", path, index), item, *field.Items); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
