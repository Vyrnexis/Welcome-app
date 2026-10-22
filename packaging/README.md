# Solus Packaging Guide

This directory contains `package.yml`, the recipe for building `solus-welcome` as an official Solus package (`.eopkg`) using `ypkg` and `solbuild`.

## Building the Package

1. **Tag and Release**: Once a release is tagged on GitHub (e.g., `v0.2.0`), GitHub creates the source tarball archive at:
   ```
   https://github.com/Vyrnexis/welcome-app/archive/refs/tags/v0.2.0.tar.gz
   ```

2. **Compute Checksum**: Download or compute the SHA-256 hash of the release source archive:
   ```bash
   sha256sum v0.2.0.tar.gz
   ```

3. **Update `package.yml`**:
   - Replace the comment placeholder next to the archive URL with the computed SHA-256 hash.
   - Update `version` and `release` if packaging a new version.
   - Verify license field when a `LICENSE` file is committed.

4. **Build with `solbuild`**:
   In a Solus packaging environment:
   ```bash
   solbuild build package.yml
   ```

> [!NOTE]
> Solus packaging policy prefers vendored modules for offline, reproducible builds. In this draft recipe, `networking: yes` is enabled so that Go can download required modules during build.
