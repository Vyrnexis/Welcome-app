#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_ROOT="$(mktemp -d)"

# cleanup removes all staged installation trees created by this test.
cleanup() {
  rm -rf -- "$TEST_ROOT"
}
trap cleanup EXIT

# fail prints an installation assertion failure and exits the test.
fail() {
  echo "install test failed: $*" >&2
  exit 1
}

# assert_mode verifies that a staged path exists with the expected permissions.
assert_mode() {
  local expected_mode="$1"
  local path="$2"
  [[ -e "$path" ]] || fail "missing $path"
  local actual_mode
  actual_mode="$(stat -c '%a' "$path")"
  [[ "$actual_mode" == "$expected_mode" ]] || fail "$path has mode $actual_mode, want $expected_mode"
}

# run_install_test verifies staged files, permissions, and prefix-aware desktop paths.
run_install_test() {
  local prefix="$1"
  local test_destdir
  local install_root
  local normalized_prefix="${prefix%/}"
  local relative_path
  local desktop_file
  test_destdir="$(mktemp -d -p "$TEST_ROOT")"

  DESTDIR="$test_destdir" PREFIX="$prefix" "$ROOT_DIR/install.sh" --prebuilt
  install_root="${test_destdir}${normalized_prefix}"

  assert_mode 755 "$install_root/bin/solus-welcome"
  assert_mode 755 "$install_root/share/solus-welcome"
  assert_mode 755 "$install_root/share/solus-welcome/assets"
  assert_mode 755 "$install_root/share/solus-welcome/assets/locales"
  assert_mode 644 "$install_root/share/solus-welcome/assets/config.toml"

  for svg in logo budgie gnome kde xfce; do
    assert_mode 644 "$install_root/share/solus-welcome/assets/$svg.svg"
  done

  while IFS= read -r -d '' directory; do
    relative_path="${directory#"$ROOT_DIR/assets/locales"}"
    assert_mode 755 "$install_root/share/solus-welcome/assets/locales$relative_path"
  done < <(find "$ROOT_DIR/assets/locales" -type d -print0)

  while IFS= read -r -d '' locale_file; do
    relative_path="${locale_file#"$ROOT_DIR/assets/locales/"}"
    assert_mode 644 "$install_root/share/solus-welcome/assets/locales/$relative_path"
  done < <(find "$ROOT_DIR/assets/locales" -type f -print0)

  desktop_file="$install_root/share/applications/solus-welcome.desktop"
  assert_mode 644 "$desktop_file"
  grep -Fqx "Exec=$normalized_prefix/bin/solus-welcome" "$desktop_file" || fail "incorrect Exec path for $prefix"
  grep -Fqx "Icon=$normalized_prefix/share/solus-welcome/assets/logo.svg" "$desktop_file" || fail "incorrect Icon path for $prefix"
  if command -v desktop-file-validate >/dev/null 2>&1; then
    desktop-file-validate "$desktop_file"
  fi
}

run_install_test "/usr"
run_install_test "/opt/solus-welcome"

echo "install test passed"
