package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"gopkg.in/yaml.v3"
)

const (
	appName       = "cephtower"
	remoteRoot    = "/opt/cephtower"
	remoteBackup  = "/opt/cephtower.backup"
	remoteBinDir  = remoteRoot + "/bin"
	remoteConfDir = remoteRoot + "/config"
	remoteDataDir = remoteRoot + "/data"
	remoteLogDir  = remoteRoot + "/log"
)

var databaseEncryptionKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{32}$`)

type deployConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	KnownHosts string `yaml:"known_hosts"`
}

type targetConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	KnownHosts string `yaml:"known_hosts"`
}

type appConfig struct {
	Config     string `yaml:"config"`
	ReleaseDir string `yaml:"release_dir"`
}

type options struct {
	RepoRoot       string
	ToolDir        string
	ConfigPath     string
	Target         targetConfig
	App            appConfig
	Replace        replaceSet
	NonInteractive bool
}

type labState struct {
	Version     int           `json:"version"`
	ClusterName string        `json:"cluster_name"`
	RegionID    string        `json:"region_id"`
	Nodes       []labListNode `json:"nodes"`
}

type discoveredMachines struct {
	StatePath   string
	ClusterName string        `json:"cluster_name"`
	RegionID    string        `json:"region_id"`
	Nodes       []labListNode `json:"nodes"`
}

type labListNode struct {
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	PublicIP  string     `json:"public_ip"`
	PrivateIP string     `json:"private_ip"`
	SSH       labNodeSSH `json:"ssh"`
}

type labNodeSSH struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type remoteInfo struct {
	GOOS   string
	GOARCH string
	RawOS  string
	RawCPU string
}

type replaceSet map[string]bool

type logger struct {
	output io.Writer
	mu     sync.Mutex
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		newLogger(os.Stderr).Errorf("command failed: %v", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	log := newLogger(stderr)
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	toolDir := filepath.Join(repoRoot, "tools", "deploy-to-machine")
	opts, err := parseOptions(args, repoRoot, toolDir)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if opts.Target.Host == "" {
		target, err := chooseLabTarget(opts, stdin, stdout)
		if err != nil {
			return err
		}
		opts.Target = target
	}
	if err := validateTarget(opts.Target); err != nil {
		return err
	}

	log.Infof("ssh: connecting to %s:%d as %s", opts.Target.Host, opts.Target.Port, opts.Target.User)
	client, err := newRemoteClient(opts.Target)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Infof("remote: detecting machine architecture")
	info, err := client.Info(ctx)
	if err != nil {
		return err
	}
	log.Infof("remote: architecture=%s/%s raw_os=%s raw_cpu=%s", info.GOOS, info.GOARCH, info.RawOS, info.RawCPU)

	buildTarget := info.GOOS + "/" + info.GOARCH
	log.Infof("release: running make release TARGET=%s", buildTarget)
	if err := makeRelease(ctx, opts.RepoRoot, buildTarget, stdout, stderr); err != nil {
		return err
	}
	log.Infof("release: build completed target=%s", buildTarget)

	binaryPath, err := selectReleaseArtifact(opts.App.ReleaseDir, info.GOOS, info.GOARCH)
	if err != nil {
		return err
	}
	log.Infof("release: selected binary %s", binaryPath)

	encryptionKey, encryptionKeySource, err := client.DatabaseEncryptionKey(ctx)
	if err != nil {
		return err
	}
	log.Infof("config: using database.encryption_key source=%s", encryptionKeySource)

	configPayload, err := configWithServerDir(opts.App.Config, remoteRoot, encryptionKey)
	if err != nil {
		return err
	}
	log.Infof("config: prepared %s with server.dir=%s", opts.App.Config, remoteRoot)

	log.Infof("remote: stopping existing service if present")
	if err := client.Stop(ctx); err != nil {
		return err
	}
	log.Infof("remote: initializing target directories replace=%s", opts.Replace.String())
	if err := client.Initialize(ctx, opts.Replace); err != nil {
		return err
	}
	log.Infof("upload: sending binary to %s", remoteBinDir+"/"+appName)
	if err := client.UploadFile(ctx, binaryPath, remoteBinDir+"/"+appName, 0o755); err != nil {
		return err
	}
	log.Infof("upload: sending config to %s", remoteConfDir+"/config.yaml")
	if err := client.UploadContent(ctx, configPayload, remoteConfDir+"/config.yaml", 0o644); err != nil {
		return err
	}
	log.Infof("remote: starting service in background")
	pid, err := client.Start(ctx)
	if err != nil {
		return err
	}

	log.Infof("remote: service started host=%s pid=%s", opts.Target.Host, pid)
	return nil
}

func parseOptions(args []string, repoRoot, toolDir string) (options, error) {
	defaults := options{
		RepoRoot: repoRoot,
		ToolDir:  toolDir,
		ConfigPath: firstExistingPath(
			filepath.Join(toolDir, "config.local.yaml"),
			filepath.Join(toolDir, "config.yaml"),
		),
		Target: targetConfig{Port: 22, User: "root", KnownHosts: ".state/known_hosts"},
		App: appConfig{
			Config:     filepath.Join(repoRoot, "config", "config.yaml"),
			ReleaseDir: filepath.Join(repoRoot, "dist"),
		},
	}
	flags := flag.NewFlagSet("deploy-to-machine", flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	configPath := flags.String("config", defaults.ConfigPath, "deploy YAML configuration")
	host := flags.String("host", "", "target SSH host or IP")
	port := flags.Int("port", 0, "target SSH port")
	user := flags.String("user", "", "target SSH user")
	password := flags.String("password", "", "target SSH password")
	knownHosts := flags.String("known-hosts", "", "known_hosts file path")
	appConfigPath := flags.String("app-config", "", "CephTower config.yaml path")
	releaseDir := flags.String("release-dir", "", "CephTower release artifact directory")
	replaceValue := flags.String("replace", "", "remote directories to replace: bin,conf,data,log,all")
	nonInteractive := flags.Bool("non-interactive", false, "fail instead of prompting for lab node selection")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}

	opts := defaults
	opts.ConfigPath = *configPath
	if opts.ConfigPath != "" {
		cfg, err := loadDeployConfig(resolvePath(opts.ConfigPath, toolDir))
		if err != nil {
			return options{}, err
		}
		opts.Target = mergeDeployConfig(opts.Target, cfg)
	}
	overrideString(&opts.Target.Host, *host)
	if *port != 0 {
		opts.Target.Port = *port
	}
	overrideString(&opts.Target.User, *user)
	overrideString(&opts.Target.Password, *password)
	overrideString(&opts.Target.KnownHosts, *knownHosts)
	overrideString(&opts.App.Config, *appConfigPath)
	overrideString(&opts.App.ReleaseDir, *releaseDir)
	opts.NonInteractive = *nonInteractive

	replace, err := parseReplace(*replaceValue)
	if err != nil {
		return options{}, err
	}
	opts.Replace = replace
	opts.Target.KnownHosts = resolvePath(opts.Target.KnownHosts, toolDir)
	opts.App.Config = resolvePath(opts.App.Config, toolDir)
	opts.App.ReleaseDir = resolvePath(opts.App.ReleaseDir, toolDir)
	return opts, nil
}

func firstExistingPath(paths ...string) string {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	if len(paths) == 0 {
		return ""
	}
	return paths[len(paths)-1]
}

func loadDeployConfig(path string) (deployConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return deployConfig{}, nil
		}
		return deployConfig{}, fmt.Errorf("read deploy config %s: %w", path, err)
	}
	var cfg deployConfig
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return deployConfig{}, fmt.Errorf("decode deploy config %s: %w", path, err)
	}
	return cfg, nil
}

func mergeDeployConfig(base targetConfig, override deployConfig) targetConfig {
	overrideString(&base.Host, override.Host)
	if override.Port != 0 {
		base.Port = override.Port
	}
	overrideString(&base.User, override.User)
	overrideString(&base.Password, override.Password)
	overrideString(&base.KnownHosts, override.KnownHosts)
	return base
}

func overrideString(target *string, value string) {
	if value != "" {
		*target = value
	}
}

func resolvePath(path, baseDir string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(baseDir, path))
}

func parseReplace(value string) (replaceSet, error) {
	result := replaceSet{}
	for _, item := range strings.Split(value, ",") {
		item = strings.ToLower(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		if item == "config" {
			item = "conf"
		}
		switch item {
		case "bin", "conf", "data", "log", "all":
			result[item] = true
		default:
			return nil, fmt.Errorf("invalid --replace value %q", item)
		}
	}
	return result, nil
}

func (r replaceSet) String() string {
	if len(r) == 0 {
		return "none"
	}
	if r["all"] {
		return "all"
	}
	var values []string
	for _, value := range []string{"bin", "conf", "data", "log"} {
		if r[value] {
			values = append(values, value)
		}
	}
	return strings.Join(values, ",")
}

func newLogger(output io.Writer) *logger {
	return &logger{output: output}
}

func (l *logger) Infof(format string, args ...any) {
	l.writef("INFO", format, args...)
}

func (l *logger) Errorf(format string, args ...any) {
	l.writef("ERROR", format, args...)
}

func (l *logger) writef(level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = fmt.Fprintf(l.output, "[%s] %s %s\n", time.Now().Format(time.RFC3339), level, fmt.Sprintf(format, args...))
	if file, ok := l.output.(*os.File); ok {
		_ = file.Sync()
	}
}

func chooseLabTarget(opts options, stdin io.Reader, stdout io.Writer) (targetConfig, error) {
	machines, err := discoverLabMachines(opts.RepoRoot)
	if err != nil {
		return targetConfig{}, err
	}
	nodes := flattenDiscoveredNodes(machines)
	if len(nodes) == 0 {
		return targetConfig{}, errors.New("no machines found; configure --host or create the aliyun lab first")
	}
	fmt.Fprintln(stdout, "Available machines:")
	for i, node := range nodes {
		host := nodeHost(node)
		fmt.Fprintf(stdout, "  %d. %s cluster=%s status=%s public=%s private=%s user=%s source=%s\n",
			i+1, node.Name, emptyAsDash(node.ClusterName), emptyAsDash(node.Status),
			emptyAsDash(node.PublicIP), emptyAsDash(node.PrivateIP), emptyAsDash(node.SSH.User),
			emptyAsDash(filepath.Base(node.StatePath)))
		if host == "" {
			fmt.Fprintln(stdout, "     warning: no SSH host or public IP is recorded for this node")
		}
	}
	if opts.NonInteractive {
		return targetConfig{}, errors.New("target host is not configured and --non-interactive was set")
	}
	fmt.Fprint(stdout, "Select a machine by number: ")
	var line string
	if _, err := fmt.Fscanln(stdin, &line); err != nil {
		return targetConfig{}, fmt.Errorf("read machine selection: %w", err)
	}
	index, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || index < 1 || index > len(nodes) {
		return targetConfig{}, fmt.Errorf("invalid machine selection %q", line)
	}
	node := nodes[index-1]
	host := nodeHost(node)
	if host == "" {
		return targetConfig{}, fmt.Errorf("selected machine %s has no SSH host or public IP", node.Name)
	}
	port := node.SSH.Port
	if port == 0 {
		port = 22
	}
	user := node.SSH.User
	if user == "" {
		user = opts.Target.User
	}
	return targetConfig{
		Host:       host,
		Port:       port,
		User:       user,
		Password:   node.SSH.Password,
		KnownHosts: opts.Target.KnownHosts,
	}, nil
}

type discoveredNode struct {
	labListNode
	ClusterName string
	RegionID    string
	StatePath   string
}

func discoverLabMachines(repoRoot string) ([]discoveredMachines, error) {
	stateFiles, err := findLabStateFiles(repoRoot)
	if err != nil {
		return nil, err
	}
	if len(stateFiles) == 0 {
		return nil, fmt.Errorf("no aliyun lab state JSON files found under %s",
			filepath.Join(repoRoot, "tools", "aliyun-ceph-lab", ".state"))
	}
	machines := make([]discoveredMachines, 0, len(stateFiles))
	var decodeErrors []string
	for _, path := range stateFiles {
		raw, err := os.ReadFile(path)
		if err != nil {
			decodeErrors = append(decodeErrors, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		var state labState
		if err := json.Unmarshal(raw, &state); err != nil {
			decodeErrors = append(decodeErrors, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		if len(state.Nodes) == 0 {
			continue
		}
		machines = append(machines, discoveredMachines{
			StatePath:   path,
			ClusterName: state.ClusterName,
			RegionID:    state.RegionID,
			Nodes:       state.Nodes,
		})
	}
	if len(machines) == 0 && len(decodeErrors) > 0 {
		return nil, fmt.Errorf("no usable aliyun lab state files found: %s", strings.Join(decodeErrors, "; "))
	}
	return machines, nil
}

func findLabStateFiles(repoRoot string) ([]string, error) {
	var candidates []string
	addCandidate := func(path string) {
		for _, current := range candidates {
			if current == path {
				return
			}
		}
		candidates = append(candidates, path)
	}
	addCandidate(filepath.Join(repoRoot, "tools", "aliyun-ceph-lab", ".state"))
	toolsDir := filepath.Join(repoRoot, "tools")
	if err := filepath.WalkDir(toolsDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		if path != toolsDir && strings.HasPrefix(entry.Name(), ".") && entry.Name() != ".state" {
			return filepath.SkipDir
		}
		if entry.Name() != ".state" {
			return nil
		}
		parent := filepath.Dir(path)
		relative, err := filepath.Rel(toolsDir, parent)
		if err != nil {
			return nil
		}
		if strings.Contains(strings.ToLower(relative), "aliyun-ceph-lab") {
			addCandidate(path)
			return filepath.SkipDir
		}
		return nil
	}); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("scan tools directory for lab state: %w", err)
	}
	seen := map[string]bool{}
	var files []string
	for _, dir := range candidates {
		if seen[dir] {
			continue
		}
		seen[dir] = true
		entries, err := os.ReadDir(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("read lab state directory %s: %w", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".json" {
				continue
			}
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func flattenDiscoveredNodes(machines []discoveredMachines) []discoveredNode {
	var nodes []discoveredNode
	for _, machine := range machines {
		for _, node := range machine.Nodes {
			nodes = append(nodes, discoveredNode{
				labListNode: node,
				ClusterName: machine.ClusterName,
				RegionID:    machine.RegionID,
				StatePath:   machine.StatePath,
			})
		}
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].ClusterName != nodes[j].ClusterName {
			return nodes[i].ClusterName < nodes[j].ClusterName
		}
		return nodes[i].Name < nodes[j].Name
	})
	return nodes
}

func nodeHost(node discoveredNode) string {
	if node.SSH.Host != "" {
		return node.SSH.Host
	}
	if node.PublicIP != "" {
		return node.PublicIP
	}
	return node.PrivateIP
}

func emptyAsDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func validateTarget(target targetConfig) error {
	if strings.TrimSpace(target.Host) == "" {
		return errors.New("target host is required")
	}
	if target.Port == 0 {
		return errors.New("target port is required")
	}
	if strings.TrimSpace(target.User) == "" {
		return errors.New("target user is required")
	}
	if target.Password == "" {
		return errors.New("target password is required")
	}
	return nil
}

func makeRelease(ctx context.Context, repoRoot, target string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, "make", "release", "TARGET="+target)
	cmd.Dir = repoRoot
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make release: %w", err)
	}
	return nil
}

func selectReleaseArtifact(releaseDir, goos, goarch string) (string, error) {
	suffix := "-" + goos + "-" + goarch
	if goos == "windows" {
		suffix += ".exe"
	}
	matches, err := filepath.Glob(filepath.Join(releaseDir, appName+"-*"+suffix))
	if err != nil {
		return "", fmt.Errorf("glob release artifacts: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no release artifact found for %s/%s in %s", goos, goarch, releaseDir)
	}
	sort.Slice(matches, func(i, j int) bool {
		left, leftErr := os.Stat(matches[i])
		right, rightErr := os.Stat(matches[j])
		if leftErr != nil || rightErr != nil {
			return matches[i] > matches[j]
		}
		return left.ModTime().After(right.ModTime())
	})
	return matches[0], nil
}

func configWithServerDir(path, serverDir, encryptionKey string) ([]byte, error) {
	if err := validateDatabaseEncryptionKey(encryptionKey); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read app config %s: %w", path, err)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("decode app config %s: %w", path, err)
	}
	if len(root.Content) == 0 {
		return nil, fmt.Errorf("app config %s is empty", path)
	}
	doc := root.Content[0]
	server := mappingValue(doc, "server")
	if server == nil {
		server = &yaml.Node{Kind: yaml.MappingNode}
		appendMappingValue(doc, "server", server)
	}
	setMappingScalar(server, "dir", serverDir)

	database := mappingValue(doc, "database")
	if database == nil {
		database = &yaml.Node{Kind: yaml.MappingNode}
		appendMappingValue(doc, "database", database)
	}
	setMappingScalar(database, "encryption_key", encryptionKey)
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(4)
	if err := encoder.Encode(&root); err != nil {
		return nil, fmt.Errorf("encode app config: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("close YAML encoder: %w", err)
	}
	return out.Bytes(), nil
}

func encryptionKeyFromConfig(raw []byte) (string, bool, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return "", false, nil
	}
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return "", false, fmt.Errorf("decode remote config: %w", err)
	}
	if len(root.Content) == 0 {
		return "", false, nil
	}
	database := mappingValue(root.Content[0], "database")
	if database == nil {
		return "", false, nil
	}
	keyNode := mappingValue(database, "encryption_key")
	if keyNode == nil || keyNode.Kind != yaml.ScalarNode || strings.TrimSpace(keyNode.Value) == "" {
		return "", false, nil
	}
	key := strings.TrimSpace(keyNode.Value)
	if err := validateDatabaseEncryptionKey(key); err != nil {
		return "", false, fmt.Errorf("remote database.encryption_key is invalid: %w", err)
	}
	return key, true, nil
}

func validateDatabaseEncryptionKey(key string) error {
	if !databaseEncryptionKeyPattern.MatchString(key) {
		return errors.New("database.encryption_key must contain exactly 32 characters from A-Z, a-z, 0-9, _ and -")
	}
	return nil
}

func generateDatabaseEncryptionKey() (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate database.encryption_key: %w", err)
	}
	for i, value := range raw {
		raw[i] = alphabet[int(value)%len(alphabet)]
	}
	return string(raw), nil
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func appendMappingValue(node *yaml.Node, key string, value *yaml.Node) {
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		value,
	)
}

func setMappingScalar(node *yaml.Node, key, value string) {
	current := mappingValue(node, key)
	if current == nil {
		appendMappingValue(node, key, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value})
		return
	}
	current.Kind = yaml.ScalarNode
	current.Tag = "!!str"
	current.Value = value
}

type remoteClient struct {
	target targetConfig
	client *ssh.Client
}

func newRemoteClient(target targetConfig) (*remoteClient, error) {
	if err := os.MkdirAll(filepath.Dir(target.KnownHosts), 0o700); err != nil {
		return nil, fmt.Errorf("create known_hosts directory: %w", err)
	}
	file, err := os.OpenFile(target.KnownHosts, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open known_hosts: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close known_hosts: %w", err)
	}
	callback, err := knownhosts.New(target.KnownHosts)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts: %w", err)
	}
	address := net.JoinHostPort(target.Host, strconv.Itoa(target.Port))
	client, err := ssh.Dial("tcp", address, &ssh.ClientConfig{
		User:            target.User,
		Auth:            []ssh.AuthMethod{ssh.Password(target.Password)},
		HostKeyCallback: learnHostKey(target.KnownHosts, callback),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", address, err)
	}
	return &remoteClient{target: target, client: client}, nil
}

func learnHostKey(path string, callback ssh.HostKeyCallback) ssh.HostKeyCallback {
	var mu sync.Mutex
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if err := callback(hostname, remote, key); err == nil {
			return nil
		} else {
			var keyError *knownhosts.KeyError
			if !errors.As(err, &keyError) || len(keyError.Want) != 0 {
				return fmt.Errorf("verify SSH host key: %w", err)
			}
		}
		mu.Lock()
		defer mu.Unlock()
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("append known_hosts: %w", err)
		}
		defer file.Close()
		line := knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key) + "\n"
		if _, err := io.WriteString(file, line); err != nil {
			return fmt.Errorf("write known_hosts: %w", err)
		}
		return nil
	}
}

func (r *remoteClient) Close() error {
	return r.client.Close()
}

func (r *remoteClient) Info(ctx context.Context) (remoteInfo, error) {
	output, err := r.Run(ctx, "printf '%s\\n' \"$(uname -s)\" \"$(uname -m)\"")
	if err != nil {
		return remoteInfo{}, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return remoteInfo{}, fmt.Errorf("unexpected uname output %q", output)
	}
	return normalizeRemoteOSArch(lines[0], lines[1])
}

func (r *remoteClient) DatabaseEncryptionKey(ctx context.Context) (string, string, error) {
	output, err := r.Run(ctx, "if test -r "+shellQuote(remoteConfDir+"/config.yaml")+"; then cat "+shellQuote(remoteConfDir+"/config.yaml")+"; fi")
	if err != nil {
		return "", "", fmt.Errorf("read remote config for database.encryption_key: %w", err)
	}
	if key, ok, err := encryptionKeyFromConfig([]byte(output)); err != nil {
		return "", "", err
	} else if ok {
		return key, "remote", nil
	}
	key, err := generateDatabaseEncryptionKey()
	if err != nil {
		return "", "", err
	}
	return key, "generated", nil
}

func normalizeRemoteOSArch(rawOS, rawCPU string) (remoteInfo, error) {
	info := remoteInfo{RawOS: strings.TrimSpace(rawOS), RawCPU: strings.TrimSpace(rawCPU)}
	switch strings.ToLower(info.RawOS) {
	case "linux":
		info.GOOS = "linux"
	case "darwin":
		info.GOOS = "darwin"
	default:
		return remoteInfo{}, fmt.Errorf("unsupported remote OS %q", rawOS)
	}
	switch strings.ToLower(info.RawCPU) {
	case "x86_64", "amd64":
		info.GOARCH = "amd64"
	case "aarch64", "arm64":
		info.GOARCH = "arm64"
	default:
		return remoteInfo{}, fmt.Errorf("unsupported remote CPU architecture %q", rawCPU)
	}
	return info, nil
}

func (r *remoteClient) Stop(ctx context.Context) error {
	processPattern := strings.Replace(remoteBinDir+"/"+appName, "cephtower", "[c]ephtower", 1)
	command := "if test -f " + shellQuote(remoteRoot+"/cephtower.pid") + "; then " +
		"pid=$(cat " + shellQuote(remoteRoot+"/cephtower.pid") + " 2>/dev/null || true); " +
		"if test -n \"$pid\" && kill -0 \"$pid\" 2>/dev/null; then kill \"$pid\" || true; " +
		"for i in $(seq 1 20); do kill -0 \"$pid\" 2>/dev/null || break; sleep 0.2; done; " +
		"kill -9 \"$pid\" 2>/dev/null || true; fi; fi; " +
		"pkill -f " + shellQuote(processPattern) + " 2>/dev/null || true"
	_, err := r.Run(ctx, command)
	return err
}

func (r *remoteClient) Initialize(ctx context.Context, replace replaceSet) error {
	commands := []string{}
	if replace["all"] {
		commands = append(commands, backupRemoteRootCommand(), "rm -rf "+shellQuote(remoteRoot))
	} else {
		if replace["bin"] {
			commands = append(commands, "rm -rf "+shellQuote(remoteBinDir))
		}
		if replace["conf"] {
			commands = append(commands, "rm -rf "+shellQuote(remoteConfDir))
		}
		if replace["data"] {
			commands = append(commands, "rm -rf "+shellQuote(remoteDataDir))
		}
		if replace["log"] {
			commands = append(commands, "rm -rf "+shellQuote(remoteLogDir))
		}
	}
	commands = append(commands,
		"install -d -m 0755 "+shellQuote(remoteBinDir)+" "+shellQuote(remoteConfDir)+" "+shellQuote(remoteDataDir)+" "+shellQuote(remoteLogDir),
	)
	_, err := r.Run(ctx, strings.Join(commands, " && "))
	return err
}

func backupRemoteRootCommand() string {
	return "if test -d " + shellQuote(remoteRoot) + "; then " +
		"backup_dir=" + shellQuote(remoteBackup) + "/$(date +%Y%m%d%H%M%S); " +
		"install -d -m 0755 \"$backup_dir\"; " +
		"cp -a " + shellQuote(remoteRoot) + "/. \"$backup_dir\"; " +
		"fi"
}

func (r *remoteClient) UploadFile(ctx context.Context, localPath, remotePath string, mode os.FileMode) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", localPath, err)
	}
	defer file.Close()
	return r.upload(ctx, file, remotePath, mode)
}

func (r *remoteClient) UploadContent(ctx context.Context, payload []byte, remotePath string, mode os.FileMode) error {
	return r.upload(ctx, bytes.NewReader(payload), remotePath, mode)
}

func (r *remoteClient) upload(ctx context.Context, input io.Reader, remotePath string, mode os.FileMode) error {
	tmpPath := fmt.Sprintf("%s.tmp.%d", remotePath, time.Now().UnixNano())
	command := "cat > " + shellQuote(tmpPath) +
		" && chmod " + shellQuote(fmt.Sprintf("%04o", mode.Perm())) + " " + shellQuote(tmpPath) +
		" && mv -f " + shellQuote(tmpPath) + " " + shellQuote(remotePath)
	if _, err := r.run(ctx, command, input); err != nil {
		return fmt.Errorf("upload %s: %w", remotePath, err)
	}
	return nil
}

func (r *remoteClient) Start(ctx context.Context) (string, error) {
	startScript := "nohup " + shellQuote(remoteBinDir+"/"+appName) +
		" -config " + shellQuote(remoteConfDir+"/config.yaml") +
		" >> " + shellQuote(remoteLogDir+"/cephtower.stdout.log") + " 2>&1 < /dev/null & " +
		"pid=$!; echo \"$pid\" > " + shellQuote(remoteRoot+"/cephtower.pid") +
		"; sleep 1; kill -0 \"$pid\"; printf '%s' \"$pid\""
	command := "cd " + shellQuote(remoteRoot) +
		" && install -d -m 0755 " + shellQuote(remoteLogDir) + " " + shellQuote(remoteDataDir) +
		" && sh -c " + shellQuote(startScript)
	output, err := r.Run(ctx, command)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *remoteClient) Run(ctx context.Context, command string) (string, error) {
	return r.run(ctx, command, nil)
}

func (r *remoteClient) run(ctx context.Context, command string, stdin io.Reader) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()
	session.Stdin = stdin
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if r.target.User != "root" {
		command = "sudo -n sh -c " + shellQuote(command)
	}
	done := make(chan error, 1)
	go func() { done <- session.Run(command) }()
	select {
	case err := <-done:
		if err != nil {
			return stdout.String(), fmt.Errorf("remote command failed: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return stdout.String(), nil
	case <-ctx.Done():
		_ = session.Close()
		<-done
		return stdout.String(), ctx.Err()
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for {
		if fileExists(filepath.Join(wd, "Makefile")) && fileExists(filepath.Join(wd, "config", "config.yaml")) {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", errors.New("could not locate CephTower repo root")
		}
		wd = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
