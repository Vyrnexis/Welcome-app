package main

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type CommandInfo struct {
	Command []string `toml:"Command"`
	Missing string   `toml:"Missing"`
}

type WelcomeCard struct {
	Title  string `toml:"Title"`
	Body   string `toml:"Body"`
	Button string `toml:"Button"`
	Action string `toml:"Action"`
}

type LinkCard struct {
	Title string `toml:"Title"`
	Body  string `toml:"Body"`
	URL   string `toml:"URL"`
}

type DesktopAction struct {
	Title   string   `toml:"Title"`
	Body    string   `toml:"Body"`
	Command []string `toml:"Command"`
}

type NavItem struct {
	Key   string `toml:"Key"`
	Label string `toml:"Label"`
}

type AppContent struct {
	NavItems               []NavItem                  `toml:"NavItems"`
	WelcomeCards           []WelcomeCard              `toml:"WelcomeCards"`
	SoftwareCommands       map[string]CommandInfo     `toml:"SoftwareCommands"`
	CustomiseCommands      map[string]CommandInfo     `toml:"CustomiseCommands"`
	SystemSettingsCommands map[string]CommandInfo     `toml:"SystemSettingsCommands"`
	DesktopActions         map[string][]DesktopAction `toml:"DesktopActions"`
	SupportLinks           []LinkCard                 `toml:"SupportLinks"`
	ContributeLinks        []LinkCard                 `toml:"ContributeLinks"`
}

var Content AppContent

// LoadContent reads and parses the external TOML configuration file.
func LoadContent(baseDir string) error {
	path := filepath.Join(baseDir, "assets", "config.toml")
	_, err := toml.DecodeFile(path, &Content)
	return err
}
