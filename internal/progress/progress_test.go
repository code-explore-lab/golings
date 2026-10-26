package progress

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	root := t.TempDir()
	want := File{Exercises: map[string]Record{"x": {CompletedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), HintsUsed: 2}}}
	if err := Save(root, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != Version || got.Exercises["x"].HintsUsed != 2 {
		t.Fatalf("unexpected file: %#v", got)
	}
	// A second save exercises replacement on Windows as well as Unix.
	want.Exercises["x"] = Record{HintsUsed: 3}
	if err := Save(root, want); err != nil {
		t.Fatal(err)
	}
	got, err = Load(root)
	if err != nil || got.Exercises["x"].HintsUsed != 3 {
		t.Fatalf("%#v %v", got, err)
	}
}

func TestLoadRejectsNegativeHints(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Dir(Path(root)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(root), []byte("{\"version\":1,\"exercises\":{\"x\":{\"hints_used\":-1}}}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("negative hint count accepted")
	}
}
func TestLoadRejectsUnknownVersion(t *testing.T) {
	root := t.TempDir()
	path := Path(root)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{\"version\":99}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("expected unknown version to fail")
	}
}
