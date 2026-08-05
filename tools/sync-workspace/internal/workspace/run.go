package workspace

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const restartDebounce = 500 * time.Millisecond

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	toolDir := filepath.Join(repoRoot, "tools", "sync-workspace")
	opts, err := parseOptions(args, repoRoot, toolDir, stdout)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	target, err := resolveTarget(opts, stdin, stdout)
	if err != nil {
		return err
	}
	opts.Target = target
	machine := remote{target: target}
	log := newLogger(stderr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Infof("ssh: target=%s:%d user=%s", target.Host, target.Port, target.User)

	switch opts.Command {
	case commandRun:
		return runWorkspace(ctx, opts, machine, stdout, log)
	case commandStop:
		log.Infof("service: stopping remote service host=%s", target.Host)
		if err := stopRemoteService(ctx, machine); err != nil {
			return err
		}
		log.Infof("service: stopped host=%s", target.Host)
		fmt.Fprintln(stdout, "Remote development service stopped.")
		return nil
	case commandStatus:
		status, err := remoteServiceStatus(ctx, machine)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Service: %s\n", status)
		fmt.Fprintf(stdout, "Web: http://%s\n", net.JoinHostPort(target.Host, fmt.Sprint(webPort)))
		return nil
	default:
		return fmt.Errorf("unsupported command %q", opts.Command)
	}
}

func runWorkspace(ctx context.Context, opts options, machine remote, stdout io.Writer, log *logger) error {
	localLogDir, err := resetLocalLogDir(opts.ToolDir)
	if err != nil {
		return err
	}
	log.Infof("logs: reset local mirror directory %s", localLogDir)
	if err := saveWorkspaceState(opts.StatePath, workspaceState{
		Target: opts.Target,
		LogDir: localLogDir,
	}); err != nil {
		return fmt.Errorf("save workspace state: %w", err)
	}

	if opts.CleanOnStart {
		log.Infof("clean: stopping service before removing %s", remoteAppDir)
		if err := stopRemoteService(ctx, machine); err != nil {
			return err
		}
		log.Infof("service: stopped before remote runtime cleanup")
		if err := cleanRemoteApp(ctx, machine); err != nil {
			return fmt.Errorf("clean remote app directory: %w", err)
		}
		log.Infof("clean: removed remote runtime data %s", remoteAppDir)
	}

	log.Infof("sync: scanning Git workspace")
	current, err := scanGitWorkspace(ctx, opts.RepoRoot)
	if err != nil {
		return err
	}
	result, err := machine.synchronize(ctx, opts.RepoRoot, current)
	if err != nil {
		return fmt.Errorf("initial source synchronization: %w", err)
	}
	logSynchronizationResult(log, result)
	if err := writeJSONFile(opts.ManifestPath, current, 0o600); err != nil {
		return fmt.Errorf("save local manifest: %w", err)
	}
	log.Infof("sync: initial workspace ready files=%d uploaded=%d deleted=%d", len(current.Files), len(result.Uploaded), len(result.Deleted))

	logCtx, stopLogs := context.WithCancel(ctx)
	defer stopLogs()
	go collectRemoteLogs(logCtx, machine, localLogDir, log)

	if err := buildAndRestart(ctx, machine, log); err != nil {
		return err
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		log.Infof("service: stopping remote service during shutdown")
		if err := stopRemoteService(stopCtx, machine); err != nil {
			log.Errorf("service: stop during shutdown failed: %v", err)
		} else {
			log.Infof("service: stopped")
		}
	}()

	address := net.JoinHostPort(opts.Target.Host, fmt.Sprint(webPort))
	fmt.Fprintf(stdout, "Web: http://%s\n", address)
	fmt.Fprintf(stdout, "Logs: %s\n", localLogDir)
	log.Infof("watch: Git workspace changes trigger a remote rebuild and restart")

	lastSynced := current
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
		next, err := scanGitWorkspace(ctx, opts.RepoRoot)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Errorf("watch: scan failed: %v", err)
			continue
		}
		if manifestsEqual(lastSynced, next) {
			continue
		}
		if !waitContext(ctx, restartDebounce) {
			return nil
		}
		settled, err := scanGitWorkspace(ctx, opts.RepoRoot)
		if err != nil {
			log.Errorf("watch: debounce scan failed: %v", err)
			continue
		}
		result, err := machine.synchronize(ctx, opts.RepoRoot, settled)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Errorf("sync: failed: %v", err)
			continue
		}
		if err := writeJSONFile(opts.ManifestPath, settled, 0o600); err != nil {
			log.Errorf("sync: save local manifest failed: %v", err)
		}
		lastSynced = settled
		logSynchronizationResult(log, result)
		log.Infof("sync: applied uploaded=%d deleted=%d; rebuilding service", len(result.Uploaded), len(result.Deleted))
		if err := buildAndRestart(ctx, machine, log); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Errorf("service: rebuild/restart failed; previous service remains when possible: %v", err)
			continue
		}
		log.Infof("service: restarted Web=http://%s", address)
	}
}

func logSynchronizationResult(log *logger, result synchronizationResult) {
	for _, localPath := range result.Uploaded {
		log.Infof("sync: uploaded %s -> %s/%s", localPath, remoteRoot, localPath)
	}
	for _, remotePath := range result.Deleted {
		log.Infof("sync: deleted %s/%s", remoteRoot, remotePath)
	}
	if len(result.Uploaded) == 0 && len(result.Deleted) == 0 {
		log.Infof("sync: remote source workspace is already current")
	}
}

func findRepoRoot() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(workingDir, "Makefile")) && fileExists(filepath.Join(workingDir, "config", "config.yaml")) {
			return workingDir, nil
		}
		parent := filepath.Dir(workingDir)
		if parent == workingDir {
			return "", errors.New("could not locate CephTower repository root")
		}
		workingDir = parent
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
