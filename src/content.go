package main

import (
	"os"
	"path/filepath"
	"strings"

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

type UIStrings struct {
	Tagline              string `toml:"Tagline"`
	ShowOnStartup        string `toml:"ShowOnStartup"`
	Close                string `toml:"Close"`
	DarkTheme            string `toml:"DarkTheme"`
	WelcomeTitle         string `toml:"WelcomeTitle"`
	WelcomeSubtitle      string `toml:"WelcomeSubtitle"`
	GettingStarted       string `toml:"GettingStarted"`
	SolusLinux           string `toml:"SolusLinux"`
	EditionSuffix        string `toml:"EditionSuffix"`
	CheckForUpdates      string `toml:"CheckForUpdates"`
	CheckingUpdates      string `toml:"CheckingUpdates"`
	UnableToCheckUpdates string `toml:"UnableToCheckUpdates"`
	SystemUpToDate       string `toml:"SystemUpToDate"`
	OneUpdateAvailable   string `toml:"OneUpdateAvailable"`
	UpdatesAvailable     string `toml:"UpdatesAvailable"`
	NoDesktopSettings    string `toml:"NoDesktopSettings"`
	OpenButton           string `toml:"OpenButton"`
	NotInstalled         string `toml:"NotInstalled"`
	LearnMoreButton      string `toml:"LearnMoreButton"`
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
	UI                     UIStrings                  `toml:"UI"`
}

var Content AppContent

// getSystemLanguage detects the system locale and returns the base language code.
func getSystemLanguage() string {
	lang := os.Getenv("LANG")
	if lang == "" || lang == "C" {
		return "en"
	}
	lang = strings.Split(lang, ".")[0] // Strip encoding (e.g. .UTF-8)
	lang = strings.Split(lang, "_")[0] // Strip region (e.g. fr_FR -> fr)
	return lang
}

// LoadContent reads the base TOML configuration and applies language-specific overrides if available.
func LoadContent(baseDir string) error {
	path := filepath.Join(baseDir, "assets", "config.toml")
	if _, err := toml.DecodeFile(path, &Content); err != nil {
		return err
	}

	lang := getSystemLanguage()
	if lang != "en" {
		localizedPath := filepath.Join(baseDir, "assets", "locales", lang+".toml")
		if fileExists(localizedPath) {
			// Decoding over the existing Content struct safely overwrites only the keys present in the translation file.
			_, _ = toml.DecodeFile(localizedPath, &Content)
		}
	}
	return nil
}
