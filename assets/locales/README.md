# Solus Welcome Translations

This directory contains community-provided translations for the Solus Welcome application.

## How to add a translation
1. Copy the main `../config.toml` file into this directory and rename it to your language code (e.g., `fr.toml`, `de.toml`, `es.toml`).
2. Delete any sections (like desktop commands) that do not need translating, to keep the file clean.
3. Translate the remaining text strings (especially the `[UI]` section) into your target language.
4. If there is a language-specific Solus community hub (like a localized forum), you can update the URLs in your translation file as well!

The application will automatically detect the user's system language and load your translated file over the default English configuration.
