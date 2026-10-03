package mutation

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type topicOwnerExecutor struct {
	raw   string
	calls []executor.CommandSpec
}

func (e *topicOwnerExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.calls = append(e.calls, spec)
	return executor.CommandResult{Stdout: []byte(e.raw)}, nil
}
func TestTopicOwnerNativeVerification(t *testing.T) {
	for _, scope := range []string{"", "team", "RGW12345678901234567"} {
		uid := "user"
		tenant, account := scope, ""
		owner := uid
		if strings.HasPrefix(scope, "RGW") {
			tenant = "team"
			account = scope
		}
		if tenant != "" {
			uid = tenant + "$ns$user"
		}
		owner = uid
		if account != "" {
			owner = account
		}
		for _, mode := range []string{"valid", "wrong uid", "wrong secret", "inactive", "duplicate", "suspended", "missing account", "subuser"} {
			t.Run(scope+mode, func(t *testing.T) {
				key := map[string]any{"user": uid, "access_key": "access", "secret_key": "sensitive-secret", "active": true}
				info := map[string]any{"full_user_id": uid, "tenant": tenant, "account_id": account, "suspended": 0, "keys": []any{key}}
				switch mode {
				case "wrong uid":
					info["full_user_id"] = "other"
				case "wrong secret":
					key["secret_key"] = "different"
				case "inactive":
					key["active"] = false
				case "duplicate":
					info["keys"] = []any{key, key}
				case "suspended":
					info["suspended"] = 1
				case "missing account":
					delete(info, "account_id")
				case "subuser":
					key["user"] = uid + ":sub"
				}
				raw, _ := json.Marshal(info)
				service, _, cluster := newCephUserService(t)
				runner := &topicOwnerExecutor{raw: string(raw)}
				service.executor = runner
				actualScope, actualOwner, err := service.VerifyTopicOwner(context.Background(), cluster, uid, "access", "sensitive-secret")
				if len(runner.calls) != 1 || runner.calls[0].Mutating || runner.calls[0].Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(runner.calls[0].Args, []string{"user", "info", "--uid=" + uid, "--format", "json"}) {
					t.Fatal("unsafe native credential lookup")
				}
				if mode == "valid" {
					if err != nil || actualScope != scope || actualOwner != owner {
						t.Fatalf("wrong owner %q %q %v", actualScope, actualOwner, err)
					}
				} else if err == nil || strings.Contains(err.Error(), "sensitive-secret") {
					t.Fatal("unverified or leaked credential")
				}
			})
		}
	}
}
