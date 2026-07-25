package cluster

import (
	"context"

	"cephtower/backend/internal/store"
)

type Discovery struct {
	Hosts         []store.CephClusterHost
	OSDs          []store.CephClusterOSD
	OSDFlags      []store.CephClusterOSDFlag
	Daemons       []store.CephClusterDaemon
	Services      []store.CephClusterService
	Mons          []store.CephClusterMon
	Mgrs          []store.CephClusterMgr
	MDSs          []store.CephClusterMDS
	MgrModules    []store.CephClusterMgrModule
	Configuration []store.CephClusterConfiguration
}

func (s *Service) loadDiscovery(ctx context.Context, clusterID uint) (Discovery, error) {
	result, err := s.database().LoadClusterDiscovery(ctx, clusterID)
	if err != nil {
		return Discovery{}, err
	}
	return Discovery{Hosts: result.Hosts, OSDs: result.OSDs, OSDFlags: result.OSDFlags, Daemons: result.Daemons, Services: result.Services, Mons: result.Mons, Mgrs: result.Mgrs, MDSs: result.MDSs, MgrModules: result.MgrModules, Configuration: result.Configuration}, nil
}
