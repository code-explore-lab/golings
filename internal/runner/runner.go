package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func GoTest(ctx context.Context, dir string, integration bool) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	// -mod=mod lets self-contained exercise modules resolve their declared
	// dependencies and write go.sum on first run instead of failing readonly.
	args := []string{"test", "-mod=mod", "-count=1", "-json", "-timeout=2m"}
	if integration {
		args = append(args, "-tags=integration")
	}
	args = append(args, "./...")
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = testEnv(integration)
	var events, diagnostics bytes.Buffer
	cmd.Stdout, cmd.Stderr = &events, &diagnostics
	runErr := cmd.Run()
	output, passed, skipped, live, parseErr := summarize(events.Bytes())
	output = diagnostics.String() + output
	if runErr != nil {
		if ctx.Err() != nil {
			return output, ctx.Err()
		}
		return output, fmt.Errorf("go test failed in %s: %w", filepath.Base(dir), runErr)
	}
	if parseErr != nil {
		return output, fmt.Errorf("read go test results: %w", parseErr)
	}
	if passed == 0 || skipped != 0 {
		return output, fmt.Errorf("required checks did not all run: %d passed tests, %d skipped tests", passed, skipped)
	}
	if integration && live == 0 {
		return output, fmt.Errorf("no TestIntegration tests passed; cannot record live acceptance")
	}
	return output, nil
}

func testEnv(integration bool) []string {
	var env []string
	for _, value := range os.Environ() {
		name, _, _ := strings.Cut(value, "=")
		if !strings.EqualFold(name, "GOWORK") && !strings.EqualFold(name, "GOFLAGS") && !strings.EqualFold(name, "GOPATH_INTEGRATION") {
			env = append(env, value)
		}
	}
	value := "0"
	if integration {
		value = "1"
	}
	// An empty GOFLAGS falls back to the user's saved go env configuration.
	// Whitespace is a nonempty override that the Go flag parser treats as no flags.
	return append(env, "GOWORK=off", "GOFLAGS= ", "GOPATH_INTEGRATION="+value)
}

func summarize(data []byte) (output string, passed, skipped, live int, err error) {
	var text strings.Builder
	decoder := json.NewDecoder(bytes.NewReader(data))
	for decoder.More() {
		var event struct {
			Action, Test, Output string
		}
		if err := decoder.Decode(&event); err != nil {
			return text.String(), passed, skipped, live, err
		}
		text.WriteString(event.Output)
		if event.Test == "" {
			continue
		}
		switch event.Action {
		case "pass":
			passed++
			if strings.HasPrefix(event.Test, "TestIntegration") {
				live++
			}
		case "skip":
			skipped++
		}
	}
	return text.String(), passed, skipped, live, nil
}
