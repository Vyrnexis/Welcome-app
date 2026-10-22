#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_ROOT="$(mktemp -d)"

created_dummy=false
if [[ ! -f "$ROOT_DIR/bin/solus-welcome" ]]; then
  mkdir -p "$ROOT_DIR/bin"
  cat <<'EOF' > "$ROOT_DIR/bin/solus-welcome"
#!/bin/sh
exit 0
EOF
  chmod 755 "$ROOT_DIR/bin/solus-welcome"
  created_dummy=true
fi

# cleanup removes temporary test directories and the dummy binary if created.
cleanup() {
  rm -rf -- "$TEST_ROOT"
  if [[ "$created_dummy" == true ]]; then
    rm -f -- "$ROOT_DIR/bin/solus-welcome"
    rmdir "$ROOT_DIR/bin" 2>/dev/null || true
  fi
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

# assert_not_exists verifies that a path does not exist.
assert_not_exists() {
  local path="$1"
  if [[ -e "$path" || -L "$path" ]]; then
    fail "path should not exist: $path"
  fi
}

# Test invalid arguments
test_invalid_arguments() {
  local test_destdir
  test_destdir="$(mktemp -d -p "$TEST_ROOT")"

  if "$ROOT_DIR/install.sh" --invalid-arg >/dev/null 2>&1; then
    fail "expected failure for unknown argument"
  fi

  if "$ROOT_DIR/install.sh" --build --prebuilt >/dev/null 2>&1; then
    fail "expected failure for multiple arguments"
  fi

  if DESTDIR="$test_destdir" PREFIX="relative/prefix" "$ROOT_DIR/install.sh" --prebuilt >/dev/null 2>&1; then
    fail "expected failure for relative PREFIX"
  fi

  if DESTDIR="$test_destdir" PREFIX="/opt/solus welcome" "$ROOT_DIR/install.sh" --prebuilt >/dev/null 2>&1; then
    fail "expected failure for whitespace in PREFIX"
  fi

  if DESTDIR="$test_destdir" PREFIX="/usr" SYSCONFDIR="relative/etc" "$ROOT_DIR/install.sh" --prebuilt >/dev/null 2>&1; then
    fail "expected failure for relative SYSCONFDIR"
  fi
}

# Test staged install and uninstall for PREFIX=/usr
test_prefix_usr() {
  local test_destdir
  test_destdir="$(mktemp -d -p "$TEST_ROOT")"

  DESTDIR="$test_destdir" PREFIX="/usr" "$ROOT_DIR/install.sh" --prebuilt

  assert_mode 755 "$test_destdir/usr/bin/solus-welcome"
  assert_mode 644 "$test_destdir/usr/share/applications/us.getsol.SolusWelcome.desktop"
  assert_mode 644 "$test_destdir/usr/share/icons/hicolor/scalable/apps/us.getsol.SolusWelcome.svg"
  assert_mode 644 "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop"

  assert_not_exists "$test_destdir/usr/share/solus-welcome"

  grep -Fqx "Exec=solus-welcome" "$test_destdir/usr/share/applications/us.getsol.SolusWelcome.desktop" || fail "Exec not solus-welcome in usr app desktop"
  grep -Fqx "Exec=solus-welcome" "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop" || fail "Exec not solus-welcome in usr autostart desktop"

  if command -v desktop-file-validate >/dev/null 2>&1; then
    desktop-file-validate "$test_destdir/usr/share/applications/us.getsol.SolusWelcome.desktop"
    desktop-file-validate "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop"
  fi

  # Uninstall and verify removal
  DESTDIR="$test_destdir" PREFIX="/usr" "$ROOT_DIR/install.sh" --uninstall

  assert_not_exists "$test_destdir/usr/bin/solus-welcome"
  assert_not_exists "$test_destdir/usr/share/applications/us.getsol.SolusWelcome.desktop"
  assert_not_exists "$test_destdir/usr/share/icons/hicolor/scalable/apps/us.getsol.SolusWelcome.svg"
  assert_not_exists "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop"
}

# Test staged install and uninstall for PREFIX=/opt/solus-welcome/ with trailing slash
test_prefix_custom_opt() {
  local test_destdir
  test_destdir="$(mktemp -d -p "$TEST_ROOT")"

  DESTDIR="$test_destdir" PREFIX="/opt/solus-welcome/" "$ROOT_DIR/install.sh" --prebuilt

  assert_mode 755 "$test_destdir/opt/solus-welcome/bin/solus-welcome"
  assert_mode 644 "$test_destdir/opt/solus-welcome/share/applications/us.getsol.SolusWelcome.desktop"
  assert_mode 644 "$test_destdir/opt/solus-welcome/share/icons/hicolor/scalable/apps/us.getsol.SolusWelcome.svg"
  assert_mode 644 "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop"

  assert_not_exists "$test_destdir/opt/solus-welcome/share/solus-welcome"

  grep -Fqx "Exec=/opt/solus-welcome/bin/solus-welcome" "$test_destdir/opt/solus-welcome/share/applications/us.getsol.SolusWelcome.desktop" || fail "Exec not absolute in opt app desktop"
  grep -Fqx "Exec=/opt/solus-welcome/bin/solus-welcome" "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop" || fail "Exec not absolute in opt autostart desktop"

  if command -v desktop-file-validate >/dev/null 2>&1; then
    desktop-file-validate "$test_destdir/opt/solus-welcome/share/applications/us.getsol.SolusWelcome.desktop"
    desktop-file-validate "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop"
  fi

  # Create legacy files to verify they are cleaned up on uninstall
  mkdir -p "$test_destdir/opt/solus-welcome/share/solus-welcome"
  touch "$test_destdir/opt/solus-welcome/share/solus-welcome/legacy.txt"
  touch "$test_destdir/opt/solus-welcome/share/applications/solus-welcome.desktop"

  # Uninstall and verify removal of all 4 files plus legacy files
  DESTDIR="$test_destdir" PREFIX="/opt/solus-welcome/" "$ROOT_DIR/install.sh" --uninstall

  assert_not_exists "$test_destdir/opt/solus-welcome/bin/solus-welcome"
  assert_not_exists "$test_destdir/opt/solus-welcome/share/applications/us.getsol.SolusWelcome.desktop"
  assert_not_exists "$test_destdir/opt/solus-welcome/share/icons/hicolor/scalable/apps/us.getsol.SolusWelcome.svg"
  assert_not_exists "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop"
  assert_not_exists "$test_destdir/opt/solus-welcome/share/solus-welcome"
  assert_not_exists "$test_destdir/opt/solus-welcome/share/applications/solus-welcome.desktop"
}

# Test release package without source (defaulting to --prebuilt and using data/ icon)
test_release_bundle() {
  local bundle_dir="$TEST_ROOT/solus-welcome-v0.2.0"
  local test_destdir="$TEST_ROOT/bundle-install"
  mkdir -p "$bundle_dir/bin" "$bundle_dir/data" "$test_destdir"

  cp "$ROOT_DIR/bin/solus-welcome" "$bundle_dir/bin/"
  cp "$ROOT_DIR/data/us.getsol.SolusWelcome.desktop" "$bundle_dir/data/"
  cp "$ROOT_DIR/data/us.getsol.SolusWelcome-autostart.desktop" "$bundle_dir/data/"
  cp "$ROOT_DIR/assets/logo.svg" "$bundle_dir/data/us.getsol.SolusWelcome.svg"
  cp "$ROOT_DIR/install.sh" "$bundle_dir/"
  cp "$ROOT_DIR/README.md" "$bundle_dir/"

  # Running without arguments in a release bundle should default to --prebuilt
  DESTDIR="$test_destdir" PREFIX="/usr" "$bundle_dir/install.sh"

  assert_mode 755 "$test_destdir/usr/bin/solus-welcome"
  assert_mode 644 "$test_destdir/usr/share/applications/us.getsol.SolusWelcome.desktop"
  assert_mode 644 "$test_destdir/usr/share/icons/hicolor/scalable/apps/us.getsol.SolusWelcome.svg"
  assert_mode 644 "$test_destdir/etc/xdg/autostart/us.getsol.SolusWelcome.desktop"

  if DESTDIR="$test_destdir" PREFIX="/usr" "$bundle_dir/install.sh" --build >/dev/null 2>&1; then
    fail "source-free release bundle unexpectedly accepted --build"
  fi
}

test_invalid_arguments
test_prefix_usr
test_prefix_custom_opt
test_release_bundle

echo "install test passed"
