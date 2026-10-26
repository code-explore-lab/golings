package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func fixture() fstest.MapFS {
	return fstest.MapFS{"x/exercise.json": {Data: []byte("{}")}, "x/starter/main.go.txt": {Data: []byte("package main\n")}, "x/tests/main_test.go.txt": {Data: []byte("package main\n")}, "x/solution/main.go.txt": {Data: []byte("SECRET\n")}, "x/README.md": {Data: []byte("lesson")}}
}

func TestDuplicateFlattenedAssetsRejectedBeforeWrites(t *testing.T) {
	source := fixture()
	source["x/starter/main_test.go.txt"] = &fstest.MapFile{Data: []byte("conflict")}
	root := filepath.Join(t.TempDir(), "learn")
	if err := Init(root, source); err == nil {
		t.Fatal("ambiguous template accepted")
	}
	if _, err := os.Stat(filepath.Join(root, "x", "main.go")); !os.IsNotExist(err) {
		t.Fatal("partial initialization")
	}
}

func TestNestedAssetsAndRequiredTestProtection(t *testing.T) {
	source := fixture()
	source["x/starter/chart/templates/deploy.yaml.txt"] = &fstest.MapFile{Data: []byte("kind: Deployment\n")}
	root := filepath.Join(t.TempDir(), "learn")
	if err := Init(root, source); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "x", "chart", "templates", "deploy.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := VerifyTests(root, "x", source); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "x", "main_test.go"), []byte("edited"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyTests(root, "x", source); err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("got %v", err)
	}
}

func TestRefusesPathEscapeAndLinkedExercise(t *testing.T) {
	if _, err := SafePath(t.TempDir(), "../outside"); err == nil {
		t.Fatal("escape accepted")
	}
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "x")); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	if err := Init(root, fixture()); err == nil {
		t.Fatal("followed linked exercise")
	}
	if _, err := os.Stat(filepath.Join(outside, "main.go")); !os.IsNotExist(err) {
		t.Fatal("wrote outside workspace")
	}
}
func TestInitCopiesTemplatesAndExcludesSolutions(t *testing.T) {
	root := filepath.Join(t.TempDir(), "learn")
	if err := Init(root, fixture()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "x", "main.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "x", "solution", "main.go")); !os.IsNotExist(err) {
		t.Fatal("solution was copied")
	}
	if _, err := os.Stat(filepath.Join(root, "x", "README.md")); err != nil {
		t.Fatal("exercise instructions were not installed", err)
	}
}
func TestInitNeverOverwritesLearnerCode(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "x", "main.go")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("learner"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Init(root, fixture()); err == nil {
		t.Fatal("expected conflict")
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != "learner" {
		t.Fatalf("learner file changed: %q %v", got, err)
	}
}
func TestRestoreRestoresStarter(t *testing.T) {
	root := t.TempDir()
	if err := Init(root, fixture()); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(root, "x", "main.go")
	if err := os.WriteFile(dest, []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := RestoreExercise(root, "x", fixture()); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != "package main\n" {
		t.Fatalf("got %q", got)
	}
}

func TestInputFingerprintIgnoresOnlyGeneratedRootGoSum(t *testing.T) {
	root := t.TempDir()
	if err := Init(root, fixture()); err != nil {
		t.Fatal(err)
	}
	input, err := InputFingerprint(root, "x")
	if err != nil {
		t.Fatal(err)
	}
	full, err := Fingerprint(root, "x")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "x", "go.sum"), []byte("generated checksum\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, err := InputFingerprint(root, "x"); err != nil || got != input {
		t.Fatalf("generated go.sum changed input fingerprint: %q %v", got, err)
	}
	if got, err := Fingerprint(root, "x"); err != nil || got == full {
		t.Fatalf("full fingerprint must include go.sum: %q %v", got, err)
	}
	for _, name := range []string{"main.go", "go.mod", "deployment.yaml", "dashboard.json", "Dockerfile", "nested/go.sum"} {
		before, err := InputFingerprint(root, "x")
		if err != nil {
			t.Fatal(err)
		}
		dest := filepath.Join(root, "x", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dest, []byte("changed "+name), 0644); err != nil {
			t.Fatal(err)
		}
		if got, err := InputFingerprint(root, "x"); err != nil || got == before {
			t.Errorf("input %s change must require recheck: %q %v", name, got, err)
		}
	}
}
