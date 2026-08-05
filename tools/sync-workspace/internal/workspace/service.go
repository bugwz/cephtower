package workspace

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func cleanRemoteApp(ctx context.Context, machine remote) error {
	if remoteRoot != "/root/cephtower" || remoteAppDir != remoteRoot+"/app" {
		return errors.New("refusing to clean an unexpected remote app directory")
	}
	command := "test " + shellQuote(remoteAppDir) + " = " + shellQuote("/root/cephtower/app") +
		" && rm -rf -- " + shellQuote(remoteAppDir) +
		" && install -d -m 0755 " + shellQuote(remoteAppDir)
	_, err := machine.run(ctx, command)
	return err
}

func buildAndRestart(ctx context.Context, machine remote, log *logger) error {
	log.Infof("config: preparing remote runtime config path=%s Web=0.0.0.0:%d", remoteConfigPath, webPort)
	if err := prepareRemoteConfig(ctx, machine); err != nil {
		return err
	}
	log.Infof("service: building frontend and backend on remote machine")
	marker := fmt.Sprintf("\n[%s] sync-workspace: build started\n", time.Now().Format(time.RFC3339))
	command := "install -d -m 0755 " + shellQuote(remoteStateDir) + " " + shellQuote(remoteLogDir) + " " + shellQuote(remoteRuntimeDir) +
		" && printf '%s' " + shellQuote(marker) + " >> " + shellQuote(remoteStdLog) +
		" && cd " + shellQuote(remoteRoot) +
		" && make build >> " + shellQuote(remoteStdLog) + " 2>&1"
	if _, err := machine.run(ctx, command); err != nil {
		return fmt.Errorf("remote build failed; see %s: %w", remoteStdLog, err)
	}
	log.Infof("service: remote build completed binary=%s", remoteRoot+"/bin/cephtower")
	log.Infof("service: stopping previous remote service if present")
	if err := stopRemoteService(ctx, machine); err != nil {
		return err
	}
	log.Infof("service: previous remote service stopped")
	log.Infof("service: starting Web service on port %d", webPort)
	pid, err := startRemoteService(ctx, machine)
	if err != nil {
		return err
	}
	log.Infof("service: started host=%s pid=%s Web=http://%s:%d", machine.target.Host, pid, machine.target.Host, webPort)
	return nil
}

func prepareRemoteConfig(ctx context.Context, machine remote) error {
	command := "install -d -m 0755 " + shellQuote(remoteStateDir) + " " + shellQuote(remoteLogDir) + " " + shellQuote(remoteRuntimeDir) +
		" && cd " + shellQuote(remoteRoot) +
		" && make ensure-run-config >> " + shellQuote(remoteStdLog) + " 2>&1"
	if _, err := machine.run(ctx, command); err != nil {
		return fmt.Errorf("initialize remote config: %w", err)
	}
	raw, err := machine.run(ctx, "cat "+shellQuote(remoteConfigPath))
	if err != nil {
		return fmt.Errorf("read remote runtime config: %w", err)
	}
	payload, changed, err := configForRemoteWeb([]byte(raw))
	if err != nil {
		return err
	}
	if changed {
		if err := machine.uploadContent(ctx, payload, remoteConfigPath, 0o600); err != nil {
			return fmt.Errorf("write remote runtime config: %w", err)
		}
	}
	return nil
}

func configForRemoteWeb(raw []byte) ([]byte, bool, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, false, fmt.Errorf("decode remote runtime config: %w", err)
	}
	if len(root.Content) == 0 {
		return nil, false, errors.New("remote runtime config is empty")
	}
	document := root.Content[0]
	if document.Kind != yaml.MappingNode {
		return nil, false, errors.New("remote runtime config root must be a mapping")
	}
	server := yamlMappingValue(document, "server")
	if server == nil {
		server = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		appendYAMLMapping(document, "server", server)
	}
	changed := false
	changed = setYAMLScalar(server, "address", "!!str", "0.0.0.0") || changed
	changed = setYAMLScalar(server, "port", "!!int", strconv.Itoa(webPort)) || changed
	changed = setYAMLScalar(server, "dir", "!!str", remoteAppDir) || changed
	if !changed {
		return raw, false, nil
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(4)
	if err := encoder.Encode(&root); err != nil {
		return nil, false, fmt.Errorf("encode remote runtime config: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, false, err
	}
	return output.Bytes(), true, nil
}

func yamlMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1]
		}
	}
	return nil
}

func appendYAMLMapping(node *yaml.Node, key string, value *yaml.Node) {
	node.Content = append(node.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		value,
	)
}

func setYAMLScalar(node *yaml.Node, key, tag, value string) bool {
	current := yamlMappingValue(node, key)
	if current == nil {
		appendYAMLMapping(node, key, &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: value})
		return true
	}
	if current.Kind == yaml.ScalarNode && current.Tag == tag && current.Value == value {
		return false
	}
	current.Kind = yaml.ScalarNode
	current.Tag = tag
	current.Value = value
	return true
}

func startRemoteService(ctx context.Context, machine remote) (string, error) {
	output, err := machine.run(ctx, startServiceCommand())
	if err != nil {
		return "", fmt.Errorf("start remote service: %w", err)
	}
	pid := strings.TrimSpace(output)
	if pid == "" {
		return "", errors.New("start remote service: remote PID is empty")
	}
	if parsed, err := strconv.Atoi(pid); err != nil || parsed < 1 {
		return "", fmt.Errorf("start remote service: invalid remote PID %q", pid)
	}
	return pid, nil
}

func startServiceCommand() string {
	startMarker := fmt.Sprintf("\n[%s] sync-workspace: service starting on port %d\n", time.Now().Format(time.RFC3339), webPort)
	return "install -d -m 0755 " + shellQuote(remoteStateDir) + " " + shellQuote(remoteLogDir) + " " + shellQuote(remoteRuntimeDir) +
		" && cd " + shellQuote(remoteRoot) +
		" && printf '%s' " + shellQuote(startMarker) + " >> " + shellQuote(remoteStdLog) +
		" && { nohup " + shellQuote(remoteRoot+"/bin/cephtower") +
		" -config " + shellQuote(remoteConfigPath) +
		" >> " + shellQuote(remoteStdLog) + " 2>&1 < /dev/null & " +
		"pid=$!; printf '%s\n' \"$pid\" > " + shellQuote(remotePIDFile) +
		"; ready=false; for i in $(seq 1 60); do " +
		"if curl -fsS --max-time 1 " + shellQuote("http://127.0.0.1:36900/") + " >/dev/null; then ready=true; break; fi; " +
		"kill -0 \"$pid\" 2>/dev/null || break; sleep 0.25; done; " +
		"if test \"$ready\" != true; then kill \"$pid\" 2>/dev/null || true; rm -f " + shellQuote(remotePIDFile) + "; false; fi; }" +
		" && printf '%s' \"$pid\""
}

func stopRemoteService(ctx context.Context, machine remote) error {
	_, err := machine.run(ctx, stopServiceCommand())
	if err != nil {
		return fmt.Errorf("stop remote service: %w", err)
	}
	return nil
}

func stopServiceCommand() string {
	pattern := remoteRoot + "/bin/cephtower"
	return "if test -s " + shellQuote(remotePIDFile) + "; then " +
		"pid=$(cat " + shellQuote(remotePIDFile) + "); " +
		"case \"$pid\" in ''|*[!0-9]*) rm -f " + shellQuote(remotePIDFile) + ";; *) " +
		"if kill -0 \"$pid\" 2>/dev/null && tr '\\000' ' ' < /proc/\"$pid\"/cmdline | grep -F -- " + shellQuote(pattern) + " >/dev/null; then " +
		"kill \"$pid\" 2>/dev/null || true; " +
		"for i in $(seq 1 75); do kill -0 \"$pid\" 2>/dev/null || break; sleep 0.2; done; " +
		"kill -9 \"$pid\" 2>/dev/null || true; fi; rm -f " + shellQuote(remotePIDFile) + ";; esac; fi"
}

func remoteServiceStatus(ctx context.Context, machine remote) (string, error) {
	output, err := machine.run(ctx, serviceStatusCommand())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func serviceStatusCommand() string {
	pattern := remoteRoot + "/bin/cephtower"
	return "if test -s " + shellQuote(remotePIDFile) + "; then " +
		"pid=$(cat " + shellQuote(remotePIDFile) + "); " +
		"if kill -0 \"$pid\" 2>/dev/null && tr '\\000' ' ' < /proc/\"$pid\"/cmdline | grep -F -- " + shellQuote(pattern) + " >/dev/null; then " +
		"printf 'running %s' \"$pid\"; else printf 'stopped'; fi; else printf 'stopped'; fi"
}
