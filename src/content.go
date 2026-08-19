package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type CommandInfo struct {
	Command []string `json:"Command"`
	Missing string   `json:"Missing"`
}

type WelcomeCard struct {
	Title  string `json:"Title"`
	Body   string `json:"Body"`
	Button string `json:"Button"`
	Action string `json:"Action"`
}

type LinkCard struct {
	Title string `json:"Title"`
	Body  string `json:"Body"`
	URL   string `json:"URL"`
}

type DesktopAction struct {
	Title   string   `json:"Title"`
	Body    string   `json:"Body"`
	Command []string `json:"Command"`
}

type NavItem struct {
	Key   string `json:"Key"`
	Label string `json:"Label"`
}

type AppContent struct {
	NavItems               []NavItem                  `json:"NavItems"`
	WelcomeCards           []WelcomeCard              `json:"WelcomeCards"`
	SoftwareCommands       map[string]CommandInfo     `json:"SoftwareCommands"`
	CustomiseCommands      map[string]CommandInfo     `json:"CustomiseCommands"`
	SystemSettingsCommands map[string]CommandInfo     `json:"SystemSettingsCommands"`
	DesktopActions         map[string][]DesktopAction `json:"DesktopActions"`
	SupportLinks           []LinkCard                 `json:"SupportLinks"`
	ContributeLinks        []LinkCard                 `json:"ContributeLinks"`
}

var Content AppContent

// LoadContent reads and parses the external configuration file.
func LoadContent(baseDir string) error {
	path := filepath.Join(baseDir, "assets", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &Content)
}
