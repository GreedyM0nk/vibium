package paths

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// findSystemBinary tries the candidate names in order and returns the first on
// PATH. The resolver is exercised directly (not gated on the arch) so it runs
// on any CI host, standing in for the arm64 lookup path (#638).
func TestFindSystemBinaryPrefersFirstFound(t *testing.T) {
	restore := lookPath
	defer func() { lookPath = restore }()

	lookPath = func(name string) (string, error) {
		if name == "chromium-browser" {
			return "/usr/bin/chromium-browser", nil
		}
		return "", errors.New("not found")
	}

	if got := findSystemBinary(systemChromeNames); got != "/usr/bin/chromium-browser" {
		t.Errorf("findSystemBinary = %q, want the one name that resolved", got)
	}
}

func TestFindSystemBinaryNoneFound(t *testing.T) {
	restore := lookPath
	defer func() { lookPath = restore }()
	lookPath = func(string) (string, error) { return "", exec.ErrNotFound }

	if got := findSystemBinary(systemChromeNames); got != "" {
		t.Errorf("findSystemBinary = %q, want empty when nothing resolves", got)
	}
}

// When no system browser is installed, the error must name the architecture
// and the fix, not leave the caller with a bare exec failure (#638).
func TestSystemChromeMissingErrorNamesArchAndFix(t *testing.T) {
	restore := lookPath
	defer func() { lookPath = restore }()
	lookPath = func(string) (string, error) { return "", exec.ErrNotFound }

	_, err := SystemChromeExecutable()
	if err == nil {
		t.Fatal("expected an error when no system chrome is installed")
	}
	msg := err.Error()
	for _, want := range []string{"Chrome for Testing has no", "apt install chromium"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q", msg, want)
		}
	}
}

// A present system chromedriver resolves to its path.
func TestSystemChromedriverResolves(t *testing.T) {
	restore := lookPath
	defer func() { lookPath = restore }()
	lookPath = func(name string) (string, error) {
		if name == "chromedriver" {
			return "/usr/bin/chromedriver", nil
		}
		return "", exec.ErrNotFound
	}

	got, err := SystemChromedriverPath()
	if err != nil || got != "/usr/bin/chromedriver" {
		t.Fatalf("SystemChromedriverPath = %q, %v; want /usr/bin/chromedriver", got, err)
	}
}
