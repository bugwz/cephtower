package clusterinspect

import (
	"context"
	"regexp"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

type UpgradeVersions struct {
	Image    string   `json:"image"`
	Registry string   `json:"registry"`
	Versions []string `json:"versions"`
}

func (s *Service) UpgradeVersions(ctx context.Context, clusterID uint64) (UpgradeVersions, error) {
	var result UpgradeVersions
	if err := s.read(ctx, clusterID, "upgrade.versions", []string{"orch", "upgrade", "ls", "--format", "json"}, &result); err != nil {
		return result, err
	}
	valid := result.Image != "" && result.Registry != "" && result.Versions != nil
	version := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	for _, item := range result.Versions {
		valid = valid && version.MatchString(item)
	}
	if !valid {
		return UpgradeVersions{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned an invalid upgrade version list"}
	}
	return result, nil
}
