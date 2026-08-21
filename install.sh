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
prebuilt_available=false
source_available=false
if [[ -f "$APP_BIN" ]]; then
  prebuilt_available=true
fi
if [[ -d "$ROOT_DIR/src" && -f "$ROOT_DIR/go.mod" && -f "$ROOT_DIR/go.sum" ]]; then
  source_available=true
fi

# usage prints the supported non-interactive installation modes.
usage() {
  echo "Usage: $0 [--prebuilt|-p|--build|-b]"
}

mode=""
if [[ $# -gt 1 ]]; then
  usage >&2
  exit 1
fi

if [[ $# -eq 1 ]]; then
  case "$1" in
    --prebuilt|-p)
      mode="prebuilt"
      ;;
    --build|-b)
      mode="build"
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac
else
  if [[ "$prebuilt_available" == true && "$source_available" == true ]]; then
    echo "Please select an installation method:"
    echo "1) Install from pre-compiled binary"
    echo "2) Compile from source and install"
    read -r -p "Enter choice [1-2]: " choice

    case "$choice" in
      1)
        mode="prebuilt"
        ;;
      2)
        mode="build"
        ;;
      *)
        echo "Invalid choice. Exiting." >&2
        exit 1
        ;;
    esac
  elif [[ "$prebuilt_available" == true ]]; then
    mode="prebuilt"
  elif [[ "$source_available" == true ]]; then
    mode="build"
  else
    echo "Error: Neither a prebuilt binary nor source files are available." >&2
    exit 1
  fi
fi

case "$mode" in
  prebuilt)
    if [[ ! -f "$APP_BIN" ]]; then
      echo "Error: Prebuilt binary not found at $APP_BIN" >&2
      exit 1
    fi
    echo "Using existing prebuilt binary..."
    ;;
  build)
    if [[ "$source_available" != true ]]; then
      echo "Error: Source files are not included in this release package." >&2
      exit 1
    fi
    echo "Compiling from source..."
    mkdir -p "$BIN_DIR"
    go build -tags wayland -trimpath -ldflags="-s -w" -o "$APP_BIN" ./src
    ;;
esac

privilege=()
if [[ -z "$DESTDIR" && "${EUID:-$(id -u)}" -ne 0 ]]; then
  privilege=(sudo)
fi

# install_file installs a regular data file with standard read permissions.
install_file() {
  local source_path="$1"
  local destination_path="$2"
  "${privilege[@]}" install -Dm644 "$source_path" "$destination_path"
}

# render_desktop_file writes prefix-aware executable and icon paths to a desktop entry.
render_desktop_file() {
  local destination_path="$1"
  awk -v prefix="$PREFIX" '
    $0 == "Exec=/usr/bin/solus-welcome" {
      $0 = "Exec=" prefix "/bin/solus-welcome"
    }
    $0 == "Icon=/usr/share/solus-welcome/assets/logo.svg" {
      $0 = "Icon=" prefix "/share/solus-welcome/assets/logo.svg"
    }
    { print }
  ' "$ROOT_DIR/solus-welcome.desktop" > "$destination_path"
}

"${privilege[@]}" install -Dm755 "$APP_BIN" "$INSTALL_ROOT/bin/solus-welcome"
install_file "$ROOT_DIR/assets/config.toml" "$INSTALL_ROOT/share/solus-welcome/assets/config.toml"

while IFS= read -r -d '' directory; do
  relative_path="${directory#"$ROOT_DIR/assets/locales"}"
  "${privilege[@]}" install -d -m755 "$INSTALL_ROOT/share/solus-welcome/assets/locales$relative_path"
done < <(find "$ROOT_DIR/assets/locales" -type d -print0)

while IFS= read -r -d '' locale_file; do
  relative_path="${locale_file#"$ROOT_DIR/assets/locales/"}"
  install_file "$locale_file" "$INSTALL_ROOT/share/solus-welcome/assets/locales/$relative_path"
done < <(find "$ROOT_DIR/assets/locales" -type f -print0)

for svg in logo budgie gnome kde xfce; do
  install_file "$ROOT_DIR/assets/$svg.svg" "$INSTALL_ROOT/share/solus-welcome/assets/$svg.svg"
done

desktop_file="$(mktemp)"
trap 'rm -f -- "$desktop_file"' EXIT
render_desktop_file "$desktop_file"
install_file "$desktop_file" "$INSTALL_ROOT/share/applications/solus-welcome.desktop"

echo "Installed Solus Welcome to $INSTALL_ROOT/bin/solus-welcome"
