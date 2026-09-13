package main

import (
	"errors"
	"fmt"
	"net/url"
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

var errBaseContent = errors.New("base content unavailable")

// getSystemLocales detects the system locale and returns ordered candidate language codes.
func getSystemLocales() []string {
	lang := os.Getenv("LC_ALL")
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}
	if lang == "" {
		lang = os.Getenv("LANG")
	}
	if lang == "" || lang == "C" || lang == "POSIX" {
		return []string{"en"}
	}
	clean := strings.Split(lang, ".")[0]
	clean = strings.Split(clean, "@")[0]
	if clean == "" || clean == "C" || clean == "POSIX" {
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
	metadata, err := toml.DecodeFile(path, &baseContent)
	if err != nil {
		return errors.Join(errBaseContent, fmt.Errorf("load base content: %w", err))
	}
	if keys := metadata.Undecoded(); len(keys) > 0 {
		return errors.Join(errBaseContent, fmt.Errorf("unknown base content keys: %v", keys))
	}
	if err := validateContent(baseContent); err != nil {
		return errors.Join(errBaseContent, err)
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
				return errors.Join(errBaseContent, fmt.Errorf("reload base content: %w", err))
			}
			metadata, err := toml.DecodeFile(localizedPath, &localizedContent)
			if err != nil {
				localeErrors = append(localeErrors, fmt.Errorf("load locale %s: %w", loc, err))
				continue
			}
			if keys := metadata.Undecoded(); len(keys) > 0 {
				localeErrors = append(localeErrors, fmt.Errorf("load locale %s: unknown keys: %v", loc, keys))
				continue
			}
			if err := validateContent(localizedContent); err != nil {
				localeErrors = append(localeErrors, fmt.Errorf("load locale %s: %w", loc, err))
				continue
			}
			Content = localizedContent
			return errors.Join(localeErrors...)
		}
	}
	return errors.Join(localeErrors...)
}

// validateContent rejects configuration that would produce missing controls, silent actions, or unsafe links.
func validateContent(content AppContent) error {
	knownNav := map[string]bool{"welcome": true, "customise": true, "support": true, "contribute": true}
	seenNav := make(map[string]bool, len(content.NavItems))
	if len(content.NavItems) == 0 {
		return errors.New("configuration has no navigation items")
	}
	for _, item := range content.NavItems {
		if !knownNav[item.Key] || strings.TrimSpace(item.Label) == "" {
			return fmt.Errorf("invalid navigation item %q", item.Key)
		}
		if seenNav[item.Key] {
			return fmt.Errorf("duplicate navigation item %q", item.Key)
		}
		seenNav[item.Key] = true
	}

	knownActions := map[string]bool{"updates": true, "software": true, "customise": true, "learn": true, "settings": true, "donate": true}
	seenActions := make(map[string]bool, len(content.WelcomeCards))
	for _, card := range content.WelcomeCards {
		if !knownActions[card.Action] || strings.TrimSpace(card.Title) == "" || strings.TrimSpace(card.Button) == "" {
			return fmt.Errorf("invalid welcome action %q", card.Action)
		}
		if seenActions[card.Action] {
			return fmt.Errorf("duplicate welcome action %q", card.Action)
		}
		seenActions[card.Action] = true
	}

	knownDesktops := map[string]bool{"budgie": true, "gnome": true, "kde": true, "xfce": true}
	commandGroups := map[string]map[string]CommandInfo{
		"software":  content.SoftwareCommands,
		"customise": content.CustomiseCommands,
		"settings":  content.SystemSettingsCommands,
	}
	for group, commands := range commandGroups {
		for desktop, command := range commands {
			if !knownDesktops[desktop] || len(command.Command) == 0 || strings.TrimSpace(command.Command[0]) == "" || strings.TrimSpace(command.Missing) == "" {
				return fmt.Errorf("invalid %s command for %q", group, desktop)
			}
		}
	}
	for desktop, actions := range content.DesktopActions {
		if !knownDesktops[desktop] {
			return fmt.Errorf("invalid desktop action group %q", desktop)
		}
		for _, action := range actions {
			if strings.TrimSpace(action.Title) == "" || len(action.Command) == 0 || strings.TrimSpace(action.Command[0]) == "" {
				return fmt.Errorf("invalid desktop action for %q", desktop)
			}
		}
	}

	linkGroups := [][]LinkCard{content.SupportLinks, content.ContributeLinks}
	for _, links := range linkGroups {
		for _, link := range links {
			parsed, err := url.ParseRequestURI(link.URL)
			if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
				return fmt.Errorf("invalid link URL %q", link.URL)
			}
		}
	}

	requiredUI := []struct {
		name  string
		value string
	}{
		{"Tagline", content.UI.Tagline},
		{"ShowOnStartup", content.UI.ShowOnStartup},
		{"Close", content.UI.Close},
		{"DarkTheme", content.UI.DarkTheme},
		{"WelcomeTitle", content.UI.WelcomeTitle},
		{"WelcomeSubtitle", content.UI.WelcomeSubtitle},
		{"GettingStarted", content.UI.GettingStarted},
		{"SolusLinux", content.UI.SolusLinux},
		{"EditionSuffix", content.UI.EditionSuffix},
		{"CheckForUpdates", content.UI.CheckForUpdates},
		{"CheckingUpdates", content.UI.CheckingUpdates},
		{"UnableToCheckUpdates", content.UI.UnableToCheckUpdates},
		{"SystemUpToDate", content.UI.SystemUpToDate},
		{"OneUpdateAvailable", content.UI.OneUpdateAvailable},
		{"UpdatesAvailable", content.UI.UpdatesAvailable},
		{"NoDesktopSettings", content.UI.NoDesktopSettings},
		{"OpenButton", content.UI.OpenButton},
		{"NotInstalled", content.UI.NotInstalled},
		{"LearnMoreButton", content.UI.LearnMoreButton},
	}
	for _, field := range requiredUI {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("missing UI field %s", field.name)
		}
	}
	if !strings.Contains(content.UI.UpdatesAvailable, "%d") {
		return errors.New("UI.UpdatesAvailable must contain %d")
	}
	if strings.Contains(fmt.Sprintf(content.UI.UpdatesAvailable, 1), "%!") {
		return errors.New("UI.UpdatesAvailable has invalid formatting")
	}
	if !strings.Contains(content.UI.NotInstalled, "%s") {
		return errors.New("UI.NotInstalled must contain %s")
	}
	if strings.Contains(fmt.Sprintf(content.UI.NotInstalled, "application"), "%!") {
		return errors.New("UI.NotInstalled has invalid formatting")
	}
	return nil
}
