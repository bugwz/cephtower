package cluster

import (
	"context"
	"errors"
	"fmt"
	"strings"

	collectorservice "cephtower/backend/internal/service/collector"
	"cephtower/backend/internal/store"
)

var ErrNotFound = errors.New("cluster not found")

type Input struct {
	Name              string
	MonitorHost       string
	Keyring           string
	DashboardUsername string
	DashboardPassword string
}

type Detail struct {
	Cluster   store.CephCluster
	Discovery Discovery
}

type Dependencies struct {
	Database        func() *store.Database
	Discover        collectorservice.ClusterDiscoverer
	CleanRuntime    func(context.Context, uint) error
	EnsureDefaults  func(context.Context, *store.Database) error
	DeleteResources func(context.Context, *store.Database, uint) error
}

type Service struct {
	database        func() *store.Database
	discover        collectorservice.ClusterDiscoverer
	cleanRuntime    func(context.Context, uint) error
	ensureDefaults  func(context.Context, *store.Database) error
	deleteResources func(context.Context, *store.Database, uint) error
}

func New(deps Dependencies) *Service {
	ensureDefaults := deps.EnsureDefaults
	if ensureDefaults == nil {
		ensureDefaults = collectorservice.EnsureDefaultSystemSettings
	}
	deleteResources := deps.DeleteResources
	if deleteResources == nil {
		deleteResources = DeleteCephClusterResources
	}
	return &Service{
		database:        deps.Database,
		discover:        deps.Discover,
		cleanRuntime:    deps.CleanRuntime,
		ensureDefaults:  ensureDefaults,
		deleteResources: deleteResources,
	}
}

func (s *Service) List(ctx context.Context) ([]store.CephCluster, error) {
	return s.database().ListClusters(ctx)
}

func (s *Service) Get(ctx context.Context, id uint) (Detail, error) {
	cluster, err := s.find(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	discovery, err := s.loadDiscovery(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	return Detail{Cluster: cluster, Discovery: discovery}, nil
}

func (s *Service) Create(ctx context.Context, input Input) error {
	cluster, err := build(input, nil)
	if err != nil {
		return err
	}
	db := s.database()
	if err := db.Transaction(func(tx *store.Database) error {
		if err := tx.CreateCluster(ctx, &cluster); err != nil {
			return err
		}
		return s.ensureDefaults(ctx, tx)
	}); err != nil {
		return err
	}

	if s.discover != nil {
		if err := s.discover(ctx, s.database(), &cluster); err != nil {
			// Discovery performs external I/O, so it deliberately runs outside the
			// create transaction. Compensate on failure to preserve API behavior.
			_ = s.database().Transaction(func(tx *store.Database) error {
				if cleanupErr := s.deleteResources(ctx, tx, cluster.ID); cleanupErr != nil {
					return cleanupErr
				}
				return tx.DeleteCluster(ctx, &cluster)
			})
			if s.cleanRuntime != nil {
				_ = s.cleanRuntime(ctx, cluster.ID)
			}
			return err
		}
	}
	return nil
}

func (s *Service) Update(ctx context.Context, id uint, input Input) error {
	current, err := s.find(ctx, id)
	if err != nil {
		return err
	}
	cluster, err := build(input, &current)
	if err != nil {
		return err
	}
	cluster.ID = current.ID
	cluster.CreatedAt = current.CreatedAt

	if err := s.database().Transaction(func(tx *store.Database) error {
		if err := tx.SaveCluster(ctx, &cluster); err != nil {
			return err
		}
		return s.ensureDefaults(ctx, tx)
	}); err != nil {
		return err
	}
	if s.discover != nil {
		return s.discover(ctx, s.database(), &cluster)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	cluster, err := s.find(ctx, id)
	if err != nil {
		return err
	}
	if err := s.database().Transaction(func(tx *store.Database) error {
		if err := s.deleteResources(ctx, tx, cluster.ID); err != nil {
			return err
		}
		return tx.DeleteCluster(ctx, &cluster)
	}); err != nil {
		return err
	}
	if s.cleanRuntime != nil {
		return s.cleanRuntime(ctx, cluster.ID)
	}
	return nil
}

func (s *Service) find(ctx context.Context, id uint) (store.CephCluster, error) {
	cluster, err := s.database().FindCluster(ctx, id)
	if errors.Is(err, store.ErrRecordNotFound) {
		return store.CephCluster{}, ErrNotFound
	}
	return cluster, err
}

func build(input Input, current *store.CephCluster) (store.CephCluster, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return store.CephCluster{}, fmt.Errorf("name is required")
	}

	cluster := store.CephCluster{
		Name:              name,
		MonitorHost:       strings.TrimSpace(input.MonitorHost),
		DashboardUsername: strings.TrimSpace(input.DashboardUsername),
	}
	if current != nil {
		cluster.ID = current.ID
		cluster.CreatedAt = current.CreatedAt
		cluster.DashboardPassword = current.DashboardPassword
		cluster.Keyring = current.Keyring
		cluster.MonitorHost = current.MonitorHost
	}
	if strings.TrimSpace(input.MonitorHost) != "" {
		cluster.MonitorHost = strings.TrimSpace(input.MonitorHost)
	}
	if input.DashboardPassword != "" {
		cluster.DashboardPassword = input.DashboardPassword
	}
	if input.Keyring != "" {
		cluster.Keyring = input.Keyring
	}

	if cluster.DashboardUsername == "" {
		return store.CephCluster{}, fmt.Errorf("dashboard username is required")
	}
	if strings.TrimSpace(cluster.MonitorHost) == "" {
		return store.CephCluster{}, fmt.Errorf("monitor host is required")
	}
	if cluster.DashboardPassword == "" {
		return store.CephCluster{}, fmt.Errorf("dashboard password is required")
	}
	if strings.TrimSpace(cluster.Keyring) == "" {
		return store.CephCluster{}, fmt.Errorf("keyring is required")
	}
	return cluster, nil
}
