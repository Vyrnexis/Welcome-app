# Solus Welcome

[![CI/CD Pipeline](https://github.com/Vyrnexis/welcome-app/actions/workflows/ci.yml/badge.svg)](https://github.com/Vyrnexis/welcome-app/actions/workflows/ci.yml)

Solus Welcome is a native desktop application that helps new and returning Solus users inspect their system, install updates, open desktop settings, discover software, and find official community resources.

The application is written in Go with the [Fyne](https://fyne.io/) GUI toolkit and supports the Budgie, GNOME, KDE Plasma, and Xfce editions of Solus.

<p align="center">
  <img src="assets/Screenshot.png" alt="Solus Welcome application showing system information and getting-started actions" width="800">
</p>

## Features

- Detects the active Budgie, GNOME, KDE Plasma, or Xfce desktop session.
- Displays the desktop edition, architecture, Linux kernel, and installed memory.
- Checks for available `eopkg` upgrades without blocking the interface.
- Opens the appropriate software centre, appearance tools, and system settings for the active desktop.
- Provides direct links to Solus documentation, support channels, contribution guides, and donations.
- Manages the per-user desktop autostart entry.
- Includes light and dark application themes.
- Loads navigation, actions, links, and interface text from TOML configuration.
- Supports regional and base-language translation fallbacks.
- Installs from a release binary or builds an optimized binary from source.

## Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| `F1` | Open Welcome |
| `F2` | Open Customise |
| `F3` | Open Support |
| `F4` | Open Contribute |
| `Esc` | Close the application |

## Installation

### Install a published release

Download the Linux x86-64 archive and `checksums.txt` from the [Releases page](https://github.com/Vyrnexis/welcome-app/releases). Keep both files in the same directory and verify the archive before extracting it:

```bash
sha256sum --check checksums.txt
```

Extract the archive into an empty directory, enter that directory, and install its prebuilt binary:

```bash
./install.sh --prebuilt
```

The release archive also contains `src/`, `go.mod`, and `go.sum`, so it can build and install directly from source:

```bash
./install.sh --build
```

Running `./install.sh` without an option presents an interactive choice between these modes.

### Install from a repository checkout

```bash
git clone https://github.com/Vyrnexis/welcome-app.git
cd welcome-app
./install.sh --build
```

Use `./install.sh --prebuilt` instead when installing the optimized binary already stored in `bin/solus-welcome`.

### Installation paths

The default installation uses `/usr` and places files under:

- `/usr/bin/solus-welcome`
- `/usr/share/solus-welcome/`
- `/usr/share/applications/solus-welcome.desktop`

The installer requests `sudo` only when installing directly into the system and the current process is not root.

An alternate absolute prefix is supported:

```bash
PREFIX=/opt/solus-welcome ./install.sh --prebuilt
```

Packagers can stage an installation without `sudo` by setting `DESTDIR`:

```bash
DESTDIR=/tmp/solus-welcome-package PREFIX=/usr ./install.sh --prebuilt
```

The generated desktop entry and runtime asset discovery follow the selected prefix.

## Build requirements

Installing a prebuilt release does not require the Go toolchain. Building from source requires:

- Go at the version declared in `go.mod`.
- A C compiler and development tools.
- OpenGL, X11, Wayland, and XKB development headers required by Fyne.

On Solus, install the development dependencies with:

```bash
sudo eopkg install -c system.devel golang mesalib-devel libxrandr-devel libxcursor-devel libxi-devel libxinerama-devel wayland-devel libxkbcommon-devel libxxf86vm-devel
```

## Development

Run the application directly from the repository root:

```bash
go run ./src
```

Format, analyze, and test the project:

```bash
gofmt -s -w .
go vet ./...
go test -v -race ./...
./test/install_test.sh
```

Build the fully stripped Wayland release binary:

```bash
mkdir -p bin
go build -tags wayland -trimpath -ldflags="-s -w" -o bin/solus-welcome ./src
```

Create the optional UPX/LZMA-compressed variant:

```bash
cp bin/solus-welcome bin/solus-welcome-upx-compressed
upx --best --lzma bin/solus-welcome-upx-compressed
```

## Configuration

The base configuration is [assets/config.toml](assets/config.toml). It defines:

- Sidebar navigation and welcome cards.
- Desktop-specific software and settings commands.
- Support and contribution links.
- User-interface text and format strings.

Installed systems load the file from the selected prefix under `share/solus-welcome/assets/config.toml`. Changes take effect the next time the application starts; recompilation is not required.

Commands in this file are executed as argument arrays rather than being concatenated into a shell command. Preserve `%d` and `%s` placeholders when editing translated format strings.

## Translations

Translations live in `assets/locales/` as TOML overrides. Locale matching tries a regional file first, then its base language, and finally retains the English values from `assets/config.toml`.

For example, `pt_BR.UTF-8` is resolved in this order:

1. `assets/locales/pt_BR.toml`
2. `assets/locales/pt.toml`
3. `assets/config.toml`

Malformed regional translations are reported while the loader continues to a valid base-language fallback. See the [translation guide](assets/locales/README.md) for naming rules and an example.

## Project layout

| Path | Purpose |
| --- | --- |
| `src/` | Application source and Go tests |
| `assets/config.toml` | Base content and desktop command configuration |
| `assets/locales/` | Translation overrides and translation documentation |
| `assets/*.svg` | Solus and desktop-environment artwork |
| `bin/` | Optimized release binaries |
| `test/install_test.sh` | Default-prefix and custom-prefix staged installer tests |
| `install.sh` | Interactive and non-interactive installer |
| `.github/workflows/ci.yml` | Formatting, analysis, tests, build, packaging, and releases |

## CI and releases

Every push to `main` and every pull request targeting `main` runs formatting validation, `go vet`, race-enabled Go tests, the stripped Wayland build, and staged installer tests.

Tags beginning with `v` run a separate release job with repository write permission. That job installs UPX, compresses the tested build artifact, verifies the archive contents, generates a SHA-256 checksum, and publishes the GitHub release.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md) before submitting code. Translation-only contributions should follow [assets/locales/README.md](assets/locales/README.md).

Use Conventional Commits for commit messages, keep functions documented, and do not commit changes that fail formatting, analysis, race tests, or installer tests.
