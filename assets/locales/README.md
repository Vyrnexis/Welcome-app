# Solus Welcome Translations

This directory contains community-provided translations for the Solus Welcome application.

## Naming Conventions
Translation files are loaded dynamically based on the system's `$LANG` environment variable:
- **Base Languages (ISO 639-1):** Use two-letter lowercase codes when a dialect distinction is unnecessary (e.g., `de.toml` for German, `fr.toml` for French, `es.toml` for Spanish, `it.toml` for Italian).
- **Regional Variants (ISO 639-1 + ISO 3166-1):** Use `<language>_<COUNTRY>.toml` when regional differences exist:
  - `zh_CN.toml`: Simplified Chinese (`zh` = Zhongwen / Chinese, `CN` = China)
  - `zh_TW.toml`: Traditional Chinese (`zh` = Zhongwen / Chinese, `TW` = Taiwan)
  - `pt_BR.toml`: Brazilian Portuguese (`pt` = Portuguese, `BR` = Brazil)
  - `pt_PT.toml`: European Portuguese (`pt` = Portuguese, `PT` = Portugal)

The application attempts to match regional files first (e.g., `pt_BR.toml`) before falling back to the base language (`pt.toml`), and finally defaults to English (`../config.toml`).

## How to Add a Translation
1. Create a new file named after your language code (e.g., `es.toml` or `pt_BR.toml`).
2. Only include keys that require translation (unspecified keys automatically retain the default English text).
3. Translate the strings, ensuring format placeholders (`%d`, `%s`) are preserved.

### Minimal Example (`de.toml`):
```toml
[UI]
Tagline = "Das persönliche Betriebssystem für persönliche Computer"
ShowOnStartup = "Diesen Willkommensbildschirm beim Start anzeigen"
Close = "Schließen"
DarkTheme = "Dunkles Design"
WelcomeTitle = "Willkommen bei Solus"
WelcomeSubtitle = "Vielen Dank, dass Sie sich für Solus entschieden haben."
GettingStarted = "Erste Schritte"
SolusLinux = "Solus Linux"
EditionSuffix = " Edition"
CheckForUpdates = "Nach Updates suchen"
CheckingUpdates = "Suche nach Updates..."
UnableToCheckUpdates = "Updates konnten nicht geprüft werden"
SystemUpToDate = "System ist auf dem neuesten Stand"
OneUpdateAvailable = "1 Update verfügbar"
UpdatesAvailable = "%d Updates verfügbar"
NoDesktopSettings = "Keine desktopspezifischen Einstellungen gefunden."
OpenButton = "Öffnen"
NotInstalled = "%s ist auf diesem System nicht installiert."
LearnMoreButton = "Mehr erfahren"
```

