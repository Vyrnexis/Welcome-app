# Contributing to Solus Welcome

First, thank you for taking the time to contribute! This document provides guidelines and instructions for developers, packagers, and translators who want to help improve the Solus Welcome app.

## Table of Contents
1. [Code of Conduct](#code-of-conduct)
2. [Translating the App (i18n)](#translating-the-app-i18n)
3. [Setting Up Your Development Environment](#setting-up-your-development-environment)
4. [Building and Testing](#building-and-testing)
5. [Code Style & Conventions](#code-style--conventions)

## Code of Conduct
Please ensure your interactions in issues and pull requests remain respectful and constructive, in line with the broader Solus community standards.

## Translating the App (i18n)
We highly encourage community translations! You do **not** need to be a Go programmer to translate this app. 
Our translations are managed entirely through simple TOML configuration files. 

Please see the dedicated **[Locales Guide](assets/locales/README.md)** for 4 easy steps on how to add your language.

## Setting Up Your Development Environment

If you want to modify the Go code or the UI, you will need the Go toolchain and the underlying C-headers required by the Fyne UI toolkit.

On Solus, you can install everything you need with a single command:
```bash
sudo eopkg install -c system.devel golang mesalib-devel libxrandr-devel libxcursor-devel libxi-devel libxinerama-devel wayland-devel libxkbcommon-devel libxxf86vm-devel
```

## Building and Testing

Once your dependencies are installed, you can build and test the application directly:

**Running the app locally (without installing):**
```bash
go run ./src
```

**Running the Test Suite:**
We use Fyne's headless test framework and standard Go testing.
```bash
go test -v ./src/...
```

**Compiling a Release Binary:**
To build a fully stripped and optimized release binary:
```bash
go build -tags wayland -ldflags="-s -w" -trimpath -o bin/solus-welcome ./src
```

**Compressing with UPX:**
To produce the lightweight standalone compressed binary:
```bash
cp bin/solus-welcome bin/solus-welcome-upx-compressed
upx --best --lzma bin/solus-welcome-upx-compressed
```

## Code Style & Conventions

We follow standard idiomatic Go conventions:
- **Formatting:** Always run `gofmt -s -w .` before committing. Our CI pipeline will fail if code is improperly formatted.
- **Linting:** We encourage running `go vet ./...` locally.
- **Documentation:** Every exported and unexported function MUST have a clean 1-to-2 line GoDoc comment directly above it explaining precisely what it does. Avoid cluttering the internal function bodies with obvious inline comments.
- **Commits:** We follow Conventional Commits (e.g., `feat: add new button`, `fix: resolve crash on startup`, `docs: update readme`).
