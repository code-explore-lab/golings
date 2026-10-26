package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSummarizeDistinguishesTestsFromPackages(t *testing.T) {
	data := "{\"Action\":\"output\",\"Output\":\"diagnostic\\n\"}\n" +
		"{\"Action\":\"pass\",\"Package\":\"lesson\"}\n" +
		"{\"Action\":\"pass\",\"Test\":\"TestOne\"}\n" +
		"{\"Action\":\"pass\",\"Test\":\"TestIntegrationLive\"}\n" +
		"{\"Action\":\"skip\",\"Test\":\"TestMissingService\"}\n"
	out, passed, skipped, live, err := summarize([]byte(data))
	if err != nil || out != "diagnostic\n" || passed != 2 || skipped != 1 || live != 1 {
		t.Fatalf("%q %d %d %d %v", out, passed, skipped, live, err)
	}
	_, passed, _, _, err = summarize([]byte("{\"Action\":\"pass\",\"Package\":\"no-tests\"}\n"))
	if err != nil || passed != 0 {
		t.Fatalf("compile-only package counted as test: %d %v", passed, err)
	}
}

func TestGoTestIgnoresSavedTestFilters(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":           "module savedfilter\n\ngo 1.23\n",
		"required_test.go": "package savedfilter\nimport \"testing\"\nfunc TestRequiredCheck(t *testing.T) {}\n",
		"goenv":            "GOFLAGS=-run=DoesNotExist\n",
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GOENV", filepath.Join(dir, "goenv"))
	t.Setenv("GOFLAGS", "")
	out, err := GoTest(context.Background(), dir, false)
	if err != nil || !strings.Contains(out, "TestRequiredCheck") {
		t.Fatalf("saved GOFLAGS hid required checks: %v\n%s", err, out)
	}
}

func TestGoTestRejectsIncompleteAcceptance(t *testing.T) {
	for _, test := range []struct {
		name, code, want string
		integration      bool
	}{
		{"no-tests", "package lesson\n", "0 passed tests", false},
		{"skipped", "package lesson\nimport \"testing\"\nfunc TestRequired(t *testing.T) { t.Skip(\"service unavailable\") }\n", "1 skipped tests", false},
		{"missing-live-check", "package lesson\nimport \"testing\"\nfunc TestLocal(t *testing.T) {}\n", "no TestIntegration tests passed", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, data := range map[string]string{"go.mod": "module lesson\n\ngo 1.23\n", "lesson_test.go": test.code} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			out, err := GoTest(context.Background(), dir, test.integration)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("incomplete checks accepted: %v\n%s", err, out)
			}
		})
	}
}

func TestGoTestHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GoTest(ctx, t.TempDir(), false); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v; want cancellation", err)
	}
}

func TestTestEnvironmentRemovesFiltersAndWorkspaces(t *testing.T) {
	t.Setenv("GOFLAGS", "-run=DoesNotExist")
	t.Setenv("GOWORK", "some-other-module")
	values := strings.Join(testEnv(true), "\n")
	if strings.Contains(values, "DoesNotExist") || strings.Contains(values, "some-other-module") ||
		!strings.Contains(values, "GOWORK=off") || !strings.Contains(values, "GOPATH_INTEGRATION=1") {
		t.Fatalf("unexpected environment: %s", values)
	}
}
