package progress

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopath/internal/workspace"
)

// Lock serializes commands that modify one workspace's progress or exercises.
// Its file deliberately survives a process crash; stale locks need manual review.
type Lock struct {
	path string
	file *os.File
}

func AcquireLock(root string) (*Lock, error) {
	dest, err := workspace.SafePath(root, filepath.Join(".gopath", "write.lock"))
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("workspace is busy: another check, hint or reset holds %s; wait for it to finish. If the previous process exited, verify no GoPath writer is running, remove this stale lock file manually, and retry", dest)
	}
	if err != nil {
		return nil, fmt.Errorf("acquire workspace write lock: %w", err)
	}
	if _, err := fmt.Fprintf(file, "pid=%d\nstarted=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339)); err != nil {
		file.Close()
		os.Remove(dest)
		return nil, fmt.Errorf("write workspace lock: %w", err)
	}
	return &Lock{path: dest, file: file}, nil
}

func (lock *Lock) Close() error {
	if lock.file == nil {
		return nil
	}
	err := lock.file.Close()
	lock.file = nil
	if removeErr := os.Remove(lock.path); removeErr != nil {
		err = errors.Join(err, fmt.Errorf("release workspace write lock %s: %w; verify no GoPath writer is running before removing a stale lock manually", lock.path, removeErr))
	}
	return err
}
