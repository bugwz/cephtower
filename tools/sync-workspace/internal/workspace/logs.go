package workspace

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type remoteLogFile struct {
	Path     string
	Inode    string
	Size     int64
	Modified string
}

type logFileMapping struct {
	Remote string
	Local  string
}

func collectRemoteLogs(ctx context.Context, machine remote, localDir string, log *logger) {
	snapshots := map[string]remoteLogFile{}
	for {
		if ctx.Err() != nil {
			return
		}
		client, err := machine.connect()
		if err != nil {
			log.Errorf("logs: connect failed: %v", err)
			if !waitContext(ctx, 3*time.Second) {
				return
			}
			continue
		}
		log.Infof("logs: mirroring remote files into %s", localDir)
		err = pollLogsWithClient(ctx, client, localDir, snapshots, log)
		client.close()
		if ctx.Err() != nil {
			return
		}
		log.Errorf("logs: connection interrupted: %v", err)
		if !waitContext(ctx, 2*time.Second) {
			return
		}
	}
}

func pollLogsWithClient(ctx context.Context, client *remoteSession, localDir string, snapshots map[string]remoteLogFile, log *logger) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if err := synchronizeLogsOnce(ctx, client, localDir, snapshots, log); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func synchronizeLogsOnce(ctx context.Context, client *remoteSession, localDir string, snapshots map[string]remoteLogFile, log *logger) error {
	files, err := client.listLogFiles(ctx)
	if err != nil {
		return err
	}
	remoteFiles := make(map[string]remoteLogFile, len(files))
	for _, file := range files {
		remoteFiles[file.Path] = file
	}
	for _, mapping := range synchronizedLogFiles() {
		localPath := filepath.Join(localDir, mapping.Local)
		remoteFile, exists := remoteFiles[mapping.Remote]
		if !exists {
			removed, err := removeLocalLog(localPath)
			if err != nil {
				return err
			}
			delete(snapshots, mapping.Remote)
			if removed {
				log.Infof("logs: removed local file missing on remote %s", localPath)
			}
			continue
		}
		previous, known := snapshots[mapping.Remote]
		if known && previous == remoteFile && localFileMatchesSize(localPath, remoteFile.Size) {
			continue
		}
		offset := localLogUpdateOffset(localPath, previous, remoteFile, known)
		if err := updateLocalLog(localPath, offset, remoteFile.Size, func(output io.Writer) error {
			length := remoteFile.Size - offset
			if length == 0 {
				return nil
			}
			return client.readLogRange(ctx, remoteFile.Path, offset, length, output)
		}); err != nil {
			return err
		}
		snapshots[mapping.Remote] = remoteFile
		if offset > 0 {
			log.Infof("logs: incrementally mirrored %s -> %s bytes=%d", mapping.Remote, localPath, remoteFile.Size-offset)
		} else {
			log.Infof("logs: fully mirrored %s -> %s bytes=%d", mapping.Remote, localPath, remoteFile.Size)
		}
	}
	return nil
}

func (r *remoteSession) listLogFiles(ctx context.Context) ([]remoteLogFile, error) {
	command := "if test -d " + shellQuote(remoteLogDir) + "; then find " + shellQuote(remoteLogDir) +
		" -maxdepth 1 -type f \\( -name 'std.log' -o -name 'cephtower.log' \\) -printf '%i\\t%s\\t%T@\\t%p\\n'; fi"
	output, err := r.run(ctx, command, nil)
	if err != nil {
		return nil, err
	}
	var files []remoteLogFile
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) != 4 {
			return nil, fmt.Errorf("unexpected remote log metadata %q", line)
		}
		size, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || size < 0 {
			return nil, fmt.Errorf("invalid remote log size %q", parts[1])
		}
		files = append(files, remoteLogFile{Path: parts[3], Inode: parts[0], Size: size, Modified: parts[2]})
	}
	return files, nil
}

func (r *remoteSession) readLogRange(ctx context.Context, path string, offset, length int64, output io.Writer) error {
	if _, ok := localLogPath(path); !ok {
		return fmt.Errorf("refusing to read unexpected remote log %q", path)
	}
	if offset < 0 || length < 0 {
		return errorsNewInvalidLogRange(offset, length)
	}
	command := "tail -c +" + strconv.FormatInt(offset+1, 10) + " -- " + shellQuote(path) +
		" | head -c " + strconv.FormatInt(length, 10)
	_, err := r.runWithOutput(ctx, command, nil, output, io.Discard)
	return err
}

func errorsNewInvalidLogRange(offset, length int64) error {
	return fmt.Errorf("invalid log range offset=%d length=%d", offset, length)
}

func localLogPath(path string) (string, bool) {
	for _, mapping := range synchronizedLogFiles() {
		if path == mapping.Remote {
			return mapping.Local, true
		}
	}
	return "", false
}

func synchronizedLogFiles() []logFileMapping {
	return []logFileMapping{
		{Remote: remoteStdLog, Local: "std.log"},
		{Remote: remoteAppLog, Local: "cephtower.log"},
	}
}

func resetLocalLogDir(toolDir string) (string, error) {
	localDir := filepath.Join(filepath.Clean(toolDir), ".state", "logs")
	if filepath.Dir(filepath.Dir(localDir)) != filepath.Clean(toolDir) {
		return "", fmt.Errorf("refusing to reset unexpected local log directory %s", localDir)
	}
	if err := os.RemoveAll(localDir); err != nil {
		return "", fmt.Errorf("reset local log directory: %w", err)
	}
	if err := os.MkdirAll(localDir, 0o700); err != nil {
		return "", fmt.Errorf("create local log directory: %w", err)
	}
	return localDir, nil
}

func removeLocalLog(path string) (bool, error) {
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("remove local log %s: %w", path, err)
	}
	return true, nil
}

func localFileMatchesSize(path string, size int64) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular() && info.Size() == size
}

func localLogUpdateOffset(path string, previous, current remoteLogFile, previousKnown bool) int64 {
	if previousKnown && previous.Inode == current.Inode && current.Size > previous.Size && localFileMatchesSize(path, previous.Size) {
		return previous.Size
	}
	return 0
}

func replaceLocalLog(path string, expectedSize int64, write func(io.Writer) error) error {
	return updateLocalLog(path, 0, expectedSize, write)
}

func updateLocalLog(path string, preservedSize, expectedSize int64, write func(io.Writer) error) error {
	if preservedSize < 0 || expectedSize < preservedSize {
		return fmt.Errorf("invalid local log update preserved=%d expected=%d", preservedSize, expectedSize)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".log-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	counter := &countingWriter{writer: tmp}
	if preservedSize > 0 {
		existing, err := os.Open(path)
		if err != nil {
			tmp.Close()
			return err
		}
		copied, copyErr := io.CopyN(counter, existing, preservedSize)
		closeErr := existing.Close()
		if copyErr != nil {
			tmp.Close()
			return fmt.Errorf("preserve local log prefix: copied %d of %d bytes: %w", copied, preservedSize, copyErr)
		}
		if closeErr != nil {
			tmp.Close()
			return closeErr
		}
	}
	if err := write(counter); err != nil {
		tmp.Close()
		return err
	}
	if counter.count != expectedSize {
		tmp.Close()
		return fmt.Errorf("remote log changed while reading: expected %d bytes, received %d", expectedSize, counter.count)
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

type countingWriter struct {
	writer io.Writer
	count  int64
}

func (w *countingWriter) Write(payload []byte) (int, error) {
	n, err := w.writer.Write(payload)
	w.count += int64(n)
	return n, err
}

func waitContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
