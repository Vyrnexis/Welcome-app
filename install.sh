#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$ROOT_DIR/bin"
APP_BIN="$BIN_DIR/solus-welcome"

mkdir -p "$BIN_DIR"
go build -ldflags="-s -w" -o "$APP_BIN" ./src

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
