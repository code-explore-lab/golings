package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopath/internal/catalog"
	"gopath/internal/progress"
	"gopath/internal/runner"
	"gopath/internal/workspace"
)

func Run(ctx context.Context, args []string, out, errOut io.Writer, in io.Reader, source fs.FS) error {
	explicit := ""
	if len(args) > 0 && args[0] == "--workspace" {
		if len(args) < 3 {
			return errors.New("usage: gopath --workspace <directory> <command>")
		}
		explicit, args = args[1], args[2:]
	}
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		return help(out)
	}
	exercises, err := catalog.Load(source)
	if err != nil {
		return err
	}
	root, err := resolveRoot(explicit)
	if err != nil {
		return err
	}
	a := App{Out: out, In: in, Root: root, Exercises: exercises, Source: source}
	switch args[0] {
	case "init":
		if len(args) > 2 {
			return errors.New("usage: gopath init [workspace]")
		}
		if len(args) == 2 {
			a.Root, err = filepath.Abs(args[1])
			if err != nil {
				return err
			}
		}
		if err := workspace.Init(a.Root, source); err != nil {
			return err
		}
		_, err := fmt.Fprintf(out, "Initialized learning workspace at %s\nUse: gopath --workspace \"%s\" next\nOr change into the workspace and run gopath next.\n", a.Root, a.Root)
		return err
	case "list", "next", "stats":
		if len(args) != 1 {
			return fmt.Errorf("usage: gopath %s", args[0])
		}
		switch args[0] {
		case "list":
			return a.list()
		case "next":
			return a.next()
		default:
			return a.stats()
		}
	case "check":
		return a.withWriteLock(func() error { return a.check(ctx, args) })
	case "hint":
		return a.withWriteLock(func() error { return a.hint(args) })
	case "reset":
		return a.withWriteLock(func() error { return a.reset(args) })
	default:
		return fmt.Errorf("unknown command %q; run gopath help", args[0])
	}
}

func resolveRoot(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	if root := os.Getenv("GOPATH_HOME"); root != "" {
		return filepath.Abs(root)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		if _, err := os.Lstat(filepath.Join(dir, ".gopath")); err == nil {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	return filepath.Join(cwd, "gopath-workspace"), nil
}

type App struct {
	Out       io.Writer
	In        io.Reader
	Root      string
	Exercises []catalog.Exercise
	Source    fs.FS
}

func (a App) withWriteLock(action func() error) (err error) {
	if err := needWorkspace(a.Root); err != nil {
		return err
	}
	lock, err := progress.AcquireLock(a.Root)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, lock.Close()) }()
	return action()
}

func (a App) load() (progress.File, error) {
	if err := needWorkspace(a.Root); err != nil {
		return progress.File{}, err
	}
	return progress.Load(a.Root)
}

// A record's presence can mean only that a hint was used. Completion requires
// a successful check and a matching source fingerprint.
func (a App) state(ex catalog.Exercise, rec progress.Record) string {
	if rec.LocalPassedAt.IsZero() && rec.CompletedAt.IsZero() {
		return "todo"
	}
	fingerprint, err := workspace.Fingerprint(a.Root, ex.ID)
	if err != nil || fingerprint != rec.Fingerprint {
		return "stale"
	}
	if !rec.CompletedAt.IsZero() {
		return "complete"
	}
	return "local"
}

func (a App) list() error {
	p, err := a.load()
	if err != nil {
		return err
	}
	for _, ex := range a.Exercises {
		mark := map[string]string{"todo": "[ ]", "stale": "[!]", "local": "[~]", "complete": "[x]"}[a.state(ex, p.Exercises[ex.ID])]
		if _, err := fmt.Fprintf(a.Out, "%s %-26s %s (%s)\n", mark, ex.ID, ex.Title, ex.Category); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(a.Out, "[x] complete  [~] local tests passed, integration pending  [!] files changed, recheck required")
	return err
}

func (a App) next() error {
	p, err := a.load()
	if err != nil {
		return err
	}
	for _, ex := range a.Exercises {
		state := a.state(ex, p.Exercises[ex.ID])
		if state == "complete" {
			continue
		}
		command := "check " + ex.ID
		if state == "local" && ex.Integration {
			command = "check --integration " + ex.ID
		}
		_, err := fmt.Fprintf(a.Out, "Next: %s — %s\n%s\nRead: %s\nEdit: %s\nRun: gopath --workspace \"%s\" %s\n", ex.ID, ex.Title, ex.Description, filepath.Join(a.Root, ex.ID, "README.md"), workspace.ExerciseDir(a.Root, ex.ID), a.Root, command)
		return err
	}
	_, err = fmt.Fprintln(a.Out, "All required exercise checks complete.")
	return err
}

func (a App) check(ctx context.Context, args []string) error {
	integration := false
	selector := []string{"check"}
	for _, arg := range args[1:] {
		if arg == "--integration" && !integration {
			integration = true
		} else if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unknown or duplicate check option %q", arg)
		} else {
			selector = append(selector, arg)
		}
	}
	p, err := a.load()
	if err != nil {
		return err
	}
	id, err := a.currentOrArg(selector, p)
	if err != nil {
		return err
	}
	ex, _ := catalog.Find(a.Exercises, id)
	if integration && !ex.Integration {
		return fmt.Errorf("%s has no integration check; run gopath check %s", id, id)
	}
	before, err := workspace.InputFingerprint(a.Root, id)
	if err != nil {
		return a.failed(p, id, err)
	}
	if err := workspace.VerifyTests(a.Root, id, a.Source); err != nil {
		return a.failed(p, id, err)
	}
	fmt.Fprintf(a.Out, "Checking %s (integration=%t); first dependency download may take a few minutes.\n", id, integration)
	output, checkErr := runner.GoTest(ctx, workspace.ExerciseDir(a.Root, id), integration)
	fmt.Fprint(a.Out, output)
	if checkErr != nil {
		return a.failed(p, id, checkErr)
	}
	if err := workspace.VerifyTests(a.Root, id, a.Source); err != nil {
		return a.failed(p, id, fmt.Errorf("exercise inputs changed while tests were running; restore required tests and run gopath check %s again: %w", id, err))
	}
	after, err := workspace.InputFingerprint(a.Root, id)
	if err != nil {
		return a.failed(p, id, err)
	}
	if before != after {
		return a.failed(p, id, fmt.Errorf("exercise inputs changed while tests were running (including possible go.mod updates); run gopath check %s again to verify the current files", id))
	}
	fingerprint, err := workspace.Fingerprint(a.Root, id)
	if err != nil {
		return a.failed(p, id, err)
	}
	// Do not bind a newly edited input to the full fingerprint collected above.
	// Edits after this point retain the old fingerprint and are shown as stale.
	confirmed, err := workspace.InputFingerprint(a.Root, id)
	if err != nil {
		return a.failed(p, id, err)
	}
	if before != confirmed {
		return a.failed(p, id, fmt.Errorf("exercise inputs changed while recording the result; run gopath check %s again", id))
	}
	rec := p.Exercises[id]
	if rec.Fingerprint != fingerprint {
		rec.CompletedAt = time.Time{}
	}
	rec.LocalPassedAt = time.Now().UTC()
	rec.Fingerprint = fingerprint
	if !ex.Integration || integration {
		rec.CompletedAt = rec.LocalPassedAt
	}
	p.Exercises[id] = rec
	if err := progress.Save(a.Root, p); err != nil {
		return err
	}
	if ex.Integration && rec.CompletedAt.IsZero() {
		_, err = fmt.Fprintf(a.Out, "Local checks passed: %s. Integration pending; follow README.md, then run gopath check --integration %s.\n", id, id)
	} else {
		_, err = fmt.Fprintf(a.Out, "Passed all required checks: %s — %s\n", id, ex.Title)
	}
	return err
}

func (a App) failed(p progress.File, id string, cause error) error {
	rec := p.Exercises[id]
	rec.CompletedAt, rec.LocalPassedAt, rec.Fingerprint = time.Time{}, time.Time{}, ""
	if rec.HintsUsed == 0 {
		delete(p.Exercises, id)
	} else {
		p.Exercises[id] = rec
	}
	saveErr := progress.Save(a.Root, p)
	fmt.Fprintln(a.Out, "Checks failed; exercise is not complete.")
	return errors.Join(cause, saveErr)
}

func (a App) hint(args []string) error {
	p, err := a.load()
	if err != nil {
		return err
	}
	id, err := a.currentOrArg(args, p)
	if err != nil {
		return err
	}
	ex, _ := catalog.Find(a.Exercises, id)
	rec := p.Exercises[id]
	if rec.HintsUsed >= len(ex.Hints) {
		_, err := fmt.Fprintln(a.Out, "No more hints available.")
		return err
	}
	if _, err := fmt.Fprintf(a.Out, "Hint %d/%d: %s\n", rec.HintsUsed+1, len(ex.Hints), ex.Hints[rec.HintsUsed]); err != nil {
		return err
	}
	rec.HintsUsed++
	p.Exercises[id] = rec
	return progress.Save(a.Root, p)
}

func (a App) stats() error {
	p, err := a.load()
	if err != nil {
		return err
	}
	complete, local, stale, hints := 0, 0, 0, 0
	for _, ex := range a.Exercises {
		rec := p.Exercises[ex.ID]
		hints += rec.HintsUsed
		switch a.state(ex, rec) {
		case "complete":
			complete++
		case "local":
			local++
		case "stale":
			stale++
		}
	}
	_, err = fmt.Fprintf(a.Out, "Completed %d/%d exercises\nLocal checks passed, integration pending: %d\nChanged since last pass: %d\nHints used: %d\n", complete, len(a.Exercises), local, stale, hints)
	return err
}

func (a App) reset(args []string) error {
	p, err := a.load()
	if err != nil {
		return err
	}
	id, err := a.exerciseArg(args)
	if err != nil {
		return err
	}
	files, err := workspace.ManagedFiles(a.Source, id)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "Reset will overwrite these managed files under %s:\n", workspace.ExerciseDir(a.Root, id))
	for _, name := range files {
		fmt.Fprintf(a.Out, "  %s\n", name)
	}
	if _, err := fmt.Fprintf(a.Out, "Completion and hint records will be cleared. Other files are kept.\nType RESET %s to continue: ", id); err != nil {
		return err
	}
	line, err := bufio.NewReader(a.In).ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}
	if strings.TrimSpace(line) != "RESET "+id {
		_, err = fmt.Fprintln(a.Out, "Cancelled.")
		return err
	}
	if err := workspace.RestoreExercise(a.Root, id, a.Source); err != nil {
		return err
	}
	delete(p.Exercises, id)
	if err := progress.Save(a.Root, p); err != nil {
		return err
	}
	_, err = fmt.Fprintf(a.Out, "Reset %s to starter code.\n", id)
	return err
}

func (a App) exerciseArg(args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("usage: gopath %s <id>", args[0])
	}
	if _, ok := catalog.Find(a.Exercises, args[1]); !ok {
		return "", fmt.Errorf("unknown exercise %q", args[1])
	}
	return args[1], nil
}

func (a App) currentOrArg(args []string, p progress.File) (string, error) {
	if len(args) > 2 {
		return "", fmt.Errorf("usage: gopath %s [id]", args[0])
	}
	if len(args) == 2 {
		return a.exerciseArg(args)
	}
	for _, ex := range a.Exercises {
		if a.state(ex, p.Exercises[ex.ID]) != "complete" {
			return ex.ID, nil
		}
	}
	return "", errors.New("all exercises complete")
}

func needWorkspace(root string) error {
	dir, err := workspace.SafePath(root, ".gopath")
	if err != nil {
		return err
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("workspace is not initialized at %s; run gopath init [workspace]", root)
	}
	return nil
}

func help(out io.Writer) error {
	_, err := fmt.Fprintln(out, "GoPath — hands-on Go and cloud-native engineering exercises\n\nUsage: gopath [--workspace <directory>] <command>\n  init [directory]         Initialize without overwriting files\n  list                     List exercises and check status\n  next                     Show the next unfinished exercise\n  check [id]               Run required local tests (default: next)\n  check --integration [id] Run local and live acceptance tests\n  hint [id]                Show the next progressive hint\n  reset <id>               List affected files and require confirmation\n  stats                    Show persisted learning progress\n\nWorkspace: --workspace, then GOPATH_HOME, then current/ancestor workspace,\notherwise ./gopath-workspace. watch is deferred from this MVP.")
	return err
}
