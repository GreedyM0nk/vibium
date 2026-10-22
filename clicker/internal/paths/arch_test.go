package paths

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func foreignArch() string {
	if runtime.GOARCH == "arm64" {
		return "amd64"
	}
	return "arm64"
}

// buildTinyBinary cross-compiles an empty program for goarch (host GOOS) to
// outPath, giving a real executable of a known architecture to match against.
func buildTinyBinary(t *testing.T, goarch, outPath string) {
	t.Helper()
	src := filepath.Join(t.TempDir(), "main.go")
	if err := os.WriteFile(src, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", outPath, src)
	cmd.Env = append(os.Environ(), "GOOS="+runtime.GOOS, "GOARCH="+goarch, "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build GOARCH=%s: %v\n%s", goarch, err, out)
	}
}

func TestBinaryArchMatches(t *testing.T) {
	dir := t.TempDir()

	host := filepath.Join(dir, "host")
	buildTinyBinary(t, runtime.GOARCH, host)
	if !binaryArchMatches(host) {
		t.Errorf("host-arch (%s) binary should match", runtime.GOARCH)
	}

	foreign := filepath.Join(dir, "foreign")
	buildTinyBinary(t, foreignArch(), foreign)
	if binaryArchMatches(foreign) {
		t.Errorf("foreign-arch (%s) binary should not match on %s", foreignArch(), runtime.GOARCH)
	}

	// An unparseable file is treated as matching: the stat checks gate
	// existence, and a parser quirk must not discard a valid install.
	junk := filepath.Join(dir, "junk")
	if err := os.WriteFile(junk, []byte("not an executable"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !binaryArchMatches(junk) {
		t.Error("unparseable file should be treated as matching")
	}
}

// A cached pin whose binaries are the wrong architecture must be rejected so
// ensure-install downloads the right build over it instead of launching into
// "exec format error" (#699).
func TestResolveVersionDirRejectsWrongArch(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("VIBIUM_CACHE_DIR", cache)
	t.Setenv("VIBIUM_ENGINE_CHANNEL", "")
	t.Setenv("VIBIUM_ENGINE_VERSION", "150.0.7000.10")

	dir := filepath.Join(cache, "chrome-for-testing", "150.0.7000.10")
	buildTinyBinary(t, foreignArch(), getChromePathInVersion(dir))
	buildTinyBinary(t, foreignArch(), getChromedriverPathInVersion(dir))

	if _, err := resolveVersionDir(""); err == nil {
		t.Error("resolveVersionDir() with a wrong-arch cached pin: want error, got nil")
	}

	// Rebuilt for the host arch, the same pin resolves.
	buildTinyBinary(t, runtime.GOARCH, getChromePathInVersion(dir))
	buildTinyBinary(t, runtime.GOARCH, getChromedriverPathInVersion(dir))
	if _, err := resolveVersionDir(""); err != nil {
		t.Errorf("resolveVersionDir() with a host-arch cached pin: %v", err)
	}
}
