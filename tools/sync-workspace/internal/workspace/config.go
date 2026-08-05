package workspace

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	remoteRoot       = "/root/cephtower"
	remoteStateDir   = remoteRoot + "/.sync-workspace"
	remoteManifest   = remoteStateDir + "/source-manifest.json"
	remoteAppDir     = remoteRoot + "/app"
	remoteLogDir     = remoteAppDir + "/log"
	remoteStdLog     = remoteLogDir + "/std.log"
	remoteAppLog     = remoteLogDir + "/cephtower.log"
	remoteRuntimeDir = remoteAppDir + "/data/runtime"
	remotePIDFile    = remoteRuntimeDir + "/server.pid"
	remoteConfigPath = remoteAppDir + "/config/config.yaml"
	webPort          = 36900
)

type command string

const (
	commandRun    command = "run"
	commandStop   command = "stop"
	commandStatus command = "status"
)

type fileConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	KnownHosts   string `yaml:"known_hosts"`
	CleanOnStart bool   `yaml:"clean_on_start"`
}

type targetConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	KnownHosts string `json:"known_hosts"`
}

type options struct {
	Command        command
	RepoRoot       string
	ToolDir        string
	ConfigPath     string
	ConfigExplicit bool
	StatePath      string
	ManifestPath   string
	Target         targetConfig
	CleanOnStart   bool
}

type workspaceState struct {
	Version    int          `json:"version"`
	UpdatedAt  time.Time    `json:"updated_at"`
	Target     targetConfig `json:"target"`
	RemoteRoot string       `json:"remote_root"`
	WebPort    int          `json:"web_port"`
	LogDir     string       `json:"log_dir,omitempty"`
}

func parseOptions(args []string, repoRoot, toolDir string, output io.Writer) (options, error) {
	if len(args) == 0 {
		printUsage(output)
		return options{}, errors.New("missing command")
	}
	var selected command
	switch args[0] {
	case "-h", "--help":
		printUsage(output)
		return options{}, flag.ErrHelp
	case string(commandRun):
		selected = commandRun
	case string(commandStop):
		selected = commandStop
	case string(commandStatus):
		selected = commandStatus
	default:
		return options{}, fmt.Errorf("unknown command %q", args[0])
	}

	defaultConfig := firstExistingPath(
		filepath.Join(toolDir, "config.local.yaml"),
		filepath.Join(toolDir, "config.yaml"),
	)
	flags := flag.NewFlagSet("sync-workspace "+string(selected), flag.ContinueOnError)
	flags.SetOutput(output)
	configPath := flags.String("config", defaultConfig, "workspace YAML configuration")
	flags.Usage = func() {
		fmt.Fprintf(output, "Usage: sync-workspace %s [--config path]\n", selected)
	}
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, err
	}
	if flags.NArg() != 0 {
		return options{}, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}

	opts := options{
		Command:      selected,
		RepoRoot:     repoRoot,
		ToolDir:      toolDir,
		StatePath:    filepath.Join(toolDir, ".state", "workspace.json"),
		ManifestPath: filepath.Join(toolDir, ".state", "manifest.json"),
		Target: targetConfig{
			Port:       22,
			User:       "root",
			KnownHosts: ".state/known_hosts",
		},
	}
	flags.Visit(func(item *flag.Flag) {
		if item.Name == "config" {
			opts.ConfigExplicit = true
		}
	})
	opts.ConfigPath = resolvePath(*configPath, toolDir)
	configDir := toolDir
	if opts.ConfigPath != "" {
		cfg, err := loadFileConfig(opts.ConfigPath)
		if err != nil {
			return options{}, err
		}
		configDir = filepath.Dir(opts.ConfigPath)
		opts.Target = mergeTarget(opts.Target, cfg)
		opts.CleanOnStart = cfg.CleanOnStart
	}
	opts.Target.KnownHosts = resolvePath(opts.Target.KnownHosts, configDir)
	return opts, nil
}

func printUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage: sync-workspace <run|stop|status> [--config path]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Commands:")
	fmt.Fprintln(output, "  run       synchronize sources, build, run, and collect logs")
	fmt.Fprintln(output, "  stop      stop the last remote development service")
	fmt.Fprintln(output, "  status    show the last remote development service status")
}

func loadFileConfig(path string) (fileConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fileConfig{}, nil
		}
		return fileConfig{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg fileConfig
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return fileConfig{}, fmt.Errorf("decode config %s: %w", path, err)
	}
	return cfg, nil
}

func mergeTarget(base targetConfig, cfg fileConfig) targetConfig {
	if cfg.Host != "" {
		base.Host = cfg.Host
	}
	if cfg.Port != 0 {
		base.Port = cfg.Port
	}
	if cfg.User != "" {
		base.User = cfg.User
	}
	if cfg.Password != "" {
		base.Password = cfg.Password
	}
	if cfg.KnownHosts != "" {
		base.KnownHosts = cfg.KnownHosts
	}
	return base
}

func validateTarget(target targetConfig) error {
	if strings.TrimSpace(target.Host) == "" {
		return errors.New("target host is required")
	}
	if target.Port < 1 || target.Port > 65535 {
		return errors.New("target port must be between 1 and 65535")
	}
	if strings.TrimSpace(target.User) == "" {
		return errors.New("target user is required")
	}
	if target.Password == "" {
		return errors.New("target password is required")
	}
	if target.KnownHosts == "" {
		return errors.New("known_hosts path is required")
	}
	return nil
}

func saveWorkspaceState(path string, state workspaceState) error {
	state.Version = 1
	state.UpdatedAt = time.Now()
	state.RemoteRoot = remoteRoot
	state.WebPort = webPort
	return writeJSONFile(path, state, 0o600)
}

func loadWorkspaceState(path string) (workspaceState, error) {
	var state workspaceState
	raw, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return state, fmt.Errorf("decode state %s: %w", path, err)
	}
	if state.Version != 1 {
		return state, fmt.Errorf("unsupported state version %d", state.Version)
	}
	if err := validateTarget(state.Target); err != nil {
		return state, fmt.Errorf("invalid state target: %w", err)
	}
	return state, nil
}

func writeJSONFile(path string, value any, mode os.FileMode) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(payload); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
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

func resolvePath(path, baseDir string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(baseDir, path))
}
