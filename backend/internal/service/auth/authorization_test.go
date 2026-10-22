package auth

import (
	"cephtower/backend/internal/store"
	"testing"
)

func TestRolePermissionsRespectScope(t *testing.T) {
	cluster := uint64(7)
	for _, tc := range []struct {
		role        string
		permissions []string
	}{
		{"viewer", []string{ReadCluster}},
		{"operator", []string{ReadCluster, OperateCluster}},
		{"storage-admin", []string{ReadCluster, OperateCluster, ManageStorage}},
		{"security-admin", []string{ReadCluster, ManageSecurity}},
		{"cluster-admin", []string{ReadCluster, OperateCluster, ManageStorage, ManageSecurity, ManageCluster}},
		{"custom", nil},
	} {
		for _, scoped := range []bool{false, true} {
			binding := store.UserRoleBinding{Role: store.Role{Name: tc.role}}
			if scoped {
				binding.ClusterID = &cluster
			}
			for _, target := range []uint64{0, 7, 8} {
				for _, permission := range []string{ReadCluster, OperateCluster, ManageStorage, ManageSecurity, ManageCluster} {
					want := false
					for _, allowed := range tc.permissions {
						if allowed == permission && (!scoped || target == cluster) {
							want = true
						}
					}
					if got := Allows([]store.UserRoleBinding{binding}, target, permission); got != want {
						t.Fatalf("role=%s scoped=%v cluster=%d permission=%s got=%v want=%v", tc.role, scoped, target, permission, got, want)
					}
				}
			}
		}
	}
	if Allows(nil, cluster, ReadCluster) {
		t.Fatal("unbound user authorized")
	}
}
