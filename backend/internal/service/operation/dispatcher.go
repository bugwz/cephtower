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
		return d.reconciler.Refresh(ctx, request.ClusterID, stringSlice(request.Parameters["modules"]))
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
	refreshed, err := d.reconciler.RefreshKindIfSupported(ctx, request.ClusterID, request.ResourceKind)
	if err != nil {
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
