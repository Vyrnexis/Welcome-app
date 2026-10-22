# Translating Solus Welcome

Solus Welcome uses Fyne's built-in i18n. Each language is one JSON catalogue in this directory, embedded into the binary at build time, so a translation needs no code changes. The language is chosen automatically from the user's locale; if a message is missing, the English text is shown.

## How messages are identified

The message ID is the **English text** exactly as written in [`../config.toml`](../config.toml) (card titles and bodies, button labels, navigation labels, `Missing` messages, and the `[UI]` strings). Commands, URLs, action names, and desktop keys are never message IDs, so you never copy or translate them.

There are two exceptions: the number of available updates (`updates_available`) and the getting-started progress (`getting_started_progress`) are plural messages with stable IDs.

## Add a language

1. Create `<language>.json` in this directory, e.g. `de.json`, `fr.json`, `pt_BR.json`, or `zh_CN.json`. Use a lowercase ISO 639-1 code, plus an uppercase ISO 3166-1 region only when the variant differs.
2. Add `"English text": "Translated text"` entries. You may translate only some strings; the rest stay English.
3. Run the app with your locale to check every page, e.g. `LANG=de_DE.UTF-8 go run ./src`.
4. Run `go test -race ./...`.

[`de.json`](de.json) is a partial example.

## Rules

- Save as UTF-8 and keep the file valid JSON.
- Copy the English ID character for character, including punctuation and leading or trailing spaces (for example `" Edition"`). The tests reject keys that no longer match any string in the configuration.
- Preserve format placeholders such as `%s` and `%.1f` exactly once each, in any order your grammar needs. The tests reject translations whose placeholders differ from the English text.
- Keep button labels short so they fit the layout.

## Plurals

The update count and the progress summary (`{{.Count}} of {{.Total}} done`) use the CLDR plural categories of your language. Provide every category your language uses (for example `one` and `other` for German, `one`, `few`, `many`, and `other` for Polish) and put `{{.Count}}` where the number belongs:

```json
"updates_available": {
  "one": "{{.Count}} Update verfügbar",
  "other": "{{.Count}} Updates verfügbar"
}
```

If a language omits this entry, the English form is shown, so translate it together with the rest.

`en.json` only defines the English singular and plural forms; it is not a translation target.
