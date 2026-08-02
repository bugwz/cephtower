package lab

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	dashboardPassword, err := generateDashboardPassword()
	if err != nil {
		return fmt.Errorf("generate Ceph Dashboard password: %w", err)
	}
	args := []string{
		"--cluster-name", m.Config.ClusterName,
		"--bootstrap-node-name", first.Name,
		"--node-names", joinNodeField(current.Nodes, func(node state.Node) string { return node.Name }),
		"--public-ips", joinNodeField(current.Nodes, func(node state.Node) string { return node.PublicIP }),
		"--private-ips", joinNodeField(current.Nodes, func(node state.Node) string { return node.PrivateIP }),
		"--data-disk-counts", joinDataDiskCounts(m.Config.Nodes),
		"--wait-timeout-seconds", fmt.Sprintf("%d", int(m.Config.WaitTimeoutDuration().Seconds())),
		"--dashboard-password", dashboardPassword,
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
	collectCtx, collectCancel := context.WithTimeout(ctx, m.Config.WaitTimeoutDuration())
	defer collectCancel()
	cephInfo, err := m.collectCephInfo(collectCtx, current, deployConnection, dashboardPassword)
	if err != nil {
		return fmt.Errorf("collect ceph connection information: %w", err)
	}
	current.Ceph = cephInfo
	if err := state.Save(m.StatePath, current); err != nil {
		return fmt.Errorf("save ceph connection information: %w", err)
	}
	logging.Infof("create: Ceph connection information saved; dashboard=%s mon=%s", cephInfo.Dashboard.URL, cephInfo.Monitors.MonitorAddresses)
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

type cephCollectWire struct {
	ClusterName        string      `json:"cluster_name"`
	FSID               string      `json:"fsid"`
	ClientAdminKey     string      `json:"client_admin_key"`
	ClientAdminKeyring string      `json:"client_admin_keyring"`
	DashboardURL       string      `json:"dashboard_url"`
	DashboardUsername  string      `json:"dashboard_username"`
	DashboardPassword  string      `json:"dashboard_password"`
	MonDump            monDumpWire `json:"mon_dump"`
}

type monDumpWire struct {
	Mons []monWire `json:"mons"`
}

type monWire struct {
	Name        string `json:"name"`
	Addr        string `json:"addr"`
	PublicAddrs struct {
		Addrvec []monAddrvecWire `json:"addrvec"`
	} `json:"public_addrs"`
}

type monAddrvecWire struct {
	Type  string `json:"type"`
	Addr  string `json:"addr"`
	Nonce uint64 `json:"nonce"`
}

func (m *Manager) collectCephInfo(
	ctx context.Context,
	current *state.State,
	connection *remote.HostConnection,
	dashboardPassword string,
) (*state.Ceph, error) {
	fallbackDashboardURL := ""
	if len(current.Nodes) > 0 && current.Nodes[0].PublicIP != "" {
		fallbackDashboardURL = fmt.Sprintf("https://%s:8443/", current.Nodes[0].PublicIP)
	}
	command := "CEPH_LAB_DASHBOARD_PASSWORD=" + shellQuote(dashboardPassword) + " " +
		"CEPH_LAB_DASHBOARD_URL=" + shellQuote(fallbackDashboardURL) + " bash -s"
	output, err := connection.CombinedOutput(ctx, command, strings.NewReader(cephCollectScript))
	if err != nil {
		return nil, err
	}
	var wire cephCollectWire
	if err := json.Unmarshal(output, &wire); err != nil {
		return nil, fmt.Errorf("decode ceph connection information: %w", err)
	}
	if strings.TrimSpace(wire.ClientAdminKey) == "" {
		return nil, errors.New("client.admin key is empty")
	}
	monitors, err := buildCephMonitors(wire.MonDump)
	if err != nil {
		return nil, err
	}
	clientUsername := "client.admin"
	clusterName := strings.TrimSpace(wire.ClusterName)
	if clusterName == "" {
		clusterName = "ceph"
	}
	dashboardUsername := strings.TrimSpace(wire.DashboardUsername)
	if dashboardUsername == "" {
		dashboardUsername = "admin"
	}
	return &state.Ceph{
		ClusterName: clusterName,
		FSID:        strings.TrimSpace(wire.FSID),
		ClientAdmin: state.CephClientAdmin{
			Username: clientUsername,
			Key:      strings.TrimSpace(wire.ClientAdminKey),
			Keyring:  wire.ClientAdminKeyring,
		},
		Dashboard: state.CephDashboard{
			URL:      strings.TrimSpace(wire.DashboardURL),
			Username: dashboardUsername,
			Password: wire.DashboardPassword,
		},
		Monitors: monitors,
		CephTowerClusterCreate: state.CephTowerCreate{
			Name:             current.ClusterName,
			MonitorAddresses: monitors.MonitorAddresses,
			ClientUsername:   clientUsername,
			ClientKey:        strings.TrimSpace(wire.ClientAdminKey),
		},
	}, nil
}

const cephCollectScript = `set -Eeuo pipefail
dashboard_url="${CEPH_LAB_DASHBOARD_URL}"
if [[ -z "${dashboard_url}" ]]; then
	dashboard_url="$(ceph mgr services --format json | jq -r '.dashboard // empty')"
fi
jq -n \
	--arg cluster_name "ceph" \
	--arg fsid "$(ceph fsid | tr -d '\r\n')" \
	--arg client_admin_key "$(ceph auth get-key client.admin | tr -d '\r\n')" \
	--rawfile client_admin_keyring /etc/ceph/ceph.client.admin.keyring \
	--arg dashboard_url "${dashboard_url}" \
	--arg dashboard_username "admin" \
	--arg dashboard_password "${CEPH_LAB_DASHBOARD_PASSWORD}" \
	--slurpfile mon_dump <(ceph mon dump --format json) \
	'{
		cluster_name: $cluster_name,
		fsid: $fsid,
		client_admin_key: $client_admin_key,
		client_admin_keyring: $client_admin_keyring,
		dashboard_url: $dashboard_url,
		dashboard_username: $dashboard_username,
		dashboard_password: $dashboard_password,
		mon_dump: $mon_dump[0]
	}'
`

func buildCephMonitors(dump monDumpWire) (state.CephMonitors, error) {
	groups := make([]string, 0, len(dump.Mons))
	var v1 []string
	var v2 []string
	var endpoints []state.CephMonitorEndpoint
	for _, mon := range dump.Mons {
		addrvec := mon.PublicAddrs.Addrvec
		if len(addrvec) == 0 && strings.TrimSpace(mon.Addr) != "" {
			addrvec = []monAddrvecWire{{Type: "v1", Addr: strings.TrimSpace(mon.Addr)}}
		}
		group := make([]string, 0, len(addrvec))
		for _, item := range addrvec {
			endpoint, formatted, err := cephMonitorEndpoint(mon.Name, item)
			if err != nil {
				return state.CephMonitors{}, err
			}
			endpoints = append(endpoints, endpoint)
			group = append(group, formatted)
			switch endpoint.Protocol {
			case "v1":
				v1 = append(v1, formatted)
			case "v2":
				v2 = append(v2, formatted)
			}
		}
		if len(group) == 1 {
			groups = append(groups, group[0])
		} else if len(group) > 1 {
			groups = append(groups, "["+strings.Join(group, ",")+"]")
		}
	}
	if len(groups) == 0 {
		return state.CephMonitors{}, errors.New("ceph mon dump contained no monitor addresses")
	}
	return state.CephMonitors{
		MonitorAddresses: strings.Join(groups, ","),
		V1Addresses:      strings.Join(v1, ","),
		V2Addresses:      strings.Join(v2, ","),
		Endpoints:        endpoints,
	}, nil
}

func cephMonitorEndpoint(name string, item monAddrvecWire) (state.CephMonitorEndpoint, string, error) {
	protocol := strings.TrimSpace(item.Type)
	if protocol == "" {
		protocol = "v1"
	}
	if protocol != "v1" && protocol != "v2" {
		return state.CephMonitorEndpoint{}, "", fmt.Errorf("unsupported monitor protocol %q", protocol)
	}
	address := strings.TrimSpace(item.Addr)
	if address == "" {
		return state.CephMonitorEndpoint{}, "", fmt.Errorf("monitor %s has empty address", name)
	}
	host, port := splitCephAddress(address)
	formatted := protocol + ":" + address
	if !strings.Contains(address, "/") {
		formatted = fmt.Sprintf("%s/%d", formatted, item.Nonce)
	}
	return state.CephMonitorEndpoint{
		Name:     name,
		Protocol: protocol,
		Address:  address,
		Host:     host,
		Port:     port,
		Nonce:    item.Nonce,
	}, formatted, nil
}

func splitCephAddress(address string) (string, uint16) {
	value := address
	if slash := strings.LastIndex(value, "/"); slash >= 0 {
		value = value[:slash]
	}
	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		colon := strings.LastIndex(value, ":")
		if colon <= 0 || strings.Contains(value[:colon], ":") {
			return strings.Trim(value, "[]"), 0
		}
		host, portText = value[:colon], value[colon+1:]
	}
	parsed, err := strconv.ParseUint(portText, 10, 16)
	if err != nil {
		return strings.Trim(host, "[]"), 0
	}
	return strings.Trim(host, "[]"), uint16(parsed)
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

func generateDashboardPassword() (string, error) {
	const length = 20
	const lower = "abcdefghijkmnopqrstuvwxyz"
	const upper = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	const digits = "23456789"
	const symbols = "#%+-_."
	alphabet := lower + upper + digits + symbols
	password := []byte{lower[0], upper[0], digits[0], symbols[0]}
	for len(password) < length {
		character, err := randomCharacter(alphabet)
		if err != nil {
			return "", err
		}
		password = append(password, character)
	}
	for index := len(password) - 1; index > 0; index-- {
		other, err := rand.Int(rand.Reader, big.NewInt(int64(index+1)))
		if err != nil {
			return "", err
		}
		password[index], password[other.Int64()] = password[other.Int64()], password[index]
	}
	return string(password), nil
}

func randomCharacter(alphabet string) (byte, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
	if err != nil {
		return 0, err
	}
	return alphabet[index.Int64()], nil
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
