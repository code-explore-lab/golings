package exercises_test

import (
	"errors"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"strings"
	"testing"

	"gopath/exercises"
	"gopath/internal/catalog"
)

func TestEmbeddedCurriculumIsUsableAndDoesNotContainSolutions(t *testing.T) {
	lessons, err := catalog.Load(exercises.FS)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, lesson := range lessons {
		ids[lesson.ID] = true
	}
	for _, lesson := range lessons {
		t.Run(lesson.ID, func(t *testing.T) {
			if lesson.Description == "" || lesson.Category == "" || lesson.Difficulty == "" || lesson.Concept == "" ||
				len(lesson.Objectives) == 0 || len(lesson.Hints) < 2 || len(lesson.Acceptance) == 0 {
				t.Fatal("exercise metadata lacks description, objectives, progressive hints, concept or acceptance")
			}
			for _, prerequisite := range lesson.Prerequisites {
				if prerequisite == lesson.ID || !ids[prerequisite] {
					t.Fatalf("invalid prerequisite %q", prerequisite)
				}
			}
			for _, name := range []string{"README.md", "starter/go.mod.txt", "starter/exercise.go.txt", "tests/exercise_test.go.txt"} {
				if _, err := fs.Stat(exercises.FS, lesson.ID+"/"+name); err != nil {
					t.Fatal(err)
				}
			}
			if lesson.Integration {
				data, err := fs.ReadFile(exercises.FS, lesson.ID+"/tests/integration_test.go.txt")
				if err != nil || !strings.Contains(string(data), "//go:build integration") ||
					!strings.Contains(string(data), "func TestIntegration") {
					t.Fatalf("missing tagged live acceptance test: %v", err)
				}
			}
			if _, err := fs.Stat(exercises.FS, lesson.ID+"/solution"); !errors.Is(err, fs.ErrNotExist) {
				t.Fatal("solution embedded")
			}
			err := fs.WalkDir(exercises.FS, lesson.ID, func(name string, entry fs.DirEntry, err error) error {
				if err != nil || entry.IsDir() || !strings.HasSuffix(name, ".go.txt") {
					return err
				}
				data, err := fs.ReadFile(exercises.FS, name)
				if err != nil {
					return err
				}
				_, err = parser.ParseFile(token.NewFileSet(), path.Base(name), data, parser.AllErrors)
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
