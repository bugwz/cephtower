package router

import "cephtower/backend/internal/api/v1/handler"

func rgwRoutes(h *handler.Handler) []Route {
	return []Route{
		{"GET", "/rgw/status", h.GetRGWStatus},
		{"GET", "/rgw/users", h.ListRGWUsers},
		{"POST", "/rgw/user", h.CreateRGWUser},
		{"GET", "/rgw/user", h.GetRGWUser},
		{"PATCH", "/rgw/user", h.UpdateRGWUser},
		{"DELETE", "/rgw/user", h.DeleteRGWUser},
		{"POST", "/rgw/user/key", h.CreateRGWUserKey},
		{"DELETE", "/rgw/user/key", h.DeleteRGWUserKey},
		{"GET", "/rgw/accounts", h.ListRGWAccounts},
		{"POST", "/rgw/account", h.CreateRGWAccount},
		{"GET", "/rgw/roles", h.ListRGWRoles},
		{"POST", "/rgw/role", h.CreateRGWRole},
		{"GET", "/rgw/buckets", h.ListRGWBuckets},
		{"POST", "/rgw/bucket", h.CreateRGWBucket},
		{"GET", "/rgw/bucket", h.GetRGWBucket},
		{"PATCH", "/rgw/bucket", h.UpdateRGWBucket},
		{"DELETE", "/rgw/bucket", h.DeleteRGWBucket},
		{"GET", "/rgw/bucket/policy", h.GetRGWBucketPolicy},
		{"PATCH", "/rgw/bucket/policy", h.UpdateRGWBucketPolicy},
		{"GET", "/rgw/realms", h.ListRGWRealms},
		{"POST", "/rgw/realm", h.CreateRGWRealm},
		{"GET", "/rgw/zonegroups", h.ListRGWZonegroups},
		{"POST", "/rgw/zonegroup", h.CreateRGWZonegroup},
		{"GET", "/rgw/zones", h.ListRGWZones},
		{"POST", "/rgw/zone", h.CreateRGWZone},
		{"POST", "/rgw/period/commit", h.CommitRGWPeriod},
	}
}
