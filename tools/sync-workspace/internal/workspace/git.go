package workspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type sourceEntry struct {
	Hash string `json:"hash"`
	Mode uint32 `json:"mode"`
	Kind string `json:"kind"`
}

type sourceManifest struct {
	Version int                    `json:"version"`
	Files   map[string]sourceEntry `json:"files"`
}

func scanGitWorkspace(ctx context.Context, repoRoot string) (sourceManifest, error) {
	cmd := exec.CommandContext(ctx, "git", "ls-files", "-co", "--exclude-standard", "--deduplicate", "-z")
	cmd.Dir = repoRoot
	raw, err := cmd.Output()
	if err != nil {
		return sourceManifest{}, fmt.Errorf("list Git workspace files: %w", err)
	}
	manifest := sourceManifest{Version: 1, Files: map[string]sourceEntry{}}
	for _, item := range strings.Split(string(raw), "\x00") {
		path := filepath.ToSlash(item)
		if path == "" || !shouldSyncPath(path) {
			continue
		}
		entry, err := hashSourceEntry(filepath.Join(repoRoot, filepath.FromSlash(path)))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return sourceManifest{}, fmt.Errorf("inspect %s: %w", path, err)
		}
		manifest.Files[path] = entry
	}
	return manifest, nil
}

func shouldSyncPath(path string) bool {
	path = filepath.ToSlash(filepath.Clean(path))
	if path == "." || path == ".." || strings.HasPrefix(path, "../") || strings.HasPrefix(path, "/") {
		return false
	}
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return false
	}
	excludedRoots := map[string]bool{
		".agents":     true,
		".build":      true,
		".codex":      true,
		".git":        true,
		".github":     true,
		".gocache":    true,
		".gomodcache": true,
		".idea":       true,
		".tools":      true,
		".vscode":     true,
		"app":         true,
		"bin":         true,
		"dist":        true,
		"docs":        true,
		"tools":       true,
	}
	if excludedRoots[parts[0]] {
		return false
	}
	excludedRootFiles := map[string]bool{
		".gitignore": true,
		"AGENTS.md":  true,
		"CLAUDE.md":  true,
		"LICENSE":    true,
		"README.md":  true,
	}
	if len(parts) == 1 && excludedRootFiles[path] {
		return false
	}
	for _, part := range parts {
		if part == ".DS_Store" || part == "node_modules" || part == ".vite" || part == ".gocache" {
			return false
		}
	}
	if path == "backend/coverage.out" || strings.HasPrefix(path, "frontend/dist/") ||
		strings.HasPrefix(path, "backend/internal/webui/frontend/dist/") ||
		strings.Contains(path, "/testdata/") {
		return false
	}
	return true
}

func hashSourceEntry(path string) (sourceEntry, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return sourceEntry{}, err
	}
	entry := sourceEntry{Mode: uint32(info.Mode().Perm())}
	hash := sha256.New()
	switch {
	case info.Mode().IsRegular():
		entry.Kind = "file"
		file, err := os.Open(path)
		if err != nil {
			return sourceEntry{}, err
		}
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return sourceEntry{}, copyErr
		}
		if closeErr != nil {
			return sourceEntry{}, closeErr
		}
	case info.Mode()&os.ModeSymlink != 0:
		entry.Kind = "symlink"
		target, err := os.Readlink(path)
		if err != nil {
			return sourceEntry{}, err
		}
		_, _ = io.WriteString(hash, target)
	default:
		return sourceEntry{}, fmt.Errorf("unsupported file type %s", info.Mode().Type())
	}
	entry.Hash = hex.EncodeToString(hash.Sum(nil))
	return entry, nil
}

func manifestChanges(previous, current sourceManifest) (changed, deleted []string) {
	for path, entry := range current.Files {
		if old, ok := previous.Files[path]; !ok || old != entry {
			changed = append(changed, path)
		}
	}
	for path := range previous.Files {
		if _, ok := current.Files[path]; !ok {
			deleted = append(deleted, path)
		}
	}
	sort.Strings(changed)
	sort.Strings(deleted)
	return changed, deleted
}

func manifestsEqual(left, right sourceManifest) bool {
	if len(left.Files) != len(right.Files) {
		return false
	}
	for path, entry := range left.Files {
		if right.Files[path] != entry {
			return false
		}
	}
	return true
}

func decodeManifest(raw []byte) (sourceManifest, error) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return sourceManifest{Version: 1, Files: map[string]sourceEntry{}}, nil
	}
	var manifest sourceManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return sourceManifest{}, err
	}
	if manifest.Version != 1 {
		return sourceManifest{}, fmt.Errorf("unsupported manifest version %d", manifest.Version)
	}
	if manifest.Files == nil {
		manifest.Files = map[string]sourceEntry{}
	}
	return manifest, nil
}
