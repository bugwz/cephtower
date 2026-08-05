package workspace

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var knownHostsMu sync.Mutex

type remote struct {
	target targetConfig
}

type remoteSession struct {
	target targetConfig
	client *ssh.Client
}

type synchronizationResult struct {
	Uploaded []string
	Deleted  []string
}

func (r remote) connect() (*remoteSession, error) {
	if err := prepareKnownHosts(r.target.KnownHosts); err != nil {
		return nil, err
	}
	callback, err := knownhosts.New(r.target.KnownHosts)
	if err != nil {
		return nil, fmt.Errorf("load known_hosts: %w", err)
	}
	address := net.JoinHostPort(r.target.Host, strconv.Itoa(r.target.Port))
	client, err := ssh.Dial("tcp", address, &ssh.ClientConfig{
		User:            r.target.User,
		Auth:            []ssh.AuthMethod{ssh.Password(r.target.Password)},
		HostKeyCallback: learnHostKey(r.target.KnownHosts, callback),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", address, err)
	}
	return &remoteSession{target: r.target, client: client}, nil
}

func prepareKnownHosts(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create known_hosts directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open known_hosts: %w", err)
	}
	return file.Close()
}

func learnHostKey(path string, callback ssh.HostKeyCallback) ssh.HostKeyCallback {
	return func(hostname string, remoteAddress net.Addr, key ssh.PublicKey) error {
		knownHostsMu.Lock()
		defer knownHostsMu.Unlock()
		if err := callback(hostname, remoteAddress, key); err == nil {
			return nil
		} else {
			var keyError *knownhosts.KeyError
			if !errors.As(err, &keyError) {
				return fmt.Errorf("verify SSH host key: %w", err)
			}
			if len(keyError.Want) != 0 {
				if err := removeKnownHostLines(path, keyError.Want); err != nil {
					return err
				}
			}
		}
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("append known_hosts: %w", err)
		}
		defer file.Close()
		_, err = io.WriteString(file, knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key)+"\n")
		return err
	}
}

func removeKnownHostLines(path string, keys []knownhosts.KnownKey) error {
	remove := map[int]bool{}
	for _, key := range keys {
		if key.Filename == path {
			remove[key.Line] = true
		}
	}
	if len(remove) == 0 {
		return errors.New("known_hosts key mismatch")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.SplitAfter(string(raw), "\n")
	var output strings.Builder
	for index, line := range lines {
		if line != "" && !remove[index+1] {
			output.WriteString(line)
		}
	}
	return os.WriteFile(path, []byte(output.String()), 0o600)
}

func (r remote) run(ctx context.Context, command string) (string, error) {
	client, err := r.connect()
	if err != nil {
		return "", err
	}
	defer client.close()
	return client.run(ctx, command, nil)
}

func (r remote) uploadContent(ctx context.Context, content []byte, path string, mode os.FileMode) error {
	client, err := r.connect()
	if err != nil {
		return err
	}
	defer client.close()
	return client.uploadContent(ctx, content, path, mode)
}

func (r remote) synchronize(ctx context.Context, repoRoot string, current sourceManifest) (synchronizationResult, error) {
	client, err := r.connect()
	if err != nil {
		return synchronizationResult{}, err
	}
	defer client.close()
	if _, err := client.run(ctx, "install -d -m 0755 "+shellQuote(remoteRoot)+" "+shellQuote(remoteStateDir), nil); err != nil {
		return synchronizationResult{}, err
	}
	raw, err := client.run(ctx, "if test -r "+shellQuote(remoteManifest)+"; then cat "+shellQuote(remoteManifest)+"; fi", nil)
	if err != nil {
		return synchronizationResult{}, err
	}
	previous, err := decodeManifest([]byte(raw))
	if err != nil {
		return synchronizationResult{}, fmt.Errorf("decode remote source manifest: %w", err)
	}
	changed, deleted := manifestChanges(previous, current)
	result := synchronizationResult{Uploaded: changed, Deleted: deleted}
	if len(changed) == 0 && len(deleted) == 0 {
		return result, nil
	}
	if err := client.removeSourcePaths(ctx, deleted); err != nil {
		return synchronizationResult{}, err
	}
	if len(changed) > 0 {
		if err := client.uploadArchive(ctx, repoRoot, changed); err != nil {
			return synchronizationResult{}, err
		}
	}
	payload, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return synchronizationResult{}, err
	}
	payload = append(payload, '\n')
	if err := client.uploadContent(ctx, payload, remoteManifest, 0o600); err != nil {
		return synchronizationResult{}, err
	}
	return result, nil
}

func (r *remoteSession) close() {
	_ = r.client.Close()
}

func (r *remoteSession) run(ctx context.Context, command string, input io.Reader) (string, error) {
	var stdout, stderr bytes.Buffer
	_, err := r.runWithOutput(ctx, command, input, &stdout, &stderr)
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return stdout.String(), fmt.Errorf("%w: %s", err, message)
		}
		return stdout.String(), err
	}
	return stdout.String(), nil
}

func (r *remoteSession) runWithOutput(ctx context.Context, command string, input io.Reader, stdout, stderr io.Writer) (string, error) {
	session, err := r.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create SSH session: %w", err)
	}
	defer session.Close()
	session.Stdin = input
	session.Stdout = stdout
	session.Stderr = stderr
	if r.target.User != "root" {
		command = "sudo -n sh -c " + shellQuote(command)
	}
	done := make(chan error, 1)
	go func() { done <- session.Run(command) }()
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("remote command failed: %w", err)
		}
		return "", nil
	case <-ctx.Done():
		_ = session.Close()
		<-done
		return "", ctx.Err()
	}
}

func (r *remoteSession) uploadContent(ctx context.Context, content []byte, path string, mode os.FileMode) error {
	tmp := fmt.Sprintf("%s.tmp.%d", path, time.Now().UnixNano())
	command := "cat > " + shellQuote(tmp) +
		" && chmod " + shellQuote(fmt.Sprintf("%04o", mode.Perm())) + " " + shellQuote(tmp) +
		" && mv -f " + shellQuote(tmp) + " " + shellQuote(path)
	_, err := r.run(ctx, command, bytes.NewReader(content))
	return err
}

func (r *remoteSession) removeSourcePaths(ctx context.Context, paths []string) error {
	const chunkSize = 80
	for start := 0; start < len(paths); start += chunkSize {
		end := start + chunkSize
		if end > len(paths) {
			end = len(paths)
		}
		var command strings.Builder
		command.WriteString("cd ")
		command.WriteString(shellQuote(remoteRoot))
		command.WriteString(" && rm -rf --")
		for _, path := range paths[start:end] {
			if !shouldSyncPath(path) {
				return fmt.Errorf("refusing to remove unsafe remote source path %q", path)
			}
			command.WriteByte(' ')
			command.WriteString(shellQuote(path))
		}
		if _, err := r.run(ctx, command.String(), nil); err != nil {
			return err
		}
	}
	return nil
}

func (r *remoteSession) uploadArchive(ctx context.Context, repoRoot string, paths []string) error {
	stagingDir := remoteStateDir + "/upload-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	installCommand, err := archiveInstallCommand(paths, stagingDir)
	if err != nil {
		return err
	}
	reader, writer := io.Pipe()
	tarDone := make(chan error, 1)
	go func() {
		tarWriter := tar.NewWriter(writer)
		var archiveErr error
		for _, path := range paths {
			if archiveErr = addArchiveEntry(tarWriter, repoRoot, path); archiveErr != nil {
				break
			}
		}
		if closeErr := tarWriter.Close(); archiveErr == nil {
			archiveErr = closeErr
		}
		_ = writer.CloseWithError(archiveErr)
		tarDone <- archiveErr
	}()
	_, runErr := r.run(ctx, installCommand, reader)
	_ = reader.CloseWithError(runErr)
	tarErr := <-tarDone
	if tarErr != nil {
		return fmt.Errorf("create source archive: %w", tarErr)
	}
	return runErr
}

func archiveInstallCommand(paths []string, stagingDir string) (string, error) {
	stagingPrefix := remoteStateDir + "/upload-"
	stagingID := strings.TrimPrefix(stagingDir, stagingPrefix)
	if stagingID == stagingDir || stagingID == "" {
		return "", fmt.Errorf("unsafe remote staging directory %q", stagingDir)
	}
	if value, err := strconv.ParseInt(stagingID, 10, 64); err != nil || value < 1 {
		return "", fmt.Errorf("unsafe remote staging directory %q", stagingDir)
	}
	var command strings.Builder
	command.WriteString("install -d -m 0700 ")
	command.WriteString(shellQuote(stagingDir))
	command.WriteString(" && tar --no-same-owner -xpf - -C ")
	command.WriteString(shellQuote(stagingDir))
	for _, path := range paths {
		if !shouldSyncPath(path) {
			return "", fmt.Errorf("unsafe archive path %q", path)
		}
		sourcePath := stagingDir + "/" + path
		targetPath := remoteRoot + "/" + path
		targetDir := filepath.ToSlash(filepath.Dir(targetPath))
		command.WriteString(" && install -d -m 0755 ")
		command.WriteString(shellQuote(targetDir))
		command.WriteString(" && if test -d ")
		command.WriteString(shellQuote(targetPath))
		command.WriteString(" && ! test -L ")
		command.WriteString(shellQuote(targetPath))
		command.WriteString("; then rm -rf -- ")
		command.WriteString(shellQuote(targetPath))
		command.WriteString("; fi && mv -f -- ")
		command.WriteString(shellQuote(sourcePath))
		command.WriteByte(' ')
		command.WriteString(shellQuote(targetPath))
	}
	command.WriteString("; status=$?; rm -rf -- ")
	command.WriteString(shellQuote(stagingDir))
	command.WriteString("; exit \"$status\"")
	return command.String(), nil
}

func addArchiveEntry(writer *tar.Writer, repoRoot, path string) error {
	if !shouldSyncPath(path) {
		return fmt.Errorf("unsafe archive path %q", path)
	}
	localPath := filepath.Join(repoRoot, filepath.FromSlash(path))
	info, err := os.Lstat(localPath)
	if err != nil {
		return err
	}
	link := ""
	if info.Mode()&os.ModeSymlink != 0 {
		link, err = os.Readlink(localPath)
		if err != nil {
			return err
		}
	}
	header, err := tar.FileInfoHeader(info, link)
	if err != nil {
		return err
	}
	header.Name = path
	if err := writer.WriteHeader(header); err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	return err
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
