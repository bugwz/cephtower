package mutation

import "testing"

func TestTenantRoleCommandsPreserveQualifiedName(t *testing.T) {
	for _, action := range []string{"rgw_role.create", "rgw_role.update", "rgw_role.policy"} {
		for _, operation := range []string{"put", "delete"} {
			parameters := map[string]any{"name": "team$reader", "assume_role_policy": `{}`, "max_session_duration": float64(3600), "action": operation, "policy_name": "example", "policy_document": `{}`}
			cmd, err := build(Request{Action: action}, parameters)
			if err != nil {
				t.Fatalf("%s: %v", action, err)
			}
			commands := append([]command{cmd}, cmd.followups...)
			for _, command := range commands {
				for _, args := range [][]string{command.args, command.check} {
					if len(args) == 0 {
						continue
					}
					found := false
					for i, arg := range args {
						if arg == "--role-name" && i+1 < len(args) && args[i+1] == "team$reader" {
							found = true
						}
					}
					if !found {
						t.Fatalf("lost tenant identity in %s: %v", action, args)
					}
				}
			}
			for _, invalid := range []any{"team$", "$reader", "a$b$c", " reader", nil, 123} {
				parameters["name"] = invalid
				if _, err := build(Request{Action: action}, parameters); err == nil {
					t.Fatalf("%s accepted invalid name %v", action, invalid)
				}
			}
		}
	}
}
