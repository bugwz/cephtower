package setup

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cephtower/backend/internal/config"
	authservice "cephtower/backend/internal/service/auth"
	"cephtower/backend/internal/store"
)

type Status struct {
	Required bool
}

type Service struct {
	Manager       *store.Manager
	CurrentConfig func() config.Config
	UpdateConfig  func(config.Config)
	OnInitialized func()
}

func (s *Service) BootstrapRequired(_ context.Context) (bool, error) {
	return s.currentConfig().Server.Bootstrap, nil
}

func (s *Service) Status(ctx context.Context) (Status, error) {
	required, err := s.BootstrapRequired(ctx)
	if err != nil {
		return Status{}, err
	}
	return Status{Required: required}, nil
}

func (s *Service) Initialize(ctx context.Context, database config.DatabaseConfig, admin authservice.CreateUserInput) (store.User, error) {
	if s == nil || s.Manager == nil {
		return store.User{}, fmt.Errorf("setup service is unavailable")
	}
	current := s.currentConfig()
	if strings.TrimSpace(database.EncryptionKey) == "" {
		database.EncryptionKey = current.Database.EncryptionKey
	}
	if strings.TrimSpace(database.EncryptionKey) == "" {
		key, err := config.GenerateDatabaseEncryptionKey()
		if err != nil {
			return store.User{}, err
		}
		database.EncryptionKey = key
	}
	normalized, err := config.NormalizeDatabaseConfig(database)
	if err != nil {
		return store.User{}, err
	}
	nextConfig := current
	nextConfig.Database = normalized
	nextConfig.Server.Bootstrap = false

	if err := store.TestInitializationTarget(ctx, normalized, current.Server.Dir); err != nil {
		return store.User{}, err
	}
	db, err := store.Initialize(ctx, normalized, current.Server.Dir)
	if err != nil {
		return store.User{}, err
	}
	installed := false
	defer func() {
		if !installed {
			_ = store.Close(db)
		}
	}()

	auth := authservice.New(func() *store.Database { return db }, func() config.Config { return nextConfig })
	if err := auth.EnsureRoles(ctx); err != nil {
		return store.User{}, fmt.Errorf("initialize roles: %w", err)
	}
	user, err := auth.CreateInitialAdmin(ctx, admin)
	if err != nil {
		return store.User{}, err
	}
	if err := config.SaveSetup(current.Path, normalized, false); err != nil {
		return store.User{}, err
	}

	previous := s.Manager.Replace(db)
	installed = true
	if previous != nil && previous != db {
		_ = store.Close(previous)
	}
	if s.UpdateConfig != nil {
		s.UpdateConfig(nextConfig)
	}
	if s.OnInitialized != nil {
		s.OnInitialized()
	}
	return user, nil
}

func (s *Service) TestDatabase(ctx context.Context, database config.DatabaseConfig) error {
	current := s.currentConfig()
	if strings.TrimSpace(database.EncryptionKey) == "" {
		database.EncryptionKey = current.Database.EncryptionKey
	}
	normalized, err := config.NormalizeDatabaseConfig(database)
	if err != nil {
		return err
	}
	return store.TestInitializationTarget(ctx, normalized, current.Server.Dir)
}

func (s *Service) currentConfig() config.Config {
	if s != nil && s.CurrentConfig != nil {
		return s.CurrentConfig()
	}
	return config.Config{}
}

var ErrAlreadyInitialized = authservice.ErrAlreadyInitialized

func IsAlreadyInitialized(err error) bool {
	return errors.Is(err, ErrAlreadyInitialized)
}
