#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$ROOT_DIR/bin"
APP_BIN="$BIN_DIR/solus-welcome"
DESTDIR="${DESTDIR:-}"
PREFIX="${PREFIX:-/usr}"
if [[ "$PREFIX" != /* || "$PREFIX" == *[$'\n\r\t ']* ]]; then
  echo "Error: PREFIX must be an absolute path without whitespace." >&2
  exit 1
fi
PREFIX="${PREFIX%/}"
INSTALL_ROOT="${DESTDIR}${PREFIX}"

SYSCONFDIR="${SYSCONFDIR:-/etc}"
if [[ "$SYSCONFDIR" != /* || "$SYSCONFDIR" == *[$'\n\r\t ']* ]]; then
  echo "Error: SYSCONFDIR must be an absolute path without whitespace." >&2
  exit 1
fi
SYSCONFDIR="${SYSCONFDIR%/}"

# BACKEND picks the windowing system compiled into a source build. GLFW, which Fyne uses, supports only one per
# binary: x11 also runs on Wayland sessions through XWayland, while wayland is native but cannot start on X11.
BACKEND="${BACKEND:-x11}"
case "$BACKEND" in
  x11) build_tags=() ;;
  wayland) build_tags=(-tags wayland) ;;
  *)
    echo "Error: BACKEND must be x11 or wayland." >&2
    exit 1
    ;;
esac

# usage prints the supported non-interactive installation and uninstallation modes.
usage() {
  echo "Usage: [BACKEND=x11|wayland] $0 [--build|-b|--prebuilt|-p|--uninstall|-u|-h|--help]"
}

mode=""
if [[ $# -gt 1 ]]; then
  usage >&2
  exit 1
fi

if [[ $# -eq 1 ]]; then
  case "$1" in
    --build|-b)
      mode="build"
      ;;
    --prebuilt|-p)
      mode="prebuilt"
      ;;
    --uninstall|-u)
      mode="uninstall"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac
else
  if [[ -d "$ROOT_DIR/src" && -f "$ROOT_DIR/go.mod" ]]; then
    mode="build"
  elif [[ -f "$APP_BIN" ]]; then
    mode="prebuilt"
  else
    echo "Error: Neither source files nor prebuilt binary found." >&2
    exit 1
  fi
fi

# needs_privilege reports whether writing target requires sudo for the current user.
needs_privilege() {
  local target="$1"
  if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
    return 1
  fi
  local dir
  dir="$(dirname "$target")"
  while [[ ! -d "$dir" && "$dir" != "/" ]]; do
    dir="$(dirname "$dir")"
  done
  if [[ ! -w "$dir" ]]; then
    return 0
  fi
  if [[ -e "$target" && ! -w "$target" ]]; then
    return 0
  fi
  return 1
}

# run_privileged runs a command with sudo only when target is not writable.
run_privileged() {
  local target="$1"
  shift
  if needs_privilege "$target"; then
    sudo "$@"
  else
    "$@"
  fi
}

remove_file() {
  local target="$1"
  if [[ -e "$target" || -L "$target" ]]; then
    run_privileged "$target" rm -f "$target"
  fi
}

remove_dir() {
  local target="$1"
  if [[ -d "$target" ]]; then
    run_privileged "$target" rm -rf "$target"
  fi
}

if [[ "$mode" == "uninstall" ]]; then
  echo "Uninstalling Solus Welcome..."
  remove_file "$INSTALL_ROOT/bin/solus-welcome"
  remove_file "$INSTALL_ROOT/share/applications/us.getsol.SolusWelcome.desktop"
  remove_file "$INSTALL_ROOT/share/icons/hicolor/scalable/apps/us.getsol.SolusWelcome.svg"
  remove_file "${DESTDIR}${SYSCONFDIR}/xdg/autostart/us.getsol.SolusWelcome.desktop"

  # Remove legacy files from older versions if present
  remove_file "$INSTALL_ROOT/share/applications/solus-welcome.desktop"
  remove_dir "$INSTALL_ROOT/share/solus-welcome"

  echo "Uninstalled Solus Welcome from $INSTALL_ROOT"
  exit 0
fi

case "$mode" in
  build)
    if [[ ! -d "$ROOT_DIR/src" || ! -f "$ROOT_DIR/go.mod" ]]; then
      echo "Error: Source files are not available for build mode." >&2
      exit 1
    fi
    echo "Compiling from source ($BACKEND backend)..."
    mkdir -p "$BIN_DIR"
    (cd "$ROOT_DIR" && go build "${build_tags[@]}" -trimpath -ldflags="-s -w" -o "$APP_BIN" ./src)
    ;;
  prebuilt)
    if [[ ! -f "$APP_BIN" ]]; then
      echo "Error: Prebuilt binary not found at $APP_BIN" >&2
      exit 1
    fi
    echo "Using existing prebuilt binary..."
    ;;
esac

icon_src=""
if [[ -f "$ROOT_DIR/data/us.getsol.SolusWelcome.svg" ]]; then
  icon_src="$ROOT_DIR/data/us.getsol.SolusWelcome.svg"
elif [[ -f "$ROOT_DIR/assets/logo.svg" ]]; then
  icon_src="$ROOT_DIR/assets/logo.svg"
else
  echo "Error: Application icon not found (checked data/us.getsol.SolusWelcome.svg and assets/logo.svg)." >&2
  exit 1
fi

install_desktop_file() {
  local src="$1"
  local dest="$2"
  if [[ "$PREFIX" == "/usr" ]]; then
    run_privileged "$dest" install -Dm644 "$src" "$dest"
  else
    local tmp
    tmp="$(mktemp)"
    if ! awk -v exec_bin="${PREFIX}/bin/solus-welcome" '
      $0 == "Exec=solus-welcome" { print "Exec=" exec_bin; next }
      { print }
    ' "$src" > "$tmp"; then
      rm -f "$tmp"
      return 1
    fi
    if ! run_privileged "$dest" install -Dm644 "$tmp" "$dest"; then
      rm -f "$tmp"
      return 1
    fi
    rm -f "$tmp"
  fi
}

echo "Installing Solus Welcome..."
run_privileged "$INSTALL_ROOT/bin/solus-welcome" install -Dm755 "$APP_BIN" "$INSTALL_ROOT/bin/solus-welcome"
install_desktop_file "$ROOT_DIR/data/us.getsol.SolusWelcome.desktop" "$INSTALL_ROOT/share/applications/us.getsol.SolusWelcome.desktop"
install_desktop_file "$ROOT_DIR/data/us.getsol.SolusWelcome-autostart.desktop" "${DESTDIR}${SYSCONFDIR}/xdg/autostart/us.getsol.SolusWelcome.desktop"
run_privileged "$INSTALL_ROOT/share/icons/hicolor/scalable/apps/us.getsol.SolusWelcome.svg" install -Dm644 "$icon_src" "$INSTALL_ROOT/share/icons/hicolor/scalable/apps/us.getsol.SolusWelcome.svg"

if [[ -z "$DESTDIR" ]]; then
  if command -v update-desktop-database >/dev/null 2>&1; then
    if needs_privilege "$PREFIX/share/applications"; then
      sudo update-desktop-database -q "$PREFIX/share/applications" >/dev/null 2>&1 || true
    else
      update-desktop-database -q "$PREFIX/share/applications" >/dev/null 2>&1 || true
    fi
  fi
  if command -v gtk-update-icon-cache >/dev/null 2>&1; then
    if needs_privilege "$PREFIX/share/icons/hicolor"; then
      sudo gtk-update-icon-cache -q -t -f "$PREFIX/share/icons/hicolor" >/dev/null 2>&1 || true
    else
      gtk-update-icon-cache -q -t -f "$PREFIX/share/icons/hicolor" >/dev/null 2>&1 || true
    fi
  fi
fi

echo "Installed Solus Welcome to $INSTALL_ROOT/bin/solus-welcome"
