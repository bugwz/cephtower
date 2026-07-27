package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"cephtower/backend/internal/security"
)

type Binary string

const (
	BinaryCeph        Binary = "ceph"
	BinaryRBD         Binary = "rbd"
	BinaryRGWAdmin    Binary = "radosgw-admin"
	BinaryCephFSShell Binary = "cephfs-shell"
	DefaultMaxOutput  int64  = 4 << 20
)

type ClusterAccess struct{ MonitorAddresses, ClientUsername, ClientKey string }
type CommandSpec struct {
	ID            string
	Binary        Binary
	Args          []string
	Stdin         []byte
	Timeout       time.Duration
	MaxOutput     int64
	Mutating      bool
	SensitiveArgs map[int]struct{}
}
type CommandResult struct {
	Stdout, Stderr []byte
	ExitCode       int
	Duration       time.Duration
	BinaryVersion  string
	RedactedArgs   []string
}

type Executor interface {
	Run(context.Context, ClusterAccess, CommandSpec) (CommandResult, error)
}
type Runner struct {
	Paths    map[Binary]string
	TempRoot string
}

type Error struct {
	Kind     string
	ExitCode int
	Summary  string
}

func (e *Error) Error() string { return e.Summary }

func (r *Runner) Run(ctx context.Context, access ClusterAccess, spec CommandSpec) (CommandResult, error) {
	if spec.ID == "" {
		return CommandResult{}, fmt.Errorf("command spec ID is required")
	}
	path, ok := r.binaryPath(spec.Binary)
	if !ok {
		return CommandResult{}, fmt.Errorf("binary %q is not allowed", spec.Binary)
	}
	if spec.Timeout <= 0 {
		return CommandResult{}, fmt.Errorf("command timeout must be positive")
	}
	max := spec.MaxOutput
	if max <= 0 {
		max = DefaultMaxOutput
	}
	runCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()
	dir, conf, keyring, err := materializeAccess(r.TempRoot, access)
	if err != nil {
		return CommandResult{}, err
	}
	defer os.RemoveAll(dir)
	args := append([]string{}, spec.Args...)
	if spec.Binary == BinaryCeph || spec.Binary == BinaryRBD || spec.Binary == BinaryCephFSShell {
		args = append([]string{"--conf", conf, "--name", access.ClientUsername, "--keyring", keyring}, args...)
	}
	cmd := exec.Command(path, args...)
	configureCommandProcess(cmd)
	cmd.Stdin = bytes.NewReader(spec.Stdin)
	stdout := &limitedBuffer{remaining: max}
	stderr := &limitedBuffer{remaining: max}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	started := time.Now()
	if err := cmd.Start(); err != nil {
		return CommandResult{}, fmt.Errorf("start %s: %w", spec.ID, err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	var waitErr error
	select {
	case waitErr = <-done:
	case <-runCtx.Done():
		_ = killCommandProcess(cmd)
		waitErr = <-done
	}
	result := CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: 0, Duration: time.Since(started), RedactedArgs: redactArgs(spec, args)}
	if stdout.exceeded || stderr.exceeded {
		return result, &Error{Kind: "output_limit", ExitCode: result.ExitCode, Summary: "command output exceeded configured limit"}
	}
	if runCtx.Err() != nil {
		return result, &Error{Kind: "timeout", ExitCode: -1, Summary: "command timed out or was cancelled"}
	}
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			return result, &Error{Kind: "exit", ExitCode: result.ExitCode, Summary: security.Redact(string(result.Stderr))}
		}
		return result, waitErr
	}
	return result, nil
}

func (r *Runner) binaryPath(binary Binary) (string, bool) {
	switch binary {
	case BinaryCeph, BinaryRBD, BinaryRGWAdmin, BinaryCephFSShell:
	default:
		return "", false
	}
	if path := r.Paths[binary]; path != "" {
		return path, true
	}
	path, err := exec.LookPath(string(binary))
	return path, err == nil
}
func materializeAccess(root string, access ClusterAccess) (string, string, string, error) {
	if access.MonitorAddresses == "" || access.ClientUsername == "" || access.ClientKey == "" {
		return "", "", "", fmt.Errorf("cluster access is incomplete")
	}
	dir, err := os.MkdirTemp(root, "cephtower-ceph-")
	if err != nil {
		return "", "", "", err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		os.RemoveAll(dir)
		return "", "", "", err
	}
	conf := filepath.Join(dir, "ceph.conf")
	keyring := filepath.Join(dir, "keyring")
	confData := []byte("[global]\nmon_host = " + access.MonitorAddresses + "\n")
	keyringData := []byte("[" + access.ClientUsername + "]\n\tkey = " + access.ClientKey + "\n")
	if err := atomicWrite(conf, confData); err != nil {
		os.RemoveAll(dir)
		return "", "", "", err
	}
	if err := atomicWrite(keyring, keyringData); err != nil {
		os.RemoveAll(dir)
		return "", "", "", err
	}
	return dir, conf, keyring, nil
}
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	file, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.Remove(tmp)
		}
	}()
	if _, err = file.Write(data); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmp, path); err != nil {
		return err
	}
	ok = true
	return nil
}

type limitedBuffer struct {
	buffer    bytes.Buffer
	remaining int64
	exceeded  bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	original := len(p)
	if b.remaining <= 0 {
		b.exceeded = true
		return original, nil
	}
	write := p
	if int64(len(write)) > b.remaining {
		write = write[:b.remaining]
		b.exceeded = true
	}
	_, _ = b.buffer.Write(write)
	b.remaining -= int64(len(write))
	return original, nil
}
func (b *limitedBuffer) Bytes() []byte { return b.buffer.Bytes() }

var _ io.Writer = (*limitedBuffer)(nil)

func redactArgs(spec CommandSpec, args []string) []string {
	result := append([]string{}, args...)
	prefix := 0
	if spec.Binary == BinaryCeph || spec.Binary == BinaryRBD || spec.Binary == BinaryCephFSShell {
		prefix = 6
		result[1] = "[TEMP_CONF]"
		result[3] = specArgsValue(args, 3)
		result[5] = "[TEMP_KEYRING]"
	}
	for index := range spec.SensitiveArgs {
		target := prefix + index
		if target >= 0 && target < len(result) {
			result[target] = "[REDACTED]"
		}
	}
	return result
}
func specArgsValue(args []string, index int) string {
	if index >= 0 && index < len(args) {
		return args[index]
	}
	return strconv.Itoa(index)
}
