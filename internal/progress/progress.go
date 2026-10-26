package progress

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopath/internal/workspace"
)

const Version = 1

type Record struct {
	CompletedAt   time.Time `json:"completed_at,omitempty"`
	LocalPassedAt time.Time `json:"local_passed_at,omitempty"`
	Fingerprint   string    `json:"fingerprint,omitempty"`
	HintsUsed     int       `json:"hints_used,omitempty"`
}
type File struct {
	Version   int               `json:"version"`
	Exercises map[string]Record `json:"exercises"`
}

func Path(workspace string) string { return filepath.Join(workspace, ".gopath", "progress.json") }

func Load(workspace string) (File, error) {
	dest, err := workspacePath(workspace)
	if err != nil {
		return File{}, err
	}
	data, err := os.ReadFile(dest)
	if errors.Is(err, os.ErrNotExist) {
		return empty(), nil
	}
	if err != nil {
		return File{}, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, fmt.Errorf("read progress: %w", err)
	}
	if f.Version != Version {
		return File{}, fmt.Errorf("unsupported progress version %d (expected %d)", f.Version, Version)
	}
	if f.Exercises == nil {
		f.Exercises = map[string]Record{}
	}
	for id, record := range f.Exercises {
		if record.HintsUsed < 0 {
			return File{}, fmt.Errorf("read progress: negative hint count for %s", id)
		}
	}
	return f, nil
}

func Save(workspace string, f File) error {
	dest, err := workspacePath(workspace)
	if err != nil {
		return err
	}
	f.Version = Version
	if f.Exercises == nil {
		f.Exercises = map[string]Record{}
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "progress-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return err
	}
	return nil
}

func workspacePath(root string) (string, error) {
	return workspace.SafePath(root, filepath.Join(".gopath", "progress.json"))
}

func empty() File { return File{Version: Version, Exercises: map[string]Record{}} }
