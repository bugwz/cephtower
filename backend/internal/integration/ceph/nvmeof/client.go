package nvmeof

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DialConfig struct {
	Address, ServerName, BearerToken            string
	CACertificate, ClientCertificate, ClientKey []byte
}
type Client struct {
	gateway    GatewayClient
	connection *grpc.ClientConn
}

func Dial(_ context.Context, config DialConfig) (*Client, error) {
	if strings.TrimSpace(config.Address) == "" {
		return nil, fmt.Errorf("NVMe-oF gateway address is required")
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: strings.TrimSpace(config.ServerName)}
	if len(config.CACertificate) > 0 {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(config.CACertificate) {
			return nil, fmt.Errorf("NVMe-oF CA certificate is invalid")
		}
		tlsConfig.RootCAs = pool
	}
	if len(config.ClientCertificate) > 0 || len(config.ClientKey) > 0 {
		certificate, err := tls.X509KeyPair(config.ClientCertificate, config.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("load NVMe-oF client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}
	options := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))}
	if token := strings.TrimSpace(config.BearerToken); token != "" {
		options = append(options, grpc.WithPerRPCCredentials(bearerCredential(token)))
	}
	connection, err := grpc.NewClient(config.Address, options...)
	if err != nil {
		return nil, fmt.Errorf("create NVMe-oF gRPC connection: %w", err)
	}
	return &Client{gateway: NewGatewayClient(connection), connection: connection}, nil
}
func NewClient(gateway GatewayClient) (*Client, error) {
	if gateway == nil {
		return nil, fmt.Errorf("NVMe-oF gateway client is required")
	}
	return &Client{gateway: gateway}, nil
}
func (c *Client) Close() error {
	if c == nil || c.connection == nil {
		return nil
	}
	return c.connection.Close()
}
func (c *Client) Gateway() GatewayClient { return c.gateway }
func (c *Client) ListSubsystems(ctx context.Context, nqn *string) (*SubsystemsInfoCli, error) {
	return c.gateway.ListSubsystems(ctx, &ListSubsystemsReq{SubsystemNqn: nqn})
}
func (c *Client) CreateSubsystem(ctx context.Context, request *CreateSubsystemReq) (*SubsysStatus, error) {
	if request == nil || strings.TrimSpace(request.SubsystemNqn) == "" {
		return nil, fmt.Errorf("subsystem NQN is required")
	}
	return c.gateway.CreateSubsystem(ctx, request)
}
func (c *Client) DeleteSubsystem(ctx context.Context, request *DeleteSubsystemReq) (*ReqStatus, error) {
	if request == nil || strings.TrimSpace(request.SubsystemNqn) == "" {
		return nil, fmt.Errorf("subsystem NQN is required")
	}
	return c.gateway.DeleteSubsystem(ctx, request)
}
func (c *Client) ListNamespaces(ctx context.Context, request *ListNamespacesReq) (*NamespacesInfo, error) {
	if request == nil {
		request = &ListNamespacesReq{}
	}
	return c.gateway.ListNamespaces(ctx, request)
}
func (c *Client) AddNamespace(ctx context.Context, request *NamespaceAddReq) (*NsidStatus, error) {
	if request == nil || request.SubsystemNqn == "" || request.RbdPoolName == "" || request.RbdImageName == "" {
		return nil, fmt.Errorf("subsystem_nqn, rbd_pool_name and rbd_image_name are required")
	}
	return c.gateway.NamespaceAdd(ctx, request)
}
func (c *Client) DeleteNamespace(ctx context.Context, request *NamespaceDeleteReq) (*ReqStatus, error) {
	if request == nil {
		return nil, fmt.Errorf("namespace request is required")
	}
	return c.gateway.NamespaceDelete(ctx, request)
}
func (c *Client) ListListeners(ctx context.Context, request *ListListenersReq) (*ListenersInfo, error) {
	if request == nil {
		request = &ListListenersReq{}
	}
	return c.gateway.ListListeners(ctx, request)
}
func (c *Client) CreateListener(ctx context.Context, request *CreateListenerReq) (*ReqStatus, error) {
	if request == nil {
		return nil, fmt.Errorf("listener request is required")
	}
	return c.gateway.CreateListener(ctx, request)
}
func (c *Client) DeleteListener(ctx context.Context, request *DeleteListenerReq) (*ReqStatus, error) {
	if request == nil {
		return nil, fmt.Errorf("listener request is required")
	}
	return c.gateway.DeleteListener(ctx, request)
}
func (c *Client) ListHosts(ctx context.Context, request *ListHostsReq) (*HostsInfo, error) {
	if request == nil {
		request = &ListHostsReq{}
	}
	return c.gateway.ListHosts(ctx, request)
}
func (c *Client) AddHost(ctx context.Context, request *AddHostReq) (*ReqStatus, error) {
	if request == nil {
		return nil, fmt.Errorf("host request is required")
	}
	return c.gateway.AddHost(ctx, request)
}
func (c *Client) RemoveHost(ctx context.Context, request *RemoveHostReq) (*ReqStatus, error) {
	if request == nil {
		return nil, fmt.Errorf("host request is required")
	}
	return c.gateway.RemoveHost(ctx, request)
}
func (c *Client) ListConnections(ctx context.Context, request *ListConnectionsReq) (*ConnectionsInfo, error) {
	if request == nil {
		request = &ListConnectionsReq{}
	}
	return c.gateway.ListConnections(ctx, request)
}

type bearerCredential string

func (b bearerCredential) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + string(b)}, nil
}
func (bearerCredential) RequireTransportSecurity() bool { return true }
