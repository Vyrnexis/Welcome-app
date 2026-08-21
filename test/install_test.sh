#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_DESTDIR="$(mktemp -d)"
PREFIX="${PREFIX:-/usr}"

cleanup() {
  rm -rf -- "$TEST_DESTDIR"
}
trap cleanup EXIT

fail() {
  echo "install test failed: $*" >&2
  exit 1
}

assert_mode() {
  local expected_mode="$1"
  local path="$2"
  [[ -e "$path" ]] || fail "missing $path"
  local actual_mode
  actual_mode="$(stat -c '%a' "$path")"
  [[ "$actual_mode" == "$expected_mode" ]] || fail "$path has mode $actual_mode, want $expected_mode"
}

DESTDIR="$TEST_DESTDIR" PREFIX="$PREFIX" "$ROOT_DIR/install.sh" --prebuilt

INSTALL_ROOT="${TEST_DESTDIR}${PREFIX}"
assert_mode 755 "$INSTALL_ROOT/bin/solus-welcome"
assert_mode 755 "$INSTALL_ROOT/share/solus-welcome"
assert_mode 755 "$INSTALL_ROOT/share/solus-welcome/assets"
assert_mode 755 "$INSTALL_ROOT/share/solus-welcome/assets/locales"
assert_mode 644 "$INSTALL_ROOT/share/solus-welcome/assets/config.toml"

for svg in logo budgie gnome kde xfce; do
  assert_mode 644 "$INSTALL_ROOT/share/solus-welcome/assets/$svg.svg"
done

while IFS= read -r -d '' directory; do
  relative_path="${directory#"$ROOT_DIR/assets/locales"}"
  assert_mode 755 "$INSTALL_ROOT/share/solus-welcome/assets/locales$relative_path"
done < <(find "$ROOT_DIR/assets/locales" -type d -print0)

while IFS= read -r -d '' locale_file; do
  relative_path="${locale_file#"$ROOT_DIR/assets/locales/"}"
  assert_mode 644 "$INSTALL_ROOT/share/solus-welcome/assets/locales/$relative_path"
done < <(find "$ROOT_DIR/assets/locales" -type f -print0)

desktop_file="$INSTALL_ROOT/share/applications/solus-welcome.desktop"
assert_mode 644 "$desktop_file"
if command -v desktop-file-validate >/dev/null 2>&1; then
  desktop-file-validate "$desktop_file"
fi

echo "install test passed"
