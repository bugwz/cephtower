package auth

import "cephtower/backend/internal/store"

// Permissions are evaluated against global bindings and the requested cluster.
const (
	ReadCluster    = "cluster:read"
	ManageCluster  = "cluster:write"
	ManageStorage  = "storage:write"
	OperateCluster = "cluster:operate"
	ManageSecurity = "security:write"
)

func Allows(bindings []store.UserRoleBinding, clusterID uint64, permission string) bool {
	for _, binding := range bindings {
		if binding.ClusterID != nil && (clusterID == 0 || *binding.ClusterID != clusterID) {
			continue
		}
		switch binding.Role.Name {
		case "cluster-admin":
			return true
		case "security-admin":
			if permission == ReadCluster || permission == ManageSecurity {
				return true
			}
		case "storage-admin":
			if permission == ReadCluster || permission == ManageStorage || permission == OperateCluster {
				return true
			}
		case "operator":
			if permission == ReadCluster || permission == OperateCluster {
				return true
			}
		case "viewer":
			if permission == ReadCluster {
				return true
			}
		}
	}
	return false
}

// GlobalRole is a display label; authorization always evaluates every binding.
func GlobalRole(bindings []store.UserRoleBinding) string {
	for _, role := range []string{"cluster-admin", "security-admin", "storage-admin", "operator", "viewer"} {
		for _, binding := range bindings {
			if binding.ClusterID == nil && binding.Role.Name == role {
				return role
			}
		}
	}
	return "viewer"
}
