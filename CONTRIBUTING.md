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
Translations are plain JSON catalogues using Fyne's built-in i18n, embedded into the binary at build time. The message ID is the English text from `assets/config.toml`, so you never copy commands or URLs.

Please see the **[Translation Guide](assets/translations/README.md)** for how to add your language.

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
Assets and translations are embedded in the binary. To try a different configuration, place it at `/etc/solus-welcome/config.toml`; it replaces the embedded one if it is valid.

**Running the Test Suite:**
We use Fyne's headless test framework and standard Go testing.
```bash
go test -v -race ./...
```

**Compiling a Release Binary:**
To build a fully stripped and optimized release binary:
```bash
go build -ldflags="-s -w" -trimpath -o bin/solus-welcome ./src
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
- **Comments:** Explain *why*, not what. Skip comments that only restate a function's name; document non-obvious behaviour, constraints, and trade-offs.
- **Registries:** Pages (`pages.go`) and welcome-card actions (`actions.go`) are registered in one map each; configuration validation derives its known keys from them.
- **Commits:** We follow Conventional Commits (e.g., `feat: add new button`, `fix: resolve crash on startup`, `docs: update readme`).
