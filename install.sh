#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$ROOT_DIR/bin"
APP_BIN="$BIN_DIR/solus-welcome"

echo "Please select an installation method:"
echo "1) Install from pre-compiled binary"
echo "2) Compile from source and install"
read -p "Enter choice [1-2]: " choice

case $choice in
  1)
    if [[ ! -f "$APP_BIN" ]]; then
      echo "Error: Prebuilt binary not found at $APP_BIN"
      exit 1
    fi
    echo "Using existing prebuilt binary..."
    ;;
  2)
    echo "Compiling from source..."
    mkdir -p "$BIN_DIR"
    go build -ldflags="-s -w" -o "$APP_BIN" ./src
    ;;
  *)
    echo "Invalid choice. Exiting."
    exit 1
    ;;
esac

if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
  install -Dm755 "$APP_BIN" /usr/bin/solus-welcome
  install -Dm644 "$ROOT_DIR/assets/logo.svg" /usr/share/solus-welcome/assets/logo.svg
  install -Dm644 "$ROOT_DIR/assets/budgie.svg" /usr/share/solus-welcome/assets/budgie.svg
  install -Dm644 "$ROOT_DIR/assets/gnome.svg" /usr/share/solus-welcome/assets/gnome.svg
  install -Dm644 "$ROOT_DIR/assets/kde.svg" /usr/share/solus-welcome/assets/kde.svg
  install -Dm644 "$ROOT_DIR/assets/xfce.svg" /usr/share/solus-welcome/assets/xfce.svg
  install -Dm644 "$ROOT_DIR/solus-welcome.desktop" /usr/share/applications/solus-welcome.desktop
else
  sudo install -Dm755 "$APP_BIN" /usr/bin/solus-welcome
  sudo install -Dm644 "$ROOT_DIR/assets/logo.svg" /usr/share/solus-welcome/assets/logo.svg
  sudo install -Dm644 "$ROOT_DIR/assets/budgie.svg" /usr/share/solus-welcome/assets/budgie.svg
  sudo install -Dm644 "$ROOT_DIR/assets/gnome.svg" /usr/share/solus-welcome/assets/gnome.svg
  sudo install -Dm644 "$ROOT_DIR/assets/kde.svg" /usr/share/solus-welcome/assets/kde.svg
  sudo install -Dm644 "$ROOT_DIR/assets/xfce.svg" /usr/share/solus-welcome/assets/xfce.svg
  sudo install -Dm644 "$ROOT_DIR/solus-welcome.desktop" /usr/share/applications/solus-welcome.desktop
fi

echo "Installed Solus Welcome to /usr/bin/solus-welcome"
