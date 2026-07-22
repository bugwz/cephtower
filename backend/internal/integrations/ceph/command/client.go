package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

var ErrCommandNotConfigured = errors.New("ceph command binary is not configured")

const defaultTimeout = 15 * time.Second

type Config struct {
	Bin     string
	Cluster string
	Conf    string
	MonHost string
	Name    string
	Keyring string
	Timeout time.Duration
}

type CommandClient struct {
	bin     string
	cluster string
	conf    string
	monHost string
	name    string
	keyring string
	timeout time.Duration
	run     CommandRunner
}

type CommandRequest struct {
	Args   []string
	Format string
	Input  []byte
}

type CommandResult struct {
	Bin      string
	Args     []string
	Stdout   []byte
	Stderr   []byte
	ExitCode int
}

type CommandError struct {
	Result CommandResult
}

type CommandRunner func(ctx context.Context, bin string, args []string, input []byte) (CommandResult, error)

func (e *CommandError) Error() string {
	message := strings.TrimSpace(string(e.Result.Stderr))
	if message == "" {
		message = strings.TrimSpace(string(e.Result.Stdout))
	}
	if message == "" {
		message = "command exited with non-zero status"
	}

	return fmt.Sprintf("ceph command %q failed with exit code %d: %s",
		strings.Join(append([]string{e.Result.Bin}, e.Result.Args...), " "),
		e.Result.ExitCode,
		message,
	)
}

func NewCommandClient(cfg Config) *CommandClient {
	bin := strings.TrimSpace(cfg.Bin)
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	return &CommandClient{
		bin:     bin,
		cluster: strings.TrimSpace(cfg.Cluster),
		conf:    strings.TrimSpace(cfg.Conf),
		monHost: strings.TrimSpace(cfg.MonHost),
		name:    strings.TrimSpace(cfg.Name),
		keyring: strings.TrimSpace(cfg.Keyring),
		timeout: timeout,
		run:     execCephCommand,
	}
}

func NewCommandClientWithRunner(cfg Config, runner CommandRunner) *CommandClient {
	c := NewCommandClient(cfg)
	if runner != nil {
		c.run = runner
	}
	return c
}

func (c *CommandClient) Run(ctx context.Context, request CommandRequest) (CommandResult, error) {
	if strings.TrimSpace(c.bin) == "" {
		return CommandResult{}, ErrCommandNotConfigured
	}

	args := c.commandArgs(request)
	runCtx := ctx
	cancel := func() {}
	if c.timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, c.timeout)
	}
	defer cancel()

	result, err := c.run(runCtx, c.bin, args, request.Input)
	if result.Bin == "" {
		result.Bin = c.bin
	}
	if result.Args == nil {
		result.Args = append([]string(nil), args...)
	}
	if err != nil {
		return result, err
	}
	if result.ExitCode != 0 {
		return result, &CommandError{Result: result}
	}

	return result, nil
}

func (c *CommandClient) JSON(ctx context.Context, request CommandRequest, out any) error {
	if strings.TrimSpace(request.Format) == "" {
		request.Format = "json"
	}

	result, err := c.Run(ctx, request)
	if err != nil {
		return err
	}
	if out == nil || len(bytes.TrimSpace(result.Stdout)) == 0 {
		return nil
	}

	return json.Unmarshal(result.Stdout, out)
}

func (c *CommandClient) RawJSON(ctx context.Context, args ...string) (json.RawMessage, error) {
	var payload json.RawMessage
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) JSONCommand(ctx context.Context, args ...string) (map[string]any, error) {
	return c.jsonObject(ctx, args...)
}

func (c *CommandClient) JSONListCommand(ctx context.Context, args ...string) ([]map[string]any, error) {
	var payload []map[string]any
	err := c.JSON(ctx, CommandRequest{Args: args}, &payload)
	return payload, err
}

func (c *CommandClient) Text(ctx context.Context, args ...string) (string, error) {
	result, err := c.Run(ctx, CommandRequest{Args: args})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(result.Stdout)), nil
}

func (c *CommandClient) Exec(ctx context.Context, args ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: args})
}

func (c *CommandClient) ExecWithInput(ctx context.Context, input []byte, args ...string) (CommandResult, error) {
	return c.Run(ctx, CommandRequest{Args: args, Input: input})
}

func (c *CommandClient) commandArgs(request CommandRequest) []string {
	args := make([]string, 0, 10+len(request.Args))
	if c.cluster != "" {
		args = append(args, "--cluster", c.cluster)
	}
	if c.conf != "" {
		args = append(args, "--conf", c.conf)
	}
	if c.monHost != "" {
		args = append(args, "--mon-host", c.monHost)
	}
	if c.name != "" {
		args = append(args, "--name", c.name)
	}
	if c.keyring != "" {
		args = append(args, "--keyring", c.keyring)
	}

	args = append(args, request.Args...)
	if request.Format != "" && !hasFormatArg(args) {
		args = append(args, "--format", request.Format)
	}

	return args
}

func hasFormatArg(args []string) bool {
	for i, arg := range args {
		if arg == "--format" {
			return true
		}
		if strings.HasPrefix(arg, "--format=") {
			return true
		}
		if i > 0 && args[i-1] == "-f" {
			return true
		}
		if arg == "-f" {
			return true
		}
	}
	return false
}

func execCephCommand(ctx context.Context, bin string, args []string, input []byte) (CommandResult, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := CommandResult{
		Bin:    bin,
		Args:   append([]string(nil), args...),
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}

	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, &CommandError{Result: result}
	}

	if ctx.Err() != nil {
		return result, ctx.Err()
	}

	return result, err
}
