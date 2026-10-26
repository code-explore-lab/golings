package cli

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"gopath/internal/progress"
	"gopath/internal/workspace"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{"01-one/exercise.json": {Data: []byte("{\"id\":\"01-one\",\"title\":\"One\",\"category\":\"Go\",\"difficulty\":\"easy\",\"hints\":[\"hint\"],\"check\":\"go-test\"}")}, "01-one/starter/go.mod.txt": {Data: []byte("module lesson\n\ngo 1.20\n")}, "01-one/starter/main.go.txt": {Data: []byte("package lesson\nfunc Answer() int { return 42 }\n")}, "01-one/tests/main_test.go.txt": {Data: []byte("package lesson\nimport \"testing\"\nfunc TestAnswer(t *testing.T){if Answer()!=42{t.Fatal(\"wrong\")}}\n")}, "01-one/solution/main.go.txt": {Data: []byte("SECRET\n")}, "01-one/README.md": {Data: []byte("lesson\n")}}
}
func call(t *testing.T, root string, source fs.FS, args ...string) (string, error) {
	t.Helper()
	t.Setenv("GOPATH_HOME", root)
	var out, errOut bytes.Buffer
	err := Run(context.Background(), args, &out, &errOut, strings.NewReader(""), source)
	return out.String() + errOut.String(), err
}
func TestInitAndCheckWorkflow(t *testing.T) {
	root := filepath.Join(t.TempDir(), "learn")
	if _, err := call(t, root, testFS(), "init"); err != nil {
		t.Fatal(err)
	}
	if _, err := call(t, root, testFS(), "check", "01-one"); err != nil {
		t.Fatal(err)
	}
	p, err := progress.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if p.Exercises["01-one"].CompletedAt.IsZero() {
		t.Fatal("not marked complete")
	}
}
func TestFailedCheckDoesNotComplete(t *testing.T) {
	root := filepath.Join(t.TempDir(), "learn")
	if _, err := call(t, root, testFS(), "init"); err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(root, "01-one", "main.go")
	if err := os.WriteFile(main, []byte("package lesson\nfunc Answer() int{return 0}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := call(t, root, testFS(), "check", "01-one")
	if err == nil || !strings.Contains(out, "not complete") {
		t.Fatalf("err=%v out=%s", err, out)
	}
	p, _ := progress.Load(root)
	if _, ok := p.Exercises["01-one"]; ok {
		t.Fatal("failed exercise marked complete")
	}
}
func TestHintAdvancesAndResetRequiresConfirmation(t *testing.T) {
	root := filepath.Join(t.TempDir(), "learn")
	if _, err := call(t, root, testFS(), "init"); err != nil {
		t.Fatal(err)
	}
	out, err := call(t, root, testFS(), "hint", "01-one")
	if err != nil || !strings.Contains(out, "hint") {
		t.Fatalf("%s %v", out, err)
	}
	main := filepath.Join(root, "01-one", "main.go")
	if err := os.WriteFile(main, []byte("learner change"), 0644); err != nil {
		t.Fatal(err)
	}
	out, err = call(t, root, testFS(), "reset", "01-one")
	if err != nil || !strings.Contains(out, "Cancelled") {
		t.Fatalf("%s %v", out, err)
	}
	got, _ := os.ReadFile(main)
	if string(got) != "learner change" {
		t.Fatal("cancelled reset overwrote code")
	}
}
func TestReferenceSolutionNeverCopied(t *testing.T) {
	root := filepath.Join(t.TempDir(), "learn")
	if _, err := call(t, root, testFS(), "init"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "01-one", "solution")); !os.IsNotExist(err) {
		t.Fatal("reference solution exposed")
	}
}

func TestHintsDoNotCompleteOrAdvance(t *testing.T) {
	root := filepath.Join(t.TempDir(), "learn")
	source := testFS()
	if _, err := call(t, root, source, "init"); err != nil {
		t.Fatal(err)
	}
	if _, err := call(t, root, source, "hint"); err != nil {
		t.Fatal(err)
	}
	out, err := call(t, root, source, "stats")
	if err != nil || !strings.Contains(out, "Completed 0/1") {
		t.Fatalf("%s %v", out, err)
	}
	out, err = call(t, root, source, "next")
	if err != nil || !strings.Contains(out, "Next: 01-one") {
		t.Fatalf("%s %v", out, err)
	}
}

func TestDeletedTestInvalidatesEarlierCompletion(t *testing.T) {
	root := filepath.Join(t.TempDir(), "learn")
	source := testFS()
	if _, err := call(t, root, source, "init"); err != nil {
		t.Fatal(err)
	}
	hash, err := workspace.Fingerprint(root, "01-one")
	if err != nil {
		t.Fatal(err)
	}
	p := progress.File{Exercises: map[string]progress.Record{"01-one": {
		CompletedAt: time.Now(), LocalPassedAt: time.Now(), Fingerprint: hash, HintsUsed: 1,
	}}}
	if err := progress.Save(root, p); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "01-one", "main_test.go")); err != nil {
		t.Fatal(err)
	}
	out, err := call(t, root, source, "stats")
	if err != nil || !strings.Contains(out, "Completed 0/1") || !strings.Contains(out, "Changed since last pass: 1") {
		t.Fatalf("%s %v", out, err)
	}
	_, err = call(t, root, source, "check")
	if err == nil || !strings.Contains(err.Error(), "required test") {
		t.Fatalf("got %v", err)
	}
	p, err = progress.Load(root)
	if err != nil || !p.Exercises["01-one"].CompletedAt.IsZero() || p.Exercises["01-one"].HintsUsed != 1 {
		t.Fatalf("completion not cleared or hints lost: %#v %v", p, err)
	}
}

func TestConfirmedResetAndCorruptProgressPreflight(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(map[bool]string{false: "confirmed", true: "corrupt"}[corrupt], func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "learn")
			source := testFS()
			if _, err := call(t, root, source, "init"); err != nil {
				t.Fatal(err)
			}
			if _, err := call(t, root, source, "hint"); err != nil {
				t.Fatal(err)
			}
			dest := filepath.Join(root, "01-one", "main.go")
			if err := os.WriteFile(dest, []byte("learner change"), 0644); err != nil {
				t.Fatal(err)
			}
			extra := filepath.Join(root, "01-one", "notes.txt")
			if err := os.WriteFile(extra, []byte("keep"), 0644); err != nil {
				t.Fatal(err)
			}
			if corrupt {
				if err := os.WriteFile(progress.Path(root), []byte("invalid"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			var out bytes.Buffer
			err := Run(context.Background(), []string{"reset", "01-one"}, &out, &out, strings.NewReader("RESET 01-one\n"), source)
			data, readErr := os.ReadFile(dest)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if corrupt {
				if err == nil || string(data) != "learner change" {
					t.Fatalf("reset ignored corrupt state: %v %s", err, data)
				}
			} else {
				if err != nil || !strings.Contains(string(data), "return 42") {
					t.Fatalf("%s %v", out.String(), err)
				}
				p, err := progress.Load(root)
				if err != nil || len(p.Exercises) != 0 {
					t.Fatalf("%#v %v", p, err)
				}
			}
			if data, err := os.ReadFile(extra); err != nil || string(data) != "keep" {
				t.Fatalf("extra file changed: %s %v", data, err)
			}
		})
	}
}

func TestIntegrationRequiresSeparatePassingStage(t *testing.T) {
	source := testFS()
	meta := strings.Replace(string(source["01-one/exercise.json"].Data), "\"check\":\"go-test\"", "\"integration\":true,\"check\":\"go-test\"", 1)
	source["01-one/exercise.json"] = &fstest.MapFile{Data: []byte(meta)}
	source["01-one/tests/integration_test.go.txt"] = &fstest.MapFile{Data: []byte("//go:build integration\n\npackage lesson\nimport \"testing\"\nfunc TestIntegrationAcceptance(t *testing.T){}\n")}
	root := filepath.Join(t.TempDir(), "learn")
	if _, err := call(t, root, source, "init"); err != nil {
		t.Fatal(err)
	}
	if out, err := call(t, root, source, "check"); err != nil {
		t.Fatalf("%s %v", out, err)
	}
	p, err := progress.Load(root)
	if err != nil || p.Exercises["01-one"].LocalPassedAt.IsZero() || !p.Exercises["01-one"].CompletedAt.IsZero() {
		t.Fatalf("%#v %v", p, err)
	}
	if out, err := call(t, root, source, "check", "--integration"); err != nil {
		t.Fatalf("%s %v", out, err)
	}
	p, err = progress.Load(root)
	if err != nil || p.Exercises["01-one"].CompletedAt.IsZero() {
		t.Fatalf("%#v %v", p, err)
	}
}

func TestExplicitWorkspaceOverridesEnvironment(t *testing.T) {
	t.Setenv("GOPATH_HOME", filepath.Join(t.TempDir(), "wrong"))
	root := filepath.Join(t.TempDir(), "chosen")
	var out bytes.Buffer
	if err := Run(context.Background(), []string{"--workspace", root, "init"}, &out, &out, strings.NewReader(""), testFS()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "01-one", "README.md")); err != nil {
		t.Fatal(err)
	}
	if err := Run(context.Background(), []string{"--workspace", root, "next"}, &out, &out, strings.NewReader(""), testFS()); err != nil {
		t.Fatal(err)
	}
}
