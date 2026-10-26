//go:build curriculum

package exercises_test

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopath/exercises"
	"gopath/internal/catalog"
	"gopath/internal/runner"
	"gopath/internal/workspace"
)

// This explicitly selected author check may read reference solutions from the
// source checkout. The learner CLI never imports or runs them.
func TestCurriculumStarterFailuresAndSolutions(t *testing.T) {
	lessons, err := catalog.Load(exercises.FS)
	if err != nil {
		t.Fatal(err)
	}
	for _, lesson := range lessons {
		t.Run(lesson.ID, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if err := workspace.Init(root, exercises.FS); err != nil {
				t.Fatal(err)
			}
			dir := workspace.ExerciseDir(root, lesson.ID)
			compile := func(integration bool) {
				t.Helper()
				args := []string{"test", "-mod=mod", "-run=^$", "-timeout=2m"}
				if integration {
					args = append(args, "-tags=integration")
				}
				cmd := exec.CommandContext(context.Background(), "go", append(args, "./...")...)
				cmd.Dir = dir
				cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS= ")
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("exercise must compile: %v\n%s", err, out)
				}
			}
			compile(false)
			out, err := runner.GoTest(context.Background(), dir, false)
			if err == nil {
				t.Fatalf("starter unexpectedly passed:\n%s", out)
			}
			if !strings.Contains(err.Error(), "go test failed") {
				t.Fatalf("starter failed without meaningful test failure: %v\n%s", err, out)
			}
			t.Logf("starter compiles and required tests fail")
			source := os.DirFS(".")
			solutionDir := lesson.ID + "/solution"
			files := 0
			err = fs.WalkDir(source, solutionDir, func(name string, entry fs.DirEntry, err error) error {
				if err != nil || entry.IsDir() {
					return err
				}
				files++
				data, err := fs.ReadFile(source, name)
				if err != nil {
					return err
				}
				rel := strings.TrimSuffix(strings.TrimPrefix(name, solutionDir+"/"), ".txt")
				dest, err := workspace.SafePath(dir, filepath.FromSlash(rel))
				if err != nil {
					return err
				}
				if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
					return err
				}
				return os.WriteFile(dest, data, 0644)
			})
			if err != nil || files == 0 {
				t.Fatalf("reference solution missing: %v", err)
			}
			if out, err := runner.GoTest(context.Background(), dir, false); err != nil {
				t.Fatalf("solution failed: %v\n%s", err, out)
			}
			if lesson.Integration {
				compile(true)
			}
			t.Log("solution passes; live test code compiles (live environment not exercised)")
		})
	}
}
