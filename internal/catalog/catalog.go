package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
)

var idPattern = regexp.MustCompile("^[a-zA-Z0-9][a-zA-Z0-9_-]*$")

type Exercise struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Category      string   `json:"category"`
	Difficulty    string   `json:"difficulty"`
	Description   string   `json:"description"`
	Objectives    []string `json:"objectives"`
	Hints         []string `json:"hints"`
	Acceptance    []string `json:"acceptance"`
	Concept       string   `json:"concept"`
	Check         string   `json:"check"`
	Integration   bool     `json:"integration,omitempty"`
	Prerequisites []string `json:"prerequisites,omitempty"`
	SourceDir     string   `json:"-"`
}

func Load(fsys fs.FS) ([]Exercise, error) {
	var out []Exercise
	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || path.Base(name) != "exercise.json" {
			return nil
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}
		var ex Exercise
		if err := json.Unmarshal(data, &ex); err != nil {
			return fmt.Errorf("%s: invalid metadata: %w", name, err)
		}
		if !idPattern.MatchString(ex.ID) {
			return fmt.Errorf("%s: invalid exercise id %q", name, ex.ID)
		}
		if path.Base(path.Dir(name)) != ex.ID {
			return fmt.Errorf("%s: directory must match id %q", name, ex.ID)
		}
		if ex.Title == "" {
			return fmt.Errorf("%s: title is required", name)
		}
		if ex.Check != "go-test" {
			return fmt.Errorf("%s: unsupported check %q (supported: go-test)", name, ex.Check)
		}
		ex.SourceDir = path.Dir(name)
		out = append(out, ex)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errors.New("exercise catalog is empty")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	for i := 1; i < len(out); i++ {
		if out[i-1].ID == out[i].ID {
			return nil, fmt.Errorf("duplicate exercise id %q", out[i].ID)
		}
	}
	return out, nil
}

func Find(exercises []Exercise, id string) (Exercise, bool) {
	for _, ex := range exercises {
		if ex.ID == id {
			return ex, true
		}
	}
	return Exercise{}, false
}
