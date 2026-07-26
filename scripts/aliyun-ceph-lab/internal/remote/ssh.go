package remote

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"cephtower/scripts/aliyun-ceph-lab/internal/logging"
)

type SSH struct {
	User           string
	Password       string
	KnownHostsPath string
	LogDir         string
	knownHostsMu   sync.Mutex
}

const scriptHeartbeatInterval = 15 * time.Second

type synchronizedWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

func (w *synchronizedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writer.Write(p)
}

func (s *SSH) Wait(ctx context.Context, host, hostname string) error {
	logFile, err := s.openHostLog(hostname)
	if err != nil {
		return err
	}
	defer logFile.Close()
	writeHostLogHeader(logFile, "WAIT", "waiting for SSH connectivity")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var lastErr error
	attempt := 0
	for {
		attempt++
		var stdout, stderr bytes.Buffer
		if err := s.run(ctx, host, "true", nil, &stdout, &stderr); err == nil {
			writeHostLogHeader(logFile, "READY", fmt.Sprintf("SSH connected on attempt %d", attempt))
			logging.Infof("ssh: connection to %s succeeded on attempt %d", host, attempt)
			return nil
		} else {
			lastErr = fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
			writeHostLogHeader(logFile, "RETRY", fmt.Sprintf("attempt %d failed: %v", attempt, lastErr))
			logging.Infof("ssh: connection to %s not ready on attempt %d: %v; retrying in 5s", host, attempt, lastErr)
		}
		select {
		case <-ctx.Done():
			writeHostLogHeader(logFile, "TIMEOUT", fmt.Sprintf("SSH wait ended: %v; last error: %v", ctx.Err(), lastErr))
			return fmt.Errorf("wait for SSH on %s: %w (last error: %v)", host, ctx.Err(), lastErr)
		case <-ticker.C:
		}
	}
}

func (s *SSH) RunScript(ctx context.Context, host, hostname, scriptPath string, environment map[string]string) error {
	logging.Infof("ssh: reading script %s for host %s", scriptPath, host)
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("read script %s: %w", scriptPath, err)
	}
	keys := make([]string, 0, len(environment))
	for key := range environment {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	// Send environment values through stdin with the script instead of exposing
	// generated credentials in the remote process command line.
	assignments := make([]string, 0, len(environment))
	for _, key := range keys {
		assignments = append(assignments, "export "+key+"="+shellQuote(environment[key]))
	}
	payload := bytes.NewBuffer(nil)
	for _, assignment := range assignments {
		payload.WriteString(assignment)
		payload.WriteByte('\n')
	}
	payload.Write(script)
	scriptName := filepath.Base(scriptPath)
	remoteLogPath := filepath.Join("/var/log/ceph-lab", safeHostFilename(scriptName)+".log")
	runID, err := newRunID()
	if err != nil {
		return fmt.Errorf("generate remote script run ID: %w", err)
	}
	remoteStatusPath := filepath.Join("/var/log/ceph-lab/status", runID+".status")
	remoteCommand := remoteScriptCommand(remoteLogPath, remoteStatusPath)
	if s.User != "root" {
		remoteCommand = "sudo " + remoteCommand
	}
	logFile, err := s.openHostLog(hostname)
	if err != nil {
		return err
	}
	defer logFile.Close()
	writeHostLogHeader(logFile, "START", fmt.Sprintf("script=%s remote_log=%s", scriptName, remoteLogPath))
	output := &synchronizedWriter{writer: io.MultiWriter(os.Stdout, logFile)}
	logging.Infof("ssh: starting %s on %s; remote log=%s", scriptName, host, remoteLogPath)
	if err := s.runTrackedScript(
		ctx, host, remoteCommand, payload, output, remoteStatusPath, scriptName,
	); err != nil {
		writeHostLogHeader(logFile, "FAILED", fmt.Sprintf("script=%s error=%v remote_log=%s", scriptName, err, remoteLogPath))
		logging.Infof("ssh: %s failed on %s: %v; remote log=%s", scriptName, host, err, remoteLogPath)
		return fmt.Errorf("run %s on %s: %w (remote log: %s)", scriptName, host, err, remoteLogPath)
	}
	writeHostLogHeader(logFile, "DONE", fmt.Sprintf("script=%s remote_log=%s", scriptName, remoteLogPath))
	logging.Infof("ssh: completed %s on %s; remote log=%s", scriptName, host, remoteLogPath)
	return nil
}

func newRunID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func remoteScriptCommand(logPath, statusPath string) string {
	quotedLogPath := shellQuote(logPath)
	quotedStatusPath := shellQuote(statusPath)
	return "install -d -m 0755 /var/log/ceph-lab /var/log/ceph-lab/status && " +
		"touch " + quotedLogPath + " && chmod 0600 " + quotedLogPath + " && " +
		"rm -f " + quotedStatusPath + " " + quotedStatusPath + ".tmp && " +
		"script_path=$(mktemp /tmp/ceph-lab-hook.XXXXXX) && " +
		"trap 'rm -f \"$script_path\"' EXIT && " +
		"cat >\"$script_path\" && chmod 0700 \"$script_path\" && " +
		"bash -o pipefail -c " + shellQuote(
		"{ printf '\\n[%s] [START] remote hook\\n' \"$(date --iso-8601=seconds)\"; "+
			"bash \"$1\"; rc=$?; printf '[%s] [EXIT] status=%s\\n' \"$(date --iso-8601=seconds)\" \"$rc\"; "+
			"printf '%s\\n' \"$rc\" >\"$2.tmp\"; chmod 0600 \"$2.tmp\"; mv -f \"$2.tmp\" \"$2\"; "+
			"exit \"$rc\"; } "+
			"2>&1 | tee -a "+quotedLogPath,
	) + " _ \"$script_path\" " + quotedStatusPath
}

func (s *SSH) runTrackedScript(
	ctx context.Context,
	host string,
	command string,
	stdin io.Reader,
	output io.Writer,
	statusPath string,
	scriptName string,
) error {
	client, err := s.dial(ctx, host)
	if err != nil {
		return err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()
	session.Stdin = stdin
	session.Stdout = output
	session.Stderr = output
	if err := session.Start(command); err != nil {
		return fmt.Errorf("start SSH command: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- session.Wait() }()
	started := time.Now()
	ticker := time.NewTicker(scriptHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case waitErr := <-done:
			if waitErr == nil {
				return nil
			}
			return s.waitForRemoteScriptCompletion(
				ctx, host, statusPath, output, scriptName, started, waitErr,
			)
		case <-ticker.C:
			elapsed := time.Since(started).Round(time.Second)
			_, _ = fmt.Fprintf(output,
				"\n[%s] [RUNNING] script=%s elapsed=%s; checking remote completion status\n",
				time.Now().Format(time.RFC3339), scriptName, elapsed,
			)
			status, complete, probeErr := s.probeScriptStatus(ctx, host, statusPath)
			if probeErr != nil {
				_, _ = fmt.Fprintf(output,
					"[%s] [RUNNING] script=%s status probe failed: %v; continuing to wait\n",
					time.Now().Format(time.RFC3339), scriptName, probeErr,
				)
				continue
			}
			if complete {
				_ = client.Close()
				waitErr := <-done
				return recoveredScriptResult(output, scriptName, status, waitErr)
			}
			_, _ = fmt.Fprintf(output,
				"[%s] [RUNNING] script=%s remote status=pending\n",
				time.Now().Format(time.RFC3339), scriptName,
			)
		case <-ctx.Done():
			_ = client.Close()
			<-done
			return ctx.Err()
		}
	}
}

func (s *SSH) waitForRemoteScriptCompletion(
	ctx context.Context,
	host string,
	statusPath string,
	output io.Writer,
	scriptName string,
	started time.Time,
	transportErr error,
) error {
	_, _ = fmt.Fprintf(output,
		"\n[%s] [RECONNECTING] script=%s SSH transport ended: %v; polling remote run status\n",
		time.Now().Format(time.RFC3339), scriptName, transportErr,
	)
	ticker := time.NewTicker(scriptHeartbeatInterval)
	defer ticker.Stop()
	for {
		status, complete, probeErr := s.probeScriptStatus(ctx, host, statusPath)
		if probeErr != nil {
			_, _ = fmt.Fprintf(output,
				"[%s] [RECONNECTING] script=%s status probe failed: %v\n",
				time.Now().Format(time.RFC3339), scriptName, probeErr,
			)
		} else if complete {
			return recoveredScriptResult(output, scriptName, status, transportErr)
		} else {
			_, _ = fmt.Fprintf(output,
				"[%s] [RECONNECTING] script=%s remote status=pending elapsed=%s\n",
				time.Now().Format(time.RFC3339), scriptName,
				time.Since(started).Round(time.Second),
			)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for remote script after SSH transport error: %w (transport error: %v)",
				ctx.Err(), transportErr)
		case <-ticker.C:
		}
	}
}

func (s *SSH) probeScriptStatus(
	ctx context.Context,
	host string,
	statusPath string,
) (status int, complete bool, err error) {
	probeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	var stdout, stderr bytes.Buffer
	command := "if test -r " + shellQuote(statusPath) + "; then " +
		"printf 'complete:'; cat " + shellQuote(statusPath) + "; else printf 'pending'; fi"
	if err := s.run(probeCtx, host, command, nil, &stdout, &stderr); err != nil {
		return 0, false, fmt.Errorf("probe remote status: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	value := strings.TrimSpace(stdout.String())
	if value == "pending" {
		return 0, false, nil
	}
	if !strings.HasPrefix(value, "complete:") {
		return 0, false, fmt.Errorf("unexpected remote status %q", value)
	}
	status, err = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(value, "complete:")))
	if err != nil {
		return 0, false, fmt.Errorf("parse remote status %q: %w", value, err)
	}
	return status, true, nil
}

func recoveredScriptResult(output io.Writer, scriptName string, status int, waitErr error) error {
	if status != 0 {
		return fmt.Errorf("remote script exited with status %d after SSH transport error: %v", status, waitErr)
	}
	_, _ = fmt.Fprintf(output,
		"[%s] [RECOVERED] script=%s completed remotely with status=0 after SSH transport interruption\n",
		time.Now().Format(time.RFC3339), scriptName,
	)
	return nil
}

func (s *SSH) HostLogPath(hostname string) string {
	return filepath.Join(s.LogDir, safeHostFilename(hostname)+".log")
}

func (s *SSH) openHostLog(hostname string) (*os.File, error) {
	if s.LogDir == "" {
		return nil, errors.New("SSH log directory is not configured")
	}
	if err := os.MkdirAll(s.LogDir, 0o700); err != nil {
		return nil, fmt.Errorf("create SSH log directory: %w", err)
	}
	if err := os.Chmod(s.LogDir, 0o700); err != nil {
		return nil, fmt.Errorf("secure SSH log directory: %w", err)
	}
	path := s.HostLogPath(hostname)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open SSH host log %s: %w", path, err)
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return nil, fmt.Errorf("secure SSH host log %s: %w", path, err)
	}
	return file, nil
}

func safeHostFilename(host string) string {
	var value strings.Builder
	for _, char := range host {
		switch {
		case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z',
			char >= '0' && char <= '9', char == '.', char == '-', char == '_':
			value.WriteRune(char)
		default:
			value.WriteByte('_')
		}
	}
	if value.Len() == 0 {
		return "unknown-host"
	}
	return value.String()
}

func writeHostLogHeader(file *os.File, phase, message string) {
	_, _ = fmt.Fprintf(file, "\n[%s] [%s] %s\n", time.Now().Format(time.RFC3339), phase, message)
	_ = file.Sync()
}

func (s *SSH) run(
	ctx context.Context,
	host string,
	command string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) error {
	client, err := s.dial(ctx, host)
	if err != nil {
		return err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()
	session.Stdin = stdin
	session.Stdout = stdout
	session.Stderr = stderr
	if err := session.Start(command); err != nil {
		return fmt.Errorf("start SSH command: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- session.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("SSH command failed: %w", err)
		}
		return nil
	case <-ctx.Done():
		_ = client.Close()
		<-done
		return ctx.Err()
	}
}

func (s *SSH) dial(ctx context.Context, host string) (*ssh.Client, error) {
	address := net.JoinHostPort(host, "22")
	connection, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", address, err)
	}
	deadline := time.Now().Add(10 * time.Second)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if err := connection.SetDeadline(deadline); err != nil {
		connection.Close()
		return nil, fmt.Errorf("set SSH handshake deadline: %w", err)
	}
	configuration := &ssh.ClientConfig{
		User:            s.User,
		Auth:            []ssh.AuthMethod{ssh.Password(s.Password)},
		HostKeyCallback: s.verifyHostKey,
		Timeout:         10 * time.Second,
	}
	clientConnection, channels, requests, err := ssh.NewClientConn(connection, address, configuration)
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("authenticate to %s with password: %w", address, err)
	}
	if err := connection.SetDeadline(time.Time{}); err != nil {
		clientConnection.Close()
		return nil, fmt.Errorf("clear SSH connection deadline: %w", err)
	}
	return ssh.NewClient(clientConnection, channels, requests), nil
}

func (s *SSH) verifyHostKey(hostname string, remote net.Addr, key ssh.PublicKey) error {
	s.knownHostsMu.Lock()
	defer s.knownHostsMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(s.KnownHostsPath), 0o700); err != nil {
		return fmt.Errorf("create known_hosts directory: %w", err)
	}
	file, err := os.OpenFile(s.KnownHostsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open known_hosts: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close known_hosts: %w", err)
	}
	callback, err := knownhosts.New(s.KnownHostsPath)
	if err != nil {
		return fmt.Errorf("load known_hosts: %w", err)
	}
	if err := callback(hostname, remote, key); err == nil {
		return nil
	} else {
		var keyError *knownhosts.KeyError
		if !errors.As(err, &keyError) || len(keyError.Want) != 0 {
			return fmt.Errorf("verify SSH host key: %w", err)
		}
	}
	line := knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key) + "\n"
	file, err = os.OpenFile(s.KnownHostsPath, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("append known_hosts: %w", err)
	}
	if _, err := io.WriteString(file, line); err != nil {
		file.Close()
		return fmt.Errorf("write known_hosts: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close known_hosts: %w", err)
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
