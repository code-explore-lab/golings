package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"gopath/internal/progress"
	"gopath/internal/workspace"
)

func TestBusyWorkspaceRejectsWritersWithoutChangingState(t *testing.T) {
	for _, command := range []string{"check", "hint", "reset"} {
		t.Run(command, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "learn")
			source := testFS()
			var output bytes.Buffer
			run := func(args []string, input string) error {
				return Run(context.Background(), append([]string{"--workspace", root}, args...), &output, &output, strings.NewReader(input), source)
			}
			if err := run([]string{"init"}, ""); err != nil {
				t.Fatal(err)
			}
			code := filepath.Join(root, "01-one", "main.go")
			const originalCode = "learner code must remain untouched\n"
			if err := os.WriteFile(code, []byte(originalCode), 0644); err != nil {
				t.Fatal(err)
			}
			p := progress.File{Exercises: map[string]progress.Record{"01-one": {HintsUsed: 1}}}
			if err := progress.Save(root, p); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(progress.Path(root))
			if err != nil {
				t.Fatal(err)
			}
			lock, err := progress.AcquireLock(root)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			err = run([]string{command, "01-one"}, "RESET 01-one\n")
			if err == nil || !strings.Contains(err.Error(), "workspace is busy") {
				t.Fatalf("writer was not rejected by lock: %v\n%s", err, output.String())
			}
			after, readErr := os.ReadFile(progress.Path(root))
			if readErr != nil || !bytes.Equal(before, after) {
				t.Fatalf("busy command changed progress: %q %v", after, readErr)
			}
			if data, err := os.ReadFile(code); err != nil || string(data) != originalCode {
				t.Fatalf("busy command changed learner code: %q %v", data, err)
			}
			for _, readOnly := range []string{"list", "next", "stats"} {
				if err := run([]string{readOnly}, ""); err != nil {
					t.Errorf("read-only %s should work while a writer holds the lock: %v", readOnly, err)
				}
			}
		})
	}
}

func TestWriterLockReleasedOnCommandFailure(t *testing.T) {
	root := filepath.Join(t.TempDir(), "learn")
	source := testFS()
	var output bytes.Buffer
	run := func(args ...string) error {
		return Run(context.Background(), append([]string{"--workspace", root}, args...), &output, &output, strings.NewReader(""), source)
	}
	if err := run("init"); err != nil {
		t.Fatal(err)
	}
	if err := run("hint", "missing"); err == nil {
		t.Fatal("invalid hint unexpectedly succeeded")
	}
	if err := run("hint", "01-one"); err != nil {
		t.Fatalf("a failed command leaked its writer lock: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gopath", "write.lock")); !os.IsNotExist(err) {
		t.Fatalf("successful command leaked its writer lock: %v", err)
	}
}

func TestCheckRejectsInputsChangedAfterCompiledTestStarts(t *testing.T) {
	for _, changed := range []string{"code", "required-test", "go.mod", "generated-go.sum"} {
		t.Run(changed, func(t *testing.T) {
			base := t.TempDir()
			root := filepath.Join(base, "learn")
			started := filepath.Join(base, "compiled-test-started")
			release := filepath.Join(base, "release-compiled-test")
			source := testFS()
			source["01-one/tests/main_test.go.txt"] = &fstest.MapFile{Data: []byte(fmt.Sprintf(`package lesson

import (
	"os"
	"testing"
	"time"
)

func TestAnswer(t *testing.T) {
	if Answer() != 42 {
		t.Fatal("wrong")
	}
	if err := os.WriteFile(%q, []byte("started"), 0600); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(%q); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("parent did not release compiled test")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
`, started, release))}
			var output bytes.Buffer
			if err := Run(context.Background(), []string{"--workspace", root, "init"}, &output, &output, strings.NewReader(""), source); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			type result struct {
				output string
				err    error
			}
			done := make(chan result, 1)
			go func() {
				var output bytes.Buffer
				err := Run(ctx, []string{"--workspace", root, "check", "01-one"}, &output, &output, strings.NewReader(""), source)
				done <- result{output.String(), err}
			}()
			// Always release the child even if an assertion above the receive fails.
			defer func() { _ = os.WriteFile(release, []byte("release"), 0600) }()
			wait := time.NewTicker(10 * time.Millisecond)
			defer wait.Stop()
		WaitForCompiledTest:
			for {
				if _, err := os.Stat(started); err == nil {
					break WaitForCompiledTest
				} else if !os.IsNotExist(err) {
					t.Fatal(err)
				}
				select {
				case result := <-done:
					t.Fatalf("check ended before compiled test reached barrier: %v\n%s", result.err, result.output)
				case <-ctx.Done():
					t.Fatal("compiled test did not reach barrier:", ctx.Err())
				case <-wait.C:
				}
			}
			dir := workspace.ExerciseDir(root, "01-one")
			var name, data string
			switch changed {
			case "code":
				name, data = "main.go", "package lesson\nfunc Answer() int { return 0 }\n"
			case "required-test":
				name, data = "main_test.go", "package lesson\n"
			case "go.mod":
				name, data = "go.mod", "module changed\n\ngo 1.23\n"
			case "generated-go.sum":
				name, data = "go.sum", ""
			}
			if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(release, []byte("release"), 0600); err != nil {
				t.Fatal(err)
			}
			var checked result
			select {
			case checked = <-done:
			case <-ctx.Done():
				t.Fatal("check did not finish after release:", ctx.Err())
			}
			if !strings.Contains(checked.output, "--- PASS: TestAnswer") {
				t.Fatalf("regression must exercise a genuinely passing compiled test:\n%s\n%v", checked.output, checked.err)
			}
			p, err := progress.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			rec := p.Exercises["01-one"]
			if changed == "generated-go.sum" {
				if checked.err != nil || rec.CompletedAt.IsZero() || rec.LocalPassedAt.IsZero() {
					t.Fatalf("generated go.sum should not invalidate checked inputs: %v\n%s\n%+v", checked.err, checked.output, rec)
				}
				full, err := workspace.Fingerprint(root, "01-one")
				if err != nil || rec.Fingerprint != full {
					t.Fatalf("completion must retain the full fingerprint including go.sum: %v %+v", err, rec)
				}
				return
			}
			if checked.err == nil || !strings.Contains(checked.err.Error(), "changed while tests were running") ||
				!strings.Contains(checked.err.Error(), "run gopath check 01-one again") {
				t.Fatalf("mid-run change was accepted or lacks actionable recheck message: %v\n%s", checked.err, checked.output)
			}
			if !rec.CompletedAt.IsZero() || !rec.LocalPassedAt.IsZero() {
				t.Fatalf("untested edits recorded as passing: %+v", rec)
			}
			if _, err := os.Stat(filepath.Join(root, ".gopath", "write.lock")); !os.IsNotExist(err) {
				t.Fatalf("rejected check leaked its writer lock: %v", err)
			}
		})
	}
}
