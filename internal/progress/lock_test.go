package progress

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceLockRejectsConcurrentWriterAndReleases(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".gopath"), 0755); err != nil {
		t.Fatal(err)
	}
	first, err := AcquireLock(root)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	path := filepath.Join(root, ".gopath", "write.lock")
	before, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(before), "pid=") || !strings.Contains(string(before), "started=") {
		t.Fatalf("lock must identify its owner and start time: %q %v", before, err)
	}
	second, err := AcquireLock(root)
	if err == nil {
		second.Close()
		t.Fatal("concurrent writer acquired an existing lock")
	}
	for _, want := range []string{"workspace is busy", path, "previous process exited", "remove this stale lock file manually"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("busy error %q lacks recovery detail %q", err, want)
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatalf("a rejected acquisition altered the existing lock: %q %v", after, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("released lock still exists: %v", err)
	}
	second, err = AcquireLock(root)
	if err != nil {
		t.Fatalf("lock was not released: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceLockDoesNotDeleteStaleFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".gopath", "write.lock")
	if err := os.Mkdir(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	const stale = "pid=999999999\nstarted=2000-01-01T00:00:00Z\n"
	if err := os.WriteFile(path, []byte(stale), 0600); err != nil {
		t.Fatal(err)
	}
	if lock, err := AcquireLock(root); err == nil {
		lock.Close()
		t.Fatal("stale-looking lock was automatically replaced")
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != stale {
		t.Fatalf("stale lock must be preserved for manual recovery: %q %v", data, err)
	}
}
