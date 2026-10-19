package paths

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Google publishes Chrome for Testing for linux64, mac-x64, mac-arm64, win32,
// and win64 only — there is no linux-arm64 build. On a Raspberry Pi and other
// aarch64 Linux the downloaded linux64 binary is x86, so every launch died
// with a bare "exec format error" that named nothing (#638). Debian and
// Raspberry Pi OS ship native, version-matched chromium and chromedriver, so
// Vibium uses those there instead of downloading.

// UseSystemChrome reports whether this platform has no Chrome for Testing
// download and must fall back to a system chromium + chromedriver.
func UseSystemChrome() bool {
	return runtime.GOOS == "linux" && runtime.GOARCH == "arm64"
}

// Candidate executable names for a system Chrome/Chromium and its driver.
// Debian and Raspberry Pi OS use "chromium" + "chromedriver"; other distros
// and manual installs vary, so a few common names are tried in order.
var systemChromeNames = []string{"chromium", "chromium-browser", "google-chrome-stable", "google-chrome"}
var systemChromedriverNames = []string{"chromedriver"}

// lookPath is indirected so tests can resolve against a controlled PATH
// without depending on what the test host happens to have installed.
var lookPath = exec.LookPath

// findSystemBinary returns the first of names found on PATH, or "".
func findSystemBinary(names []string) string {
	for _, name := range names {
		if p, err := lookPath(name); err == nil {
			return p
		}
	}
	return ""
}

// SystemChromeExecutable returns the system Chrome/Chromium path, or an error
// naming the architecture and the fix when none is installed.
func SystemChromeExecutable() (string, error) {
	if p := findSystemBinary(systemChromeNames); p != "" {
		return p, nil
	}
	return "", systemChromeMissingError("chromium")
}

// SystemChromedriverPath returns the system chromedriver path, or an error.
func SystemChromedriverPath() (string, error) {
	if p := findSystemBinary(systemChromedriverNames); p != "" {
		return p, nil
	}
	return "", systemChromeMissingError("chromedriver")
}

// systemChromeMissingError explains why there is nothing to run and how to fix
// it, instead of the old "exec format error" that gave no clue the cause was
// architecture.
func systemChromeMissingError(what string) error {
	return fmt.Errorf(
		"%s not found: Chrome for Testing has no %s-%s build, so Vibium uses a system browser here; install one with `apt install chromium chromium-driver` (or set VIBIUM_ENGINE_PATH)",
		what, runtime.GOOS, runtime.GOARCH,
	)
}
