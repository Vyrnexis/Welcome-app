package main

import (
	"errors"
	"fmt"
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

// NavLabel retrieves the label for a given navigation key or returns the fallback if not found.
func (c *AppContent) NavLabel(key, fallback string) string {
	for _, item := range c.NavItems {
		if item.Key == key {
			return item.Label
		}
	}
	return fallback
}

// WelcomeCardTitle retrieves the title for a card action identifier or returns the fallback.
func (c *AppContent) WelcomeCardTitle(action, fallback string) string {
	for _, card := range c.WelcomeCards {
		if card.Action == action {
			return card.Title
		}
	}
	return fallback
}

var Content AppContent

// getSystemLocales detects the system locale and returns ordered candidate language codes.
func getSystemLocales() []string {
	lang := os.Getenv("LANG")
	if lang == "" || lang == "C" || lang == "POSIX" {
		return []string{"en"}
	}
	clean := strings.Split(lang, ".")[0]
	clean = strings.Split(clean, "@")[0]
	if clean == "" {
		return []string{"en"}
	}
	base := strings.Split(clean, "_")[0]
	if clean != base {
		return []string{clean, base}
	}
	return []string{base}
}

// LoadContent reads the base TOML configuration and applies language-specific overrides if available.
func LoadContent(baseDir string) error {
	path := filepath.Join(baseDir, "assets", "config.toml")
	var baseContent AppContent
	if _, err := toml.DecodeFile(path, &baseContent); err != nil {
		return err
	}
	Content = baseContent

	var localeErrors []error

	for _, loc := range getSystemLocales() {
		if loc == "en" {
			continue
		}
		localizedPath := filepath.Join(baseDir, "assets", "locales", loc+".toml")
		if fileExists(localizedPath) {
			var localizedContent AppContent
			if _, err := toml.DecodeFile(path, &localizedContent); err != nil {
				return err
			}
			if _, err := toml.DecodeFile(localizedPath, &localizedContent); err != nil {
				localeErrors = append(localeErrors, fmt.Errorf("load locale %s: %w", loc, err))
				continue
			}
			Content = localizedContent
			return errors.Join(localeErrors...)
		}
	}
	return errors.Join(localeErrors...)
}
