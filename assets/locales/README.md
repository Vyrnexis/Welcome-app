# Translating Solus Welcome

Solus Welcome loads community translations from this directory at runtime. Translation files are TOML overlays applied to the English base configuration in [`../config.toml`](../config.toml), so recompiling the application is not required.

## Locale file names

The application reads the system's `LANG` environment variable, removes its encoding and modifier, and searches for matching TOML files.

Use these file-name formats:

| Translation | File name |
| --- | --- |
| German | `de.toml` |
| French | `fr.toml` |
| Brazilian Portuguese | `pt_BR.toml` |
| European Portuguese | `pt_PT.toml` |
| Simplified Chinese | `zh_CN.toml` |
| Traditional Chinese | `zh_TW.toml` |

- Use a lowercase ISO 639-1 language code for a base translation.
- Add an uppercase ISO 3166-1 country code only when a regional variant is needed.
- Keep the separator as an underscore and the extension as `.toml`.
- Do not include encodings such as `.UTF-8` or modifiers such as `@latin` in the file name.

For `LANG=pt_BR.UTF-8`, resolution occurs in this order:

1. `pt_BR.toml`
2. `pt.toml`
3. English values from `../config.toml`

The `C`, `POSIX`, and empty locales use English directly. If a regional file is malformed, the error is reported and the loader continues to the base-language file.

## Start a translation

1. Check whether a base-language or regional file already exists.
2. Create the correctly named UTF-8 TOML file in this directory.
3. Begin with a partial `[UI]` table, or copy complete collections from `../config.toml` when translating cards and links.
4. Preserve all identifiers, commands, URLs, and format placeholders.
5. Run the application with the target `LANG` value and inspect every page.
6. Run the project tests before submitting the translation.

## Minimal translation

Scalar fields in `[UI]` can be translated individually. Omitted fields retain their English values:

```toml
[UI]
Tagline = "Das persönliche Betriebssystem für persönliche Computer"
ShowOnStartup = "Diesen Willkommensbildschirm beim Start anzeigen"
Close = "Schließen"
DarkTheme = "Dunkles Design"
WelcomeTitle = "Willkommen bei Solus"
WelcomeSubtitle = "Vielen Dank, dass Sie sich für Solus entschieden haben."
GettingStarted = "Erste Schritte"
CheckForUpdates = "Nach Updates suchen"
CheckingUpdates = "Suche nach Updates..."
UnableToCheckUpdates = "Updates konnten nicht geprüft werden"
SystemUpToDate = "System ist auf dem neuesten Stand"
OneUpdateAvailable = "1 Update verfügbar"
UpdatesAvailable = "%d Updates verfügbar"
OpenButton = "Öffnen"
NotInstalled = "%s ist auf diesem System nicht installiert."
LearnMoreButton = "Mehr erfahren"
```

You may include the remaining `[UI]` fields from `../config.toml` when their English values are unsuitable for the target language.

## Translate cards and navigation

TOML arrays of tables replace the corresponding English collection. If a translation includes one of the following collections, it must reproduce every entry from `../config.toml`:

- `[[NavItems]]`
- `[[WelcomeCards]]`
- `[[SupportLinks]]`
- `[[ContributeLinks]]`
- Each translated `[[DesktopActions.<desktop>]]` collection

Translate only the user-facing values while preserving routing keys, actions, URLs, and commands. For example:

```toml
[[NavItems]]
Key = "welcome"
Label = "Willkommen"

[[NavItems]]
Key = "customise"
Label = "Anpassen"

[[NavItems]]
Key = "support"
Label = "Hilfe"

[[NavItems]]
Key = "contribute"
Label = "Mitwirken"
```

Do not submit only one translated navigation item or card. Doing so replaces the entire English collection and removes the omitted entries from the interface.

## Translate missing-application messages

Command tables contain both executable arguments and a user-facing `Missing` message. A translated command table must retain the complete English `Command` array because overriding a map entry replaces that entry's complete value:

```toml
[SoftwareCommands.gnome]
Command = ["gnome-software"]
Missing = "GNOME Software ist nicht installiert."
```

Apply the same rule to `CustomiseCommands` and `SystemSettingsCommands`. Never translate executable names, command arguments, action identifiers, or desktop keys.

## Values that must remain unchanged

| Value | Reason |
| --- | --- |
| `Key` | Routes sidebar navigation |
| `Action` | Selects application behavior |
| `Command` | Launches installed programs |
| `URL` | Opens the configured web resource |
| Desktop table names | Match detected desktop identifiers |
| `%d` and `%s` | Supply runtime values to translated messages |

The `%d` placeholder in `UpdatesAvailable` receives the update count. The `%s` placeholder in `NotInstalled` receives an executable name. Preserve the placeholder type and include it exactly once.

## TOML rules

- Save files as UTF-8.
- Keep table and field names exactly as written in `../config.toml`.
- Quote string values and escape embedded double quotes as `\"`.
- Do not define the same table or field twice.
- Keep command arrays and URLs byte-for-byte identical to the English configuration.
- Use natural wording rather than translating individual English words without context.
- Keep button labels concise enough for the application layout.

## Test a translation

From the repository root, launch the application with the target locale:

```bash
LANG=de_DE.UTF-8 go run ./src
```

Check the Welcome, Customise, Support, and Contribute pages, both themes, update-status messages, and missing-application messages. Confirm that no navigation items, cards, or desktop actions disappear.

Then run the automated checks:

```bash
go test -v -race ./...
./test/install_test.sh
```

## Submission checklist

- The file name follows the locale naming convention.
- The file parses without a startup warning.
- The translation uses consistent terminology and punctuation.
- Every included collection contains all required entries.
- Identifiers, commands, URLs, and desktop keys are unchanged.
- `%d` and `%s` placeholders are preserved correctly.
- All pages and both themes were checked visually.
- The race-enabled Go tests and installer tests pass.
