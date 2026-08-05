package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type labState struct {
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

type discoveredNode struct {
	labListNode
	ClusterName string
	StatePath   string
}

func resolveTarget(opts options, stdin io.Reader, stdout io.Writer) (targetConfig, error) {
	if opts.Target.Host != "" {
		return opts.Target, validateTarget(opts.Target)
	}
	if opts.Command != commandRun && !opts.ConfigExplicit {
		state, err := loadWorkspaceState(opts.StatePath)
		if err == nil {
			return state.Target, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return targetConfig{}, err
		}
	}
	target, err := chooseLabTarget(opts, stdin, stdout)
	if err != nil {
		return targetConfig{}, err
	}
	return target, validateTarget(target)
}

func chooseLabTarget(opts options, stdin io.Reader, stdout io.Writer) (targetConfig, error) {
	nodes, err := discoverLabNodes(opts.RepoRoot)
	if err != nil {
		return targetConfig{}, err
	}
	if len(nodes) == 0 {
		return targetConfig{}, errors.New("no machines found; configure host or create the aliyun lab first")
	}
	fmt.Fprintln(stdout, "Available machines:")
	for i, node := range nodes {
		fmt.Fprintf(stdout, "  %d. %s cluster=%s status=%s public=%s private=%s user=%s source=%s\n",
			i+1, node.Name, emptyAsDash(node.ClusterName), emptyAsDash(node.Status),
			emptyAsDash(node.PublicIP), emptyAsDash(node.PrivateIP),
			emptyAsDash(node.SSH.User), filepath.Base(node.StatePath))
	}
	fmt.Fprint(stdout, "Select a machine by number: ")
	var value string
	if _, err := fmt.Fscanln(stdin, &value); err != nil {
		return targetConfig{}, fmt.Errorf("read machine selection: %w", err)
	}
	index, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || index < 1 || index > len(nodes) {
		return targetConfig{}, fmt.Errorf("invalid machine selection %q", value)
	}
	node := nodes[index-1]
	host := node.SSH.Host
	if host == "" {
		host = node.PublicIP
	}
	if host == "" {
		host = node.PrivateIP
	}
	if host == "" {
		return targetConfig{}, fmt.Errorf("selected machine %s has no SSH host", node.Name)
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

func discoverLabNodes(repoRoot string) ([]discoveredNode, error) {
	stateFiles, err := findLabStateFiles(repoRoot)
	if err != nil {
		return nil, err
	}
	if len(stateFiles) == 0 {
		return nil, fmt.Errorf("no aliyun lab state JSON files found under %s",
			filepath.Join(repoRoot, "tools", "aliyun-ceph-lab", ".state"))
	}
	var nodes []discoveredNode
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
		for _, node := range state.Nodes {
			nodes = append(nodes, discoveredNode{
				labListNode: node,
				ClusterName: state.ClusterName,
				StatePath:   path,
			})
		}
	}
	if len(nodes) == 0 && len(decodeErrors) > 0 {
		return nil, fmt.Errorf("no usable aliyun lab states: %s", strings.Join(decodeErrors, "; "))
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].ClusterName != nodes[j].ClusterName {
			return nodes[i].ClusterName < nodes[j].ClusterName
		}
		return nodes[i].Name < nodes[j].Name
	})
	return nodes, nil
}

func findLabStateFiles(repoRoot string) ([]string, error) {
	toolsDir := filepath.Join(repoRoot, "tools")
	candidates := []string{filepath.Join(toolsDir, "aliyun-ceph-lab", ".state")}
	seenCandidates := map[string]bool{candidates[0]: true}
	if err := filepath.WalkDir(toolsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
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
		if err == nil && strings.Contains(strings.ToLower(relative), "aliyun-ceph-lab") && !seenCandidates[path] {
			seenCandidates[path] = true
			candidates = append(candidates, path)
		}
		return filepath.SkipDir
	}); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("scan tools directory for lab state: %w", err)
	}

	var files []string
	for _, dir := range candidates {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("read lab state directory %s: %w", dir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
				files = append(files, filepath.Join(dir, entry.Name()))
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

func emptyAsDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
