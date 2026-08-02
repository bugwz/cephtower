package hostprofile

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

const (
	AuthMethodPassword   = "password"
	AuthMethodPrivateKey = "private_key"
)

type Service struct {
	database      func() *store.Database
	encryptionKey string
}

type SaveInput struct {
	ClusterID        uint64
	Hostname         string
	SSHAddress       string
	SSHPort          uint16
	SSHUser          string
	SSHAuthMethod    string
	SSHPassword      *string
	SSHPrivateKey    *string
	SSHKeyPassphrase *string
	Notes            *string
}

type View struct {
	Hostname         string    `json:"hostname"`
	SSHAddress       string    `json:"ssh_address"`
	SSHPort          uint16    `json:"ssh_port"`
	SSHUser          string    `json:"ssh_user"`
	SSHAuthMethod    string    `json:"ssh_auth_method"`
	SSHPasswordSet   bool      `json:"ssh_password_set"`
	SSHPrivateKeySet bool      `json:"ssh_private_key_set"`
	Notes            string    `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
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
	hostname := strings.TrimSpace(input.Hostname)
	address := strings.TrimSpace(input.SSHAddress)
	user := strings.TrimSpace(input.SSHUser)
	method := strings.TrimSpace(input.SSHAuthMethod)
	if input.ClusterID == 0 || hostname == "" || address == "" || user == "" {
		return View{}, fmt.Errorf("cluster_id, hostname, ssh_address and ssh_user are required")
	}
	if method == "" {
		method = AuthMethodPassword
	}
	if method != AuthMethodPassword && method != AuthMethodPrivateKey {
		return View{}, fmt.Errorf("ssh_auth_method is not supported")
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
		SSHAuthMethod:     method,
		Notes:             trimOptional(input.Notes),
		CreatedAt:         now,
		UpdatedAt:         now,
		SSHPasswordSecret: nil,
		DiscoveredData:    "{}",
		ResourceVersion:   1,
	}
	existing, err := s.database().FindCephHost(ctx, input.ClusterID, hostname)
	if err != nil && !errors.Is(err, store.ErrRecordNotFound) {
		return View{}, err
	}
	if err == nil {
		row.ID = existing.ID
		row.CreatedAt = existing.CreatedAt
		row.SSHPasswordSecret = existing.SSHPasswordSecret
		row.SSHPrivateKeySecret = existing.SSHPrivateKeySecret
		row.SSHKeyPassphraseSecret = existing.SSHKeyPassphraseSecret
	}
	if input.SSHPassword != nil {
		secret, err := s.encryptOptional(*input.SSHPassword)
		if err != nil {
			return View{}, err
		}
		row.SSHPasswordSecret = secret
	}
	if input.SSHPrivateKey != nil {
		secret, err := s.encryptOptional(*input.SSHPrivateKey)
		if err != nil {
			return View{}, err
		}
		row.SSHPrivateKeySecret = secret
	}
	if input.SSHKeyPassphrase != nil {
		secret, err := s.encryptOptional(*input.SSHKeyPassphrase)
		if err != nil {
			return View{}, err
		}
		row.SSHKeyPassphraseSecret = secret
	}
	if method == AuthMethodPassword {
		row.SSHPrivateKeySecret = nil
		row.SSHKeyPassphraseSecret = nil
	}
	if method == AuthMethodPrivateKey {
		row.SSHPasswordSecret = nil
	}
	if err := s.database().UpsertCephHost(ctx, &row); err != nil {
		return View{}, err
	}
	saved, err := s.database().FindCephHost(ctx, input.ClusterID, hostname)
	if err != nil {
		return View{}, err
	}
	return toView(saved), nil
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
	notes := ""
	if row.Notes != nil {
		notes = *row.Notes
	}
	return View{
		Hostname:         row.Hostname,
		SSHAddress:       row.SSHAddress,
		SSHPort:          row.SSHPort,
		SSHUser:          row.SSHUser,
		SSHAuthMethod:    row.SSHAuthMethod,
		SSHPasswordSet:   row.SSHPasswordSecret != nil && *row.SSHPasswordSecret != "",
		SSHPrivateKeySet: row.SSHPrivateKeySecret != nil && *row.SSHPrivateKeySecret != "",
		Notes:            notes,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
