package nvmeof

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type fakeGateway struct{ UnimplementedGatewayServer }

func (fakeGateway) ListSubsystems(context.Context, *ListSubsystemsReq) (*SubsystemsInfoCli, error) {
	return &SubsystemsInfoCli{Subsystems: []*SubsystemCli{{Nqn: "nqn.2026-07.example:test"}}}, nil
}

func TestGeneratedGatewayClientAgainstInMemoryGRPCServer(t *testing.T) {
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	RegisterGatewayServer(server, fakeGateway{})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	connection, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	client, err := NewClient(NewGatewayClient(connection))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ListSubsystems(context.Background(), nil)
	if err != nil || len(result.Subsystems) != 1 || result.Subsystems[0].Nqn == "" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestDialRejectsInvalidCA(t *testing.T) {
	if _, err := Dial(context.Background(), DialConfig{Address: "gateway.example.test:5500", CACertificate: []byte("not a certificate")}); err == nil {
		t.Fatal("invalid CA accepted")
	}
}
