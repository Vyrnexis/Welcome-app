// Package assets embeds the application's configuration, icons, and translations so the binary is self-contained.
package assets

import "embed"

// FS holds config.toml, the SVG icons, and the JSON translation catalogues.
//
//go:embed config.toml *.svg translations/*.json
var FS embed.FS
