#!/usr/bin/env bash
# Manual verification for #638: linux-arm64 (Raspberry Pi, aarch64 Debian) must
# use a system chromium + chromedriver instead of downloading an x86 Chrome for
# Testing that dies with "exec format error". CI is x86 Linux and cannot
# exercise this; run it on the aarch64 hardware.
#
# PASS when, with system chromium + chromedriver installed, a navigation works;
# and when they are absent, the error names the architecture and the fix rather
# than failing with a bare exec error.
#
# Usage on the Pi (from a built checkout, or with vibium on PATH):
#   ./clicker/internal/browser/testdata/verify-linux-arm64-chrome.sh ./clicker/bin/vibium
set -u
VIBIUM="${1:-vibium}"

arch="$(uname -m)"
if [ "$arch" != "aarch64" ] && [ "$arch" != "arm64" ]; then
  echo "This check is for aarch64 Linux; here uname -m = $arch. Skipping."
  exit 2
fi

echo "=== 1. system browser present? ==="
have_chromium=0; have_driver=0
command -v chromium >/dev/null 2>&1 || command -v chromium-browser >/dev/null 2>&1 && have_chromium=1
command -v chromedriver >/dev/null 2>&1 && have_driver=1
echo "chromium=$have_chromium chromedriver=$have_driver"

if [ "$have_chromium" = 1 ] && [ "$have_driver" = 1 ]; then
  echo "=== 2. navigation should work (no download, no exec format error) ==="
  if out="$("$VIBIUM" --headless go https://example.com 2>&1)"; then
    echo "$out" | grep -qi "example" && echo "PASS: navigated using the system browser" || {
      echo "FAIL: go succeeded but output looks wrong:"; echo "$out"; exit 1; }
  else
    echo "FAIL: go failed with system browser installed:"; echo "$out"; exit 1
  fi
else
  echo "=== 2. with the system browser absent, the error must name the cause ==="
  out="$("$VIBIUM" --headless go https://example.com 2>&1)"; rc=$?
  echo "$out"
  if [ "$rc" = 0 ]; then echo "FAIL: expected an error without a system browser"; exit 1; fi
  echo "$out" | grep -qiE "no linux-arm64 build|apt install chromium" \
    && echo "PASS: error names the architecture and the fix" \
    || { echo "FAIL: error did not explain the architecture cause"; exit 1; }
  echo "(install chromium + chromium-driver and rerun to check the happy path)"
fi
