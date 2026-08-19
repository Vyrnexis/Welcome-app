# Solus Welcome App

[![Build & Test](https://github.com/Vyrnexis/welcome-app/actions/workflows/ci.yml/badge.svg)](https://github.com/Vyrnexis/welcome-app/actions)

A lightweight Go-based welcome application for SolusOS.

This application provides users with a SolusOS welcome app and is built using the [Fyne](https://fyne.io/) GUI toolkit for Go, ensuring it is both exceptionally fast and seamlessly integrated.

<p align="center">
  <img src="assets/Screenshot.png" alt="Solus Welcome App Screenshot" width="800">
</p>

## Features

- **Smart Environment Detection**: Automatically creates shortcuts and settings based on your active desktop (Budgie, GNOME, KDE, XFCE).
- **System Updates**: Checks for available packages and lets you upgrade your system with a single click.
- **Quick Launcher**: Easy access to open your software center, system settings, or Solus documentation.
- **User Preferences**: Includes a built-in dark mode toggle and options to manage autostart behavior on boot.
- **Highly Configurable**: Every UI string, URL, and system command is dynamically loaded from a clean `config.toml` file, making it incredibly easy to tweak without recompiling.
- **Full i18n Support**: Automatically detects your system language and loads community-provided translations seamlessly.

## Installation

Pre-compiled, highly optimized binaries can be found in the `bin/` directory of this repository.

You can easily install the application using the provided install script:

```bash
./install.sh
```

Running the script presents a quick interactive menu:
1. **Install from pre-compiled binary**: Installs the existing binary directly without needing the Go toolchain.
2. **Compile from source and install**: Automatically builds a stripped, optimized Go binary and installs it.

Both methods will safely install the app and its associated assets (desktop entries, SVGs, and translations) into your system directories (`/usr/bin` and `/usr/share`).

## Configuration

If you wish to change the default links, modify the text, or adjust what terminal commands the buttons execute, simply edit the `assets/config.toml` file. The application parses this configuration at runtime!

## Contributing & Translations

We welcome all community contributions, whether it's fixing a bug, adding a feature, or translating the application into your native language!

- **Developers:** Please read our **[Contributing Guide](CONTRIBUTING.md)** for instructions on installing build dependencies (like Wayland/X11 headers), running the test suite, and our code formatting conventions.
- **Translators:** No programming experience is required to translate this app! Check out the **[Locales Guide](assets/locales/README.md)** to see how you can add your language in just a few minutes.
