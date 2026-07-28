package lab

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"cephtower/tools/aliyun-ceph-lab/internal/cloud"
	"cephtower/tools/aliyun-ceph-lab/internal/config"
	"cephtower/tools/aliyun-ceph-lab/internal/logging"
	"cephtower/tools/aliyun-ceph-lab/internal/remote"
	"cephtower/tools/aliyun-ceph-lab/internal/state"
)

type Manager struct {
	Config    *config.Config
	Cloud     *cloud.Client
	StatePath string
	SSH       *remote.SSH
}

func New(cfg *config.Config, statePath string) (*Manager, error) {
	client, err := cloud.New(cfg)
	if err != nil {
		return nil, err
	}
	if statePath == "" {
		statePath = cfg.DefaultStatePath()
	} else if abs, err := filepath.Abs(statePath); err == nil {
		statePath = abs
	}
	knownHostsPath := filepath.Join(filepath.Dir(statePath), cfg.ClusterName+".known_hosts")
	logDir := filepath.Join(filepath.Dir(statePath), "log")
	return &Manager{
		Config: cfg, Cloud: client, StatePath: statePath,
		SSH: &remote.SSH{
			User: cfg.SSHUser, Password: cfg.SSHPassword,
			KnownHostsPath: knownHostsPath, LogDir: logDir,
		},
	}, nil
}

func (m *Manager) Create(ctx context.Context) error {
	if _, err := os.Stat(m.StatePath); err == nil {
		return fmt.Errorf("state already exists at %s; delete the existing lab first", m.StatePath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(m.SSH.KnownHostsPath), 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	if err := os.MkdirAll(m.SSH.LogDir, 0o700); err != nil {
		return fmt.Errorf("create per-node log directory: %w", err)
	}
	if err := os.Chmod(m.SSH.LogDir, 0o700); err != nil {
		return fmt.Errorf("secure per-node log directory: %w", err)
	}
	initScript := m.Config.ResolvePath(m.Config.InitScript)
	deployScript := m.Config.ResolvePath(m.Config.DeployScript)
	for _, path := range []string{initScript, deployScript} {
		if info, err := os.Stat(path); err != nil {
			return fmt.Errorf("inspect hook %s: %w", path, err)
		} else if !info.Mode().IsRegular() {
			return fmt.Errorf("hook %s is not a regular file", path)
		}
	}

	now := time.Now().UTC()
	expiresAt := now.Add(m.Config.MaxRuntimeDuration()).Add(59 * time.Second).Truncate(time.Minute)
	current := &state.State{
		Version: state.Version, ClusterName: m.Config.ClusterName, RegionID: m.Config.RegionID,
		CreatedAt: now, ExpiresAt: expiresAt,
	}
	if err := state.Save(m.StatePath, current); err != nil {
		return err
	}
	networkCtx, cancel := context.WithTimeout(ctx, m.Config.WaitTimeoutDuration())
	network, err := m.Cloud.EnsureNetwork(networkCtx, m.Config, current.CreatedAt, func(value state.Network) error {
		current.Network = value
		return state.Save(m.StatePath, current)
	})
	cancel()
	if err != nil {
		if !stateHasCloudResources(current) {
			if cleanupErr := cleanupLocalState(m.StatePath, m.SSH.KnownHostsPath); cleanupErr != nil {
				return fmt.Errorf("prepare network: %w; cleanup empty state: %v", err, cleanupErr)
			}
			return fmt.Errorf("prepare network: %w", err)
		}
		return fmt.Errorf("prepare network: %w; created resources are recorded in %s", err, m.StatePath)
	}
	current.Network = network
	m.Config.VPCID = network.VPCID
	m.Config.VSwitchID = network.VSwitchID
	m.Config.SecurityGroupID = network.SecurityGroupID
	logging.Infof("create: network ready; vpc=%s vSwitch=%s securityGroup=%s", network.VPCID, network.VSwitchID, network.SecurityGroupID)

	for index, node := range m.Config.Nodes {
		logging.Infof("create: creating node %d/%d %s (%s); automatic release=%s", index+1, len(m.Config.Nodes), node.Name, node.InstanceType, expiresAt.Format(time.RFC3339))
		instanceID, err := m.Cloud.RunNode(ctx, m.Config, node, expiresAt)
		if err != nil {
			return fmt.Errorf("%w; created instance IDs remain protected by auto-release and are recorded in %s", err, m.StatePath)
		}
		current.Nodes = append(current.Nodes, state.Node{
			Name: node.Name, InstanceID: instanceID, Status: "Creating",
			SSH: m.nodeSSHConnection(node.Name, ""),
		})
		logging.Infof("create: node %s request accepted; instance=%s", node.Name, instanceID)
		if err := state.Save(m.StatePath, current); err != nil {
			return fmt.Errorf("save instance %s to state: %w", instanceID, err)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, m.Config.WaitTimeoutDuration())
	logging.Infof("create: waiting for %d instance(s) to become Running with public and private IPs", len(current.Nodes))
	if err := m.waitUntilRunning(waitCtx, current); err != nil {
		cancel()
		return err
	}
	cancel()
	if err := state.Save(m.StatePath, current); err != nil {
		return err
	}

	logging.Infof("create: starting SSH initialization on %d node(s)", len(current.Nodes))
	privateKey, publicKey, err := generateClusterSSHKey(m.Config.ClusterName)
	if err != nil {
		return fmt.Errorf("generate cluster SSH key: %w", err)
	}
	nodeConnections, err := m.initializeNodes(ctx, current, initScript, privateKey, publicKey)
	if err != nil {
		return fmt.Errorf("initialize nodes: %w; run delete --yes or wait for automatic release", err)
	}
	defer closeNodeConnections(nodeConnections)
	logging.Infof("create: SSH initialization completed on all %d node(s)", len(current.Nodes))
	if len(current.Nodes) < 3 {
		logging.Infof("create: skipping Ceph deployment because %d node(s) were configured; at least 3 are required", len(current.Nodes))
		return nil
	}
	first := current.Nodes[0]
	args := []string{
		"--cluster-name", m.Config.ClusterName,
		"--bootstrap-node-name", first.Name,
		"--node-names", joinNodeField(current.Nodes, func(node state.Node) string { return node.Name }),
		"--public-ips", joinNodeField(current.Nodes, func(node state.Node) string { return node.PublicIP }),
		"--private-ips", joinNodeField(current.Nodes, func(node state.Node) string { return node.PrivateIP }),
		"--data-disk-counts", joinDataDiskCounts(m.Config.Nodes),
		"--wait-timeout-seconds", fmt.Sprintf("%d", int(m.Config.WaitTimeoutDuration().Seconds())),
	}
	logging.Infof("create: running deployment hook on first node %s (%s)", first.Name, first.PublicIP)
	deploymentStages := time.Duration(2*len(current.Nodes) + 6)
	deployCtx, deployCancel := context.WithTimeout(ctx, deploymentStages*m.Config.WaitTimeoutDuration())
	defer deployCancel()
	deployConnection, ok := nodeConnections[first.Name]
	if !ok {
		return fmt.Errorf("missing established SSH connection for bootstrap node %s", first.Name)
	}
	if err := deployConnection.RunScript(deployCtx, deployScript, args); err != nil {
		return fmt.Errorf("run ceph deployment hook: %w", err)
	}
	logging.Infof("create: deployment hook completed on %s", first.Name)
	return nil
}
func (m *Manager) List() (*state.State, error) {
	current, err := state.Load(m.StatePath)
	if err != nil {
		return nil, err
	}
	if err := m.refresh(context.Background(), current); err != nil {
		return nil, err
	}
	if err := state.Save(m.StatePath, current); err != nil {
		return nil, err
	}
	return current, nil
}

func (m *Manager) Delete(ctx context.Context) error {
	current, err := state.Load(m.StatePath)
	if err != nil {
		return err
	}
	instances, err := m.Cloud.Describe(ctx, current.RegionID, instanceIDs(current.Nodes))
	if err != nil {
		return err
	}
	existing := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		existing[instance.ID] = struct{}{}
	}
	var failures []string
	for _, node := range current.Nodes {
		if _, exists := existing[node.InstanceID]; !exists {
			logging.Infof("delete: node already released; name=%s instance=%s", node.Name, node.InstanceID)
			continue
		}
		logging.Infof("delete: releasing node %s (%s)", node.Name, node.InstanceID)
		if err := m.Cloud.Delete(ctx, node.InstanceID); err != nil {
			failures = append(failures, err.Error())
		} else {
			logging.Infof("delete: release request accepted for %s (%s)", node.Name, node.InstanceID)
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("some instances could not be released: %s", strings.Join(failures, "; "))
	}
	cleanupCtx, cancel := context.WithTimeout(ctx, m.Config.WaitTimeoutDuration())
	defer cancel()
	if len(current.Nodes) > 0 {
		logging.Infof("delete: waiting for ECS instances to disappear before network cleanup")
		if err := m.waitInstancesReleased(cleanupCtx, current); err != nil {
			return err
		}
	}
	if m.Config.CleanupCreatedNetworkResources() {
		logging.Infof("delete: cleanup_created_resources=true; deleting owned network resources")
		if err := m.Cloud.DeleteNetwork(cleanupCtx, current.RegionID, current.Network, func(value state.Network) error {
			current.Network = value
			return state.Save(m.StatePath, current)
		}); err != nil {
			return err
		}
	} else {
		logging.Infof("delete: cleanup_created_resources=false; retaining network resources for reuse")
	}
	if err := cleanupLocalState(m.StatePath, m.SSH.KnownHostsPath); err != nil {
		return err
	}
	return nil
}

func cleanupLocalState(statePath, knownHostsPath string) error {
	stateDir := filepath.Clean(filepath.Dir(statePath))
	if filepath.Base(stateDir) == ".state" {
		if err := os.RemoveAll(stateDir); err != nil {
			return fmt.Errorf("remove state directory: %w", err)
		}
		return nil
	}

	if err := os.Remove(statePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove state: %w", err)
	}
	if err := os.Remove(knownHostsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove known_hosts: %w", err)
	}
	return nil
}

func stateHasCloudResources(current *state.State) bool {
	if current == nil {
		return false
	}
	if len(current.Nodes) > 0 {
		return true
	}
	return current.Network.CreatedVPC ||
		current.Network.CreatedVSwitch ||
		current.Network.CreatedSecurityGroup
}

func (m *Manager) waitInstancesReleased(ctx context.Context, current *state.State) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	emptyChecks := 0
	lastRemaining := -1
	for {
		instances, err := m.Cloud.Describe(ctx, current.RegionID, instanceIDs(current.Nodes))
		if err != nil {
			return err
		}
		if len(instances) == 0 {
			emptyChecks++
			if emptyChecks >= 2 {
				logging.Infof("delete: all recorded instances are released")
				return nil
			}
		} else {
			emptyChecks = 0
			if len(instances) != lastRemaining {
				logging.Infof("delete: waiting for %d instance(s) to be released", len(instances))
				lastRemaining = len(instances)
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for ECS instances to be released before network cleanup: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func (m *Manager) waitUntilRunning(ctx context.Context, current *state.State) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	lastSummary := ""
	for {
		if err := m.refresh(ctx, current); err != nil {
			return err
		}
		allRunning := true
		for _, node := range current.Nodes {
			if node.Status != "Running" || node.PublicIP == "" || node.PrivateIP == "" {
				allRunning = false
				break
			}
		}
		if allRunning {
			logging.Infof("create: all %d instance(s) are Running and have required IP addresses", len(current.Nodes))
			return nil
		}
		summary := nodeStatusSummary(current.Nodes)
		if summary != lastSummary {
			logging.Infof("create: waiting for instances: %s", summary)
			lastSummary = summary
		}
		if err := state.Save(m.StatePath, current); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for instances to become Running: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func (m *Manager) refresh(ctx context.Context, current *state.State) error {
	instances, err := m.Cloud.Describe(ctx, current.RegionID, instanceIDs(current.Nodes))
	if err != nil {
		return err
	}
	byID := make(map[string]cloud.Instance, len(instances))
	for _, instance := range instances {
		byID[instance.ID] = instance
	}
	for i := range current.Nodes {
		if instance, exists := byID[current.Nodes[i].InstanceID]; exists {
			current.Nodes[i].Status = instance.Status
			current.Nodes[i].PublicIP = instance.PublicIP
			current.Nodes[i].PrivateIP = instance.PrivateIP
			if current.Nodes[i].SSH.User == "" {
				current.Nodes[i].SSH = m.nodeSSHConnection(current.Nodes[i].Name, instance.PublicIP)
			} else {
				current.Nodes[i].SSH.Host = instance.PublicIP
				current.Nodes[i].SSH.Port = 22
				current.Nodes[i].SSH.LogPath = m.SSH.HostLogPath(current.Nodes[i].Name)
			}
		} else {
			current.Nodes[i].Status = "ReleasedOrNotFound"
		}
	}
	return nil
}

func (m *Manager) nodeSSHConnection(hostname, host string) state.SSH {
	connection := state.SSH{
		Host: host, Port: 22, User: m.Config.SSHUser, Password: m.Config.SSHPassword,
		PasswordGenerated: m.Config.SSHPasswordWasGenerated(), LogPath: m.SSH.HostLogPath(hostname),
	}
	return connection
}

func (m *Manager) initializeNodes(
	ctx context.Context,
	current *state.State,
	scriptPath string,
	privateKey string,
	publicKey string,
) (map[string]*remote.HostConnection, error) {
	connections := make(map[string]*remote.HostConnection, len(current.Nodes))
	sshCtx, cancel := context.WithTimeout(ctx, m.Config.WaitTimeoutDuration())
	for _, node := range current.Nodes {
		connection, err := m.SSH.Connect(sshCtx, node.PublicIP, node.Name)
		if err != nil {
			cancel()
			closeNodeConnections(connections)
			return nil, err
		}
		connections[node.Name] = connection
	}
	cancel()

	scriptCtx, cancel := context.WithTimeout(ctx, 2*m.Config.WaitTimeoutDuration())
	defer cancel()
	var wg sync.WaitGroup
	errorsByNode := make(chan error, len(current.Nodes))
	for index, node := range current.Nodes {
		index := index
		node := node
		wg.Add(1)
		go func() {
			defer wg.Done()
			connection, ok := connections[node.Name]
			if !ok {
				errorsByNode <- fmt.Errorf("missing established SSH connection for node %s", node.Name)
				return
			}
			if err := connection.RunScript(scriptCtx, scriptPath, []string{
				"--cluster-name", m.Config.ClusterName,
				"--node-name", node.Name,
				"--node-names", joinNodeField(current.Nodes, func(item state.Node) string { return item.Name }),
				"--public-ips", joinNodeField(current.Nodes, func(item state.Node) string { return item.PublicIP }),
				"--private-ips", joinNodeField(current.Nodes, func(item state.Node) string { return item.PrivateIP }),
				"--data-disk-count", fmt.Sprintf("%d", len(m.Config.Nodes[index].DataDisks)),
				"--ssh-private-key-base64", privateKey,
				"--ssh-public-key-base64", publicKey,
			}); err != nil {
				errorsByNode <- err
				return
			}
		}()
	}
	wg.Wait()
	close(errorsByNode)
	var messages []string
	for err := range errorsByNode {
		messages = append(messages, err.Error())
	}
	if len(messages) > 0 {
		sort.Strings(messages)
		closeNodeConnections(connections)
		return nil, errors.New(strings.Join(messages, "; "))
	}
	return connections, nil
}

func closeNodeConnections(connections map[string]*remote.HostConnection) {
	names := make([]string, 0, len(connections))
	for name := range connections {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := connections[name].Close(); err != nil {
			logging.Warnf("ssh: close connection for %s: %v", name, err)
		}
	}
}

func nodeStatusSummary(nodes []state.Node) string {
	values := make([]string, 0, len(nodes))
	for _, node := range nodes {
		values = append(values, fmt.Sprintf("%s=%s(public=%t,private=%t)",
			node.Name, node.Status, node.PublicIP != "", node.PrivateIP != ""))
	}
	return strings.Join(values, ", ")
}

func instanceIDs(nodes []state.Node) []string {
	result := make([]string, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, node.InstanceID)
	}
	return result
}

func joinNodeField(nodes []state.Node, field func(state.Node) string) string {
	values := make([]string, 0, len(nodes))
	for _, node := range nodes {
		values = append(values, field(node))
	}
	return strings.Join(values, ",")
}

func joinDataDiskCounts(nodes []config.Node) string {
	values := make([]string, 0, len(nodes))
	for _, node := range nodes {
		values = append(values, fmt.Sprintf("%d", len(node.DataDisks)))
	}
	return strings.Join(values, ",")
}

func generateClusterSSHKey(clusterName string) (string, string, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	block, err := ssh.MarshalPrivateKey(private, clusterName)
	if err != nil {
		return "", "", err
	}
	sshPublic, err := ssh.NewPublicKey(public)
	if err != nil {
		return "", "", err
	}
	privatePEM := pem.EncodeToMemory(block)
	publicAuthorizedKey := ssh.MarshalAuthorizedKey(sshPublic)
	return base64.StdEncoding.EncodeToString(privatePEM),
		base64.StdEncoding.EncodeToString(publicAuthorizedKey), nil
}
