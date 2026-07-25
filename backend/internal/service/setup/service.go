package setup

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cephtower/backend/internal/config"
	authservice "cephtower/backend/internal/service/auth"
	"cephtower/backend/internal/service/collector"
	"cephtower/backend/internal/store"
)

var (
	ErrAlreadyInitialized = errors.New("system has already been initialized")
	ErrTargetInitialized  = errors.New("selected database has already been initialized")
)

type Input struct {
	Database config.DatabaseConfig
	Username string
	Email    string
	Password string
}

type Service struct {
	database *store.Manager
	current  func() config.Config
	updated  func(config.Config)
}

func New(database *store.Manager, current func() config.Config, updated func(config.Config)) *Service {
	return &Service{database: database, current: current, updated: updated}
}

func (s *Service) Status(ctx context.Context) (bool, config.DatabaseConfig, error) {
	initialized, err := hasUsers(ctx, s.database.Current())
	return initialized, s.current().Database, err
}

func (s *Service) TestDatabase(ctx context.Context, database config.DatabaseConfig) error {
	initialized, err := hasUsers(ctx, s.database.Current())
	if err != nil {
		return err
	}
	if initialized {
		return ErrAlreadyInitialized
	}

	database, err = config.NormalizeDatabaseConfig(database)
	if err != nil {
		return err
	}
	return store.TestConnection(ctx, database, s.current().Server.Dir)
}

func (s *Service) Initialize(ctx context.Context, input Input) error {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(input.Email)
	if input.Username == "" || input.Email == "" || input.Password == "" {
		return fmt.Errorf("admin username, email and password are required")
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	initialized, err := hasUsers(ctx, s.database.Current())
	if err != nil {
		return err
	}
	if initialized {
		return ErrAlreadyInitialized
	}
	databaseCfg, err := config.NormalizeDatabaseConfig(input.Database)
	if err != nil {
		return err
	}
	if databaseCfg.Engine == store.EngineMySQL && strings.TrimSpace(databaseCfg.MySQL.Password) == "" {
		return fmt.Errorf("mysql password is required")
	}
	current := s.current()
	newDB, err := store.Open(databaseCfg, current.Server.Dir)
	if err != nil {
		return err
	}
	keep := false
	defer func() {
		if !keep {
			_ = store.Close(newDB)
		}
	}()
	targetInitialized, err := hasUsers(ctx, newDB)
	if err != nil {
		return err
	}
	if targetInitialized {
		return ErrTargetInitialized
	}
	passwordHash, err := store.HashPassword(input.Password)
	if err != nil {
		return err
	}
	admin := store.User{Username: input.Username, DisplayName: input.Username, Email: input.Email, Role: store.UserRoleAdmin, Permissions: authservice.PermissionsJSON(nil, store.UserRoleAdmin), PasswordHash: passwordHash, Enabled: true}
	if err := newDB.Transaction(func(tx *store.Database) error {
		if err := tx.CreateUser(ctx, &admin); err != nil {
			return err
		}
		if err := collector.EnsureDefaultSystemSettings(ctx, tx); err != nil {
			return err
		}
		return config.SaveDatabase(current.Path, databaseCfg)
	}); err != nil {
		return err
	}
	current.Database = databaseCfg
	previous := s.database.Replace(newDB)
	if s.updated != nil {
		s.updated(current)
	}
	keep = true
	if previous != nil && previous != newDB {
		_ = store.Close(previous)
	}
	return nil
}

func hasUsers(ctx context.Context, db *store.Database) (bool, error) {
	return db.HasUsers(ctx)
}
