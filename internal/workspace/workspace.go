package workspace

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type item struct {
	path string
	data []byte
	test bool
}

func assets(source fs.FS, id string) ([]item, error) {
	var out []item
	seen := map[string]bool{}
	err := fs.WalkDir(source, id, func(name string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			if path.Base(name) == "solution" {
				return fs.SkipDir
			}
			return nil
		}
		rel := strings.TrimPrefix(name, id+"/")
		isTest := strings.HasPrefix(rel, "tests/")
		if rel != "README.md" && !strings.HasPrefix(rel, "starter/") && !isTest {
			return nil
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		rel = strings.TrimSuffix(rel, ".txt")
		rel = strings.TrimPrefix(strings.TrimPrefix(rel, "starter/"), "tests/")
		if !filepath.IsLocal(filepath.FromSlash(rel)) || seen[strings.ToLower(rel)] {
			return fmt.Errorf("unsafe or duplicate template path %s", name)
		}
		seen[strings.ToLower(rel)] = true
		out = append(out, item{filepath.FromSlash(rel), data, isTest})
		return nil
	})
	return out, err
}

// SafePath rejects links and non-directory parents before a read or write.
// It is a guard against accidental path escapes, not a system-level sandbox.
func SafePath(root, rel string) (string, error) {
	if !filepath.IsLocal(rel) {
		return "", fmt.Errorf("path must stay inside workspace: %s", rel)
	}
	base, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	dest := filepath.Join(base, rel)
	current := base
	parts := append([]string{""}, strings.Split(filepath.Clean(rel), string(filepath.Separator))...)
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			break
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("refusing linked path %s", current)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return "", fmt.Errorf("parent is not a directory: %s", current)
		}
	}
	return dest, nil
}

func Init(root string, source fs.FS) error {
	if root == "" {
		return errors.New("workspace path is required")
	}
	// Preflight every destination before writing any files.
	marker, err := SafePath(root, ".gopath")
	if err != nil {
		return err
	}
	if _, err := os.Lstat(marker); err == nil {
		return fmt.Errorf("workspace already initialized at %s", root)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	entries, err := fs.ReadDir(source, ".")
	if err != nil {
		return err
	}
	var planned []item
	for _, ex := range entries {
		if !ex.IsDir() {
			continue
		}
		if _, err := fs.Stat(source, ex.Name()+"/exercise.json"); err != nil {
			continue
		}
		parts, err := assets(source, ex.Name())
		if err != nil {
			return err
		}
		for _, a := range parts {
			a.path, err = SafePath(root, filepath.Join(ex.Name(), a.path))
			if err != nil {
				return err
			}
			if _, err := os.Lstat(a.path); err == nil {
				return fmt.Errorf("refusing to overwrite existing file %s", a.path)
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
			planned = append(planned, a)
		}
	}
	for _, a := range planned {
		if err := os.MkdirAll(filepath.Dir(a.path), 0755); err != nil {
			return err
		}
		f, err := os.OpenFile(a.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		_, writeErr := f.Write(a.data)
		closeErr := f.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			return err
		}
	}
	return os.MkdirAll(marker, 0755)
}

func ExerciseDir(root, id string) string { return filepath.Join(root, id) }

func ManagedFiles(source fs.FS, id string) ([]string, error) {
	parts, err := assets(source, id)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(parts))
	for _, a := range parts {
		names = append(names, a.path)
	}
	sort.Strings(names)
	return names, nil
}

func RestoreExercise(root, id string, source fs.FS) error {
	base, err := SafePath(root, id)
	if err != nil {
		return err
	}
	info, err := os.Stat(base)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("exercise %s is not installed as a directory", id)
	}
	parts, err := assets(source, id)
	if err != nil {
		return err
	}
	// Preflight the entire reset, so a later linked path cannot cause an
	// otherwise avoidable partial reset.
	for _, a := range parts {
		dest, err := SafePath(base, a.path)
		if err != nil {
			return err
		}
		if info, err := os.Lstat(dest); err == nil && !info.Mode().IsRegular() {
			return fmt.Errorf("cannot restore over non-file %s", dest)
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	for _, a := range parts {
		dest := filepath.Join(base, a.path)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		// Replace instead of truncating to preserve hard-linked copies.
		if err := replaceFile(dest, a.data); err != nil {
			return err
		}
	}
	return nil
}

func replaceFile(dest string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(dest), ".gopath-reset-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(f.Name(), dest)
}

// VerifyTests prevents accidentally passing by deleting or changing required
// tests. Learners may add their own test files.
func VerifyTests(root, id string, source fs.FS) error {
	parts, err := assets(source, id)
	if err != nil {
		return err
	}
	count := 0
	for _, a := range parts {
		if !a.test {
			continue
		}
		count++
		dest, err := SafePath(root, filepath.Join(id, a.path))
		if err != nil {
			return err
		}
		data, err := os.ReadFile(dest)
		if err != nil {
			return fmt.Errorf("required test %s is missing: %w", a.path, err)
		}
		if !bytes.Equal(data, a.data) {
			return fmt.Errorf("required test %s was changed; restore it from the curriculum (reset requires confirmation)", a.path)
		}
	}
	if count == 0 {
		return errors.New("exercise has no required tests")
	}
	return nil
}

// Fingerprint invalidates recorded checks when relevant learner files change.
func Fingerprint(root, id string) (string, error) {
	return fingerprint(root, id, false)
}

// InputFingerprint identifies the inputs checked by go test. The root go.sum
// may be generated during dependency resolution, so it is excluded here only;
// the full recorded Fingerprint still includes it.
func InputFingerprint(root, id string) (string, error) {
	return fingerprint(root, id, true)
}

func fingerprint(root, id string, ignoreGoSum bool) (string, error) {
	dir, err := SafePath(root, id)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	err = fs.WalkDir(os.DirFS(dir), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing linked exercise file %s", name)
		}
		if entry.IsDir() {
			return nil
		}
		if ignoreGoSum && name == "go.sum" {
			return nil
		}
		ext := path.Ext(name)
		if ext != ".go" && ext != ".json" && ext != ".yaml" && ext != ".yml" &&
			path.Base(name) != "go.mod" && path.Base(name) != "go.sum" && path.Base(name) != "Dockerfile" {
			return nil
		}
		data, err := fs.ReadFile(os.DirFS(dir), name)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%d:%s:%d:", len(name), name, len(data))
		hash.Write(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
