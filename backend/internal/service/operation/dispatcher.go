package operation

import (
	"context"

	cephdomain "cephtower/backend/internal/domain/ceph"
	externalservice "cephtower/backend/internal/service/external"
	mutationservice "cephtower/backend/internal/service/mutation"
)

type mutationExecutor interface {
	Execute(context.Context, mutationservice.Request) (cephdomain.ActionResult, error)
}

type externalExecutor interface {
	Execute(context.Context, externalservice.Request) (cephdomain.ActionResult, error)
}

type reconcileExecutor interface {
	Refresh(context.Context, uint64, []string) (cephdomain.ActionResult, error)
	RefreshKinds(context.Context, uint64, []string) (cephdomain.ActionResult, error)
	RefreshKindIfSupported(context.Context, uint64, string) (bool, error)
}

type ActionDispatcher struct {
	mutations  mutationExecutor
	external   externalExecutor
	reconciler reconcileExecutor
}

func NewActionDispatcher(mutations mutationExecutor, external externalExecutor, reconciler reconcileExecutor) *ActionDispatcher {
	return &ActionDispatcher{mutations: mutations, external: external, reconciler: reconciler}
}

func (d *ActionDispatcher) Execute(ctx context.Context, request ExecutionRequest) (cephdomain.ActionResult, error) {
	if request.Action == "cluster.refresh" {
		if d.reconciler == nil {
			return unavailable("resource refresh is unavailable")
		}
		kinds := stringSlice(request.Parameters["kinds"])
		if kind, _ := request.Parameters["kind"].(string); kind != "" {
			kinds = append(kinds, kind)
		}
		if len(kinds) > 0 {
			return d.reconciler.RefreshKinds(ctx, request.ClusterID, kinds)
		}
		modules := stringSlice(request.Parameters["modules"])
		if module, _ := request.Parameters["module"].(string); module != "" {
			modules = append(modules, module)
		}
		return d.reconciler.Refresh(ctx, request.ClusterID, modules)
	}
	if externalservice.Supports(request.Action) {
		if d.external == nil {
			return unavailable("external action is unavailable")
		}
		return d.external.Execute(ctx, externalservice.Request{
			ClusterID: request.ClusterID, Action: request.Action,
			ResourceKey: request.ResourceKey, Parameters: request.Parameters,
		})
	}
	if d.mutations == nil {
		return unavailable("native action is unavailable")
	}
	result, err := d.mutations.Execute(ctx, mutationservice.Request{
		ClusterID: request.ClusterID, Action: request.Action,
		ResourceKey: request.ResourceKey, Parameters: request.Parameters,
	})
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if request.Action == "osd_deployment.preview" || d.reconciler == nil {
		return result, nil
	}
	var refreshed bool
	_, migration := request.Parameters["target_account_id"]
	if request.Action == "rgw_encryption.update" {
		_, err = d.reconciler.RefreshKinds(ctx, request.ClusterID, []string{"config_value"})
		refreshed = err == nil
	} else if request.Action == "rgw_realm.setup" || request.Action == "rgw_realm.migrate" {
		_, err = d.reconciler.RefreshKinds(ctx, request.ClusterID, []string{"rgw_realm", "rgw_zonegroup", "rgw_zone", "rgw_user", "pool", "service"})
		refreshed = err == nil
	} else if request.Action == "rgw_zonegroup.cloud_create" || request.Action == "rgw_zonegroup.cloud_connection" || request.Action == "rgw_zonegroup.cloud_target" || request.Action == "rgw_zonegroup.cloud_acl" || request.Action == "rgw_zonegroup.cloud_restore" || request.Action == "rgw_zonegroup.delete" || request.Action == "rgw_zone.delete" || request.Action == "rgw_zone.placement" || request.Action == "rgw_zone.placement_create" || request.Action == "rgw_zonegroup.storage_class_delete" || request.Action == "rgw_zonegroup.storage_class_delete_local" {
		_, err = d.reconciler.RefreshKinds(ctx, request.ClusterID, []string{"rgw_realm", "rgw_zonegroup", "rgw_zone"})
		refreshed = err == nil
	} else if request.Action == "rgw_realm.import" {
		_, err = d.reconciler.RefreshKinds(ctx, request.ClusterID, []string{"rgw_realm", "rgw_zonegroup", "rgw_zone", "service"})
		refreshed = err == nil
	} else if request.Action == "rgw_user.update" && migration {
		_, err = d.reconciler.RefreshKinds(ctx, request.ClusterID, []string{"rgw_user", "rgw_account", "rgw_bucket"})
		refreshed = err == nil
	} else if account, _ := request.Parameters["account_id"].(string); request.Action == "rgw_user.create" && account != "" {
		_, err = d.reconciler.RefreshKinds(ctx, request.ClusterID, []string{"rgw_user", "rgw_account"})
		refreshed = err == nil
	} else if request.Action == "rgw_key.create" || request.Action == "rgw_key.update" || request.Action == "rgw_key.delete" {
		refreshed, err = d.reconciler.RefreshKindIfSupported(ctx, request.ClusterID, "rgw_user")
	} else if request.Action == "filesystem.rename" {
		_, err = d.reconciler.RefreshKinds(ctx, request.ClusterID, []string{"filesystem", "pool"})
		refreshed = err == nil
	} else {
		refreshed, err = d.reconciler.RefreshKindIfSupported(ctx, request.ClusterID, request.ResourceKind)
	}
	if err != nil {
		if request.Action == "rgw_encryption.update" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "encryption configuration verified but inventory refresh failed; refresh without repeating the mutation", Retryable: false}
		}
		if request.Action == "rgw_zonegroup.cloud_create" || request.Action == "rgw_zonegroup.cloud_connection" || request.Action == "rgw_zonegroup.cloud_target" || request.Action == "rgw_zonegroup.cloud_acl" || request.Action == "rgw_zonegroup.cloud_restore" || request.Action == "rgw_zonegroup.placement_default" || request.Action == "rgw_zonegroup.placement_tags" || request.Action == "rgw_zonegroup.storage_class_delete" || request.Action == "rgw_zonegroup.storage_class_delete_local" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "placement configuration verified but inventory refresh failed; refresh without repeating the mutation", Retryable: false}
		}
		if request.Action == "rgw_zone.placement" || request.Action == "rgw_zone.placement_create" || request.Action == "rgw_zone.storage_class_create" || request.Action == "rgw_zonegroup.storage_class_create" || request.Action == "rgw_zonegroup.placement_create" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "placement update verified but inventory refresh failed; refresh without repeating the mutation", Retryable: false}
		}
		if request.Action == "rgw_realm.setup" || request.Action == "rgw_realm.migrate" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "primary setup completed but inventory refresh failed; refresh inventory without repeating setup", Retryable: false}
		}
		if request.Action == "rgw_realm.delete" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "Realm absence was verified but inventory refresh failed; refresh without repeating deletion", Retryable: false}
		}
		if request.Action == "rgw_zonegroup.delete" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "Zonegroup removal was verified but inventory refresh failed; refresh without repeating deletion", Retryable: false}
		}
		if request.Action == "rgw_zone.delete" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "Zone removal was verified but inventory refresh failed; refresh without repeating deletion", Retryable: false}
		}
		if request.Action == "rgw_realm.import" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "import was verified but inventory refresh failed; refresh inventory without repeating import", Retryable: false}
		}
		if request.Action == "rgw_user.caps" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "capability change was verified but inventory refresh failed; refresh user inventory without repeating the change", Retryable: false}
		}
		if request.Action == "rgw_user.delete" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "user removal was verified but inventory refresh failed; refresh user inventory without repeating removal", Retryable: false}
		}
		if request.Action == "rgw_user.create" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "user creation completed but inventory refresh failed; refresh user and account inventory without repeating creation", Retryable: false}
		}
		if request.Action == "rgw_user.subuser" || request.Action == "rgw_key.create" || request.Action == "rgw_key.update" || request.Action == "rgw_key.delete" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "user credential change was verified but inventory refresh failed; refresh user inventory without repeating the change", Retryable: false}
		}
		if request.Action == "rgw_user.update" && migration {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "account migration was verified but inventory refresh failed; refresh user, account and bucket inventory without repeating migration", Retryable: false}
		}
		if request.Action == "rgw_user.update" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "user update was verified but inventory refresh failed; refresh user inventory without repeating the update", Retryable: false}
		}
		if request.Action == "rbd_mirroring.global_schedule" {
			return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_reconcile_failed", Message: "global schedule was verified but inventory refresh failed; refresh inventory before another change", Retryable: false}
		}
		return cephdomain.ActionResult{}, &cephdomain.ActionError{
			Code: "post_reconcile_failed", Message: "command succeeded but the cached state could not be refreshed", Retryable: true,
		}
	}
	if refreshed {
		if details, ok := result.Details.(map[string]any); ok {
			details["reconciled"] = true
		}
	}
	return result, nil
}

func unavailable(message string) (cephdomain.ActionResult, error) {
	return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "capability_unavailable", Message: message}
}

func stringSlice(value any) []string {
	switch values := value.(type) {
	case []string:
		return append([]string(nil), values...)
	case []any:
		result := make([]string, 0, len(values))
		for _, value := range values {
			if text, ok := value.(string); ok && text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}
