package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cephtower/scripts/aliyun-ceph-lab/internal/config"
	"cephtower/scripts/aliyun-ceph-lab/internal/lab"
	"cephtower/scripts/aliyun-ceph-lab/internal/logging"
	"cephtower/scripts/aliyun-ceph-lab/internal/state"
)

const usageText = `aliyun-ceph-lab creates and initializes a temporary ECS lab from YAML.

Usage:
  aliyun-ceph-lab [command] [options]

Commands:
  validate  validate the YAML configuration without accessing cloud resources
  create    create and initialize every ECS node declared in the configuration
  list      query and print resources recorded in the local state
  delete    release recorded ECS resources and apply the network cleanup policy

Run "aliyun-ceph-lab <command> --help" for command-specific options.

Credentials are read from credentials.access_key_id, credentials.access_key_secret,
and optionally credentials.security_token in the YAML configuration.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		logging.Infof("command failed: %v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(usageText)
		return nil
	}
	if args[0] == "help" {
		if len(args) == 1 {
			fmt.Print(usageText)
			return nil
		}
		if len(args) == 2 && validCommand(args[1]) {
			fmt.Print(commandUsage(args[1]))
			return nil
		}
		return fmt.Errorf("usage: aliyun-ceph-lab help [command]")
	}
	command := args[0]
	if !validCommand(command) {
		return fmt.Errorf("unknown command %q\n\n%s", command, usageText)
	}

	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(os.Stdout)
	flags.Usage = func() {
		fmt.Fprint(flags.Output(), commandUsage(command))
		flags.PrintDefaults()
	}
	helpShort := flags.Bool("h", false, "show command help")
	helpLong := flags.Bool("help", false, "show command help")
	configPath := flags.String("config", "config.yaml", "path to YAML configuration")
	var statePath *string
	if command != "validate" {
		statePath = flags.String("state", "", "state path (default: .state/<cluster>.json beside config)")
	}
	var yes *bool
	if command == "delete" {
		yes = flags.Bool("yes", false, "confirm deletion of recorded resources")
	}
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *helpShort || *helpLong {
		flags.Usage()
		return nil
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	logging.Infof("%s: command started", command)
	logging.Infof("%s: loading configuration from %s", command, *configPath)
	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	logging.Infof("%s: configuration loaded for cluster %s with %d node(s)", command, cfg.ClusterName, len(cfg.Nodes))
	if command == "validate" {
		logging.Infof("validate: checking Alibaba Cloud credentials")
		if err := cfg.ValidateCloudCredentials(); err != nil {
			return err
		}
		logging.Infof("validate: configuration is valid; cluster=%s nodes=%d max-runtime=%s", cfg.ClusterName, len(cfg.Nodes), cfg.MaxRuntime)
		return nil
	}
	logging.Infof("%s: initializing Alibaba Cloud clients", command)
	manager, err := lab.New(cfg, *statePath)
	if err != nil {
		return err
	}
	logging.Infof("%s: using state file %s", command, manager.StatePath)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	switch command {
	case "create":
		logging.Infof("create: starting resource creation")
		if err := manager.Create(ctx); err != nil {
			return err
		}
		current, err := manager.List()
		if err != nil {
			return err
		}
		logging.Infof("create: lab is ready; automatic release=%s state=%s", current.ExpiresAt.Format(time.RFC3339), manager.StatePath)
		return printJSON(newCreateOutput(current))
	case "list":
		logging.Infof("list: querying recorded resources")
		current, err := manager.List()
		if err != nil {
			return err
		}
		logging.Infof("list: query completed with %d node(s)", len(current.Nodes))
		return printJSON(current)
	case "delete":
		current, err := state.Load(manager.StatePath)
		if err != nil {
			return err
		}
		if err := printJSON(newDeletePreview(current)); err != nil {
			return err
		}
		if !*yes {
			confirmed, err := confirmDelete(os.Stdin, os.Stdout)
			if err != nil {
				return err
			}
			if !confirmed {
				logging.Infof("delete: cancelled by user; no resources were deleted")
				return nil
			}
		}
		logging.Infof("delete: starting recorded resource deletion")
		if err := manager.Delete(ctx); err != nil {
			return err
		}
		logging.Infof("delete: recorded ECS instances released, network cleanup policy applied, and local state removed")
		return nil
	default:
		return errors.New("unreachable command")
	}
}

func validCommand(command string) bool {
	switch command {
	case "validate", "create", "list", "delete":
		return true
	default:
		return false
	}
}

func commandUsage(command string) string {
	switch command {
	case "validate":
		return "Usage: aliyun-ceph-lab validate [--config path]\n\nOptions:\n"
	case "create":
		return "Usage: aliyun-ceph-lab create [--config path] [--state path]\n\nOptions:\n"
	case "list":
		return "Usage: aliyun-ceph-lab list [--config path] [--state path]\n\nOptions:\n"
	case "delete":
		return "Usage: aliyun-ceph-lab delete [--yes] [--config path] [--state path]\n\nOptions:\n"
	default:
		return usageText
	}
}

type createOutput struct {
	Version     int              `json:"version"`
	ClusterName string           `json:"cluster_name"`
	RegionID    string           `json:"region_id"`
	CreatedAt   time.Time        `json:"created_at"`
	ExpiresAt   time.Time        `json:"expires_at"`
	Network     state.Network    `json:"network,omitempty"`
	Nodes       []createNodeInfo `json:"nodes"`
}

type createNodeInfo struct {
	Name                 string `json:"name"`
	InstanceID           string `json:"instance_id"`
	Status               string `json:"status,omitempty"`
	PublicIP             string `json:"public_ip,omitempty"`
	PrivateIP            string `json:"private_ip,omitempty"`
	SSHUser              string `json:"ssh_user"`
	SSHPassword          string `json:"ssh_password"`
	SSHPort              int    `json:"ssh_port"`
	SSHPasswordGenerated bool   `json:"ssh_password_generated"`
	LogPath              string `json:"log_path,omitempty"`
}

type deletePreview struct {
	ClusterName string           `json:"cluster_name"`
	RegionID    string           `json:"region_id"`
	Nodes       []deleteNodeInfo `json:"nodes"`
}

type deleteNodeInfo struct {
	Name       string `json:"name"`
	InstanceID string `json:"instance_id"`
	Status     string `json:"status,omitempty"`
	PublicIP   string `json:"public_ip,omitempty"`
	PrivateIP  string `json:"private_ip,omitempty"`
}

func newCreateOutput(current *state.State) createOutput {
	nodes := make([]createNodeInfo, 0, len(current.Nodes))
	for _, node := range current.Nodes {
		nodes = append(nodes, createNodeInfo{
			Name: node.Name, InstanceID: node.InstanceID, Status: node.Status,
			PublicIP: node.PublicIP, PrivateIP: node.PrivateIP,
			SSHUser: node.SSH.User, SSHPassword: node.SSH.Password, SSHPort: node.SSH.Port,
			SSHPasswordGenerated: node.SSH.PasswordGenerated, LogPath: node.SSH.LogPath,
		})
	}
	return createOutput{
		Version: current.Version, ClusterName: current.ClusterName, RegionID: current.RegionID,
		CreatedAt: current.CreatedAt, ExpiresAt: current.ExpiresAt, Network: current.Network,
		Nodes: nodes,
	}
}

func newDeletePreview(current *state.State) deletePreview {
	nodes := make([]deleteNodeInfo, 0, len(current.Nodes))
	for _, node := range current.Nodes {
		nodes = append(nodes, deleteNodeInfo{
			Name: node.Name, InstanceID: node.InstanceID, Status: node.Status,
			PublicIP: node.PublicIP, PrivateIP: node.PrivateIP,
		})
	}
	return deletePreview{
		ClusterName: current.ClusterName,
		RegionID:    current.RegionID,
		Nodes:       nodes,
	}
}

func confirmDelete(reader io.Reader, writer io.Writer) (bool, error) {
	scanner := bufio.NewScanner(reader)
	for {
		fmt.Fprintln(writer, "Delete the machines listed above? Enter y/yes or n/no:")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return false, fmt.Errorf("read deletion confirmation: %w", err)
			}
			return false, errors.New("read deletion confirmation: input closed")
		}
		switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(writer, "Invalid response; enter y/yes or n/no.")
		}
	}
}

func printJSON(value any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
