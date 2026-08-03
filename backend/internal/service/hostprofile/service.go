package hostprofile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

type Service struct {
	database      func() *store.Database
	encryptionKey string
}

type SaveInput struct {
	ClusterID     uint64
	Hostname      string
	SSHAddress    string
	SSHPort       uint16
	SSHUser       string
	SSHPassword   *string
	SyncHostnames []string
}

type View struct {
	Hostname        string    `json:"hostname"`
	SSHAddress      string    `json:"ssh_address"`
	SSHPort         uint16    `json:"ssh_port"`
	SSHUser         string    `json:"ssh_user"`
	SSHPasswordSet  bool      `json:"ssh_password_set"`
	SyncedHostnames []string  `json:"synced_hostnames,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func New(database func() *store.Database, encryptionKey string) *Service {
	return &Service{database: database, encryptionKey: encryptionKey}
}

func (s *Service) List(ctx context.Context, clusterID uint64) ([]View, error) {
	rows, err := s.database().ListCephHosts(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	views := make([]View, 0, len(rows))
	for _, row := range rows {
		views = append(views, toView(row))
	}
	return views, nil
}

func (s *Service) Get(ctx context.Context, clusterID uint64, hostname string) (View, error) {
	row, err := s.database().FindCephHost(ctx, clusterID, strings.TrimSpace(hostname))
	if err != nil {
		return View{}, err
	}
	return toView(row), nil
}

func (s *Service) Save(ctx context.Context, input SaveInput) (View, error) {
	var saved View
	var syncedHostnames []string
	err := s.database().Transaction(func(tx *store.Database) error {
		var err error
		saved, err = s.saveOne(ctx, tx, input)
		if err != nil {
			return err
		}
		syncedHostnames, err = s.syncHosts(ctx, tx, input)
		return err
	})
	if err != nil {
		return View{}, err
	}
	saved.SyncedHostnames = syncedHostnames
	return saved, nil
}

func (s *Service) saveOne(ctx context.Context, database *store.Database, input SaveInput) (View, error) {
	hostname := strings.TrimSpace(input.Hostname)
	address := strings.TrimSpace(input.SSHAddress)
	user := strings.TrimSpace(input.SSHUser)
	if input.ClusterID == 0 || hostname == "" || address == "" || user == "" {
		return View{}, fmt.Errorf("cluster_id, hostname, ssh_address and ssh_user are required")
	}
	port := input.SSHPort
	if port == 0 {
		port = 22
	}

	now := time.Now().UTC()
	row := store.CephHost{
		ClusterID:         input.ClusterID,
		Hostname:          hostname,
		SSHAddress:        address,
		SSHPort:           port,
		SSHUser:           user,
		CreatedAt:         now,
		UpdatedAt:         now,
		SSHPasswordSecret: nil,
		DiscoveredData:    "{}",
		ResourceVersion:   1,
	}
	existing, err := database.FindCephHost(ctx, input.ClusterID, hostname)
	if err != nil && !errors.Is(err, store.ErrRecordNotFound) {
		return View{}, err
	}
	if err == nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
		row.SSHPasswordSecret = existing.SSHPasswordSecret
	}
	if input.SSHPassword != nil {
		secret, err := s.encryptOptional(*input.SSHPassword)
		if err != nil {
			return View{}, err
		}
		row.SSHPasswordSecret = secret
	}
	if err := database.UpsertCephHost(ctx, &row); err != nil {
		return View{}, err
	}
	saved, err := database.FindCephHost(ctx, input.ClusterID, hostname)
	if err != nil {
		return View{}, err
	}
	return toView(saved), nil
}

func (s *Service) syncHosts(ctx context.Context, database *store.Database, input SaveInput) ([]string, error) {
	hostnames := cleanSyncHostnames(input.Hostname, input.SyncHostnames)
	if len(hostnames) == 0 {
		return nil, nil
	}
	synced := make([]string, 0, len(hostnames))
	for _, hostname := range hostnames {
		target, err := database.FindCephHost(ctx, input.ClusterID, hostname)
		if err != nil {
			if errors.Is(err, store.ErrRecordNotFound) {
				return nil, fmt.Errorf("sync host %q was not found", hostname)
			}
			return nil, err
		}
		address := hostAddress(target)
		if address == "" {
			return nil, fmt.Errorf("sync host %q has no address", hostname)
		}
		_, err = s.saveOne(ctx, database, SaveInput{
			ClusterID:   input.ClusterID,
			Hostname:    hostname,
			SSHAddress:  address,
			SSHPort:     input.SSHPort,
			SSHUser:     input.SSHUser,
			SSHPassword: input.SSHPassword,
		})
		if err != nil {
			return nil, err
		}
		synced = append(synced, hostname)
	}
	return synced, nil
}

func (s *Service) encryptOptional(value string) (*string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	encrypted, err := security.Encrypt([]byte(trimmed), s.encryptionKey)
	if err != nil {
		return nil, err
	}
	return &encrypted, nil
}

func toView(row store.CephHost) View {
	return View{
		Hostname:       row.Hostname,
		SSHAddress:     row.SSHAddress,
		SSHPort:        row.SSHPort,
		SSHUser:        row.SSHUser,
		SSHPasswordSet: row.SSHPasswordSecret != nil && *row.SSHPasswordSecret != "",
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func cleanSyncHostnames(current string, values []string) []string {
	current = strings.TrimSpace(current)
	seen := map[string]struct{}{}
	var hostnames []string
	for _, value := range values {
		hostname := strings.TrimSpace(value)
		if hostname == "" || hostname == current {
			continue
		}
		if _, exists := seen[hostname]; exists {
			continue
		}
		seen[hostname] = struct{}{}
		hostnames = append(hostnames, hostname)
	}
	return hostnames
}

func hostAddress(row store.CephHost) string {
	if address := strings.TrimSpace(row.SSHAddress); address != "" {
		return address
	}
	if row.Address != nil {
		if address := strings.TrimSpace(*row.Address); address != "" {
			return address
		}
	}
	var discovered map[string]any
	if err := json.Unmarshal([]byte(row.DiscoveredData), &discovered); err != nil {
		return ""
	}
	for _, field := range []string{"address", "addr", "ip", "public_addr"} {
		if address, ok := discovered[field].(string); ok {
			if trimmed := strings.TrimSpace(address); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}
