package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/Vyrnexis/welcome-app/assets"
)

// systemConfigPath is the optional distro-provided override for the embedded configuration.
const systemConfigPath = "/etc/solus-welcome/config.toml"

// defaultDesktopKey names the command entry used when the detected desktop has none.
const defaultDesktopKey = "default"

// Fields tagged i18n:"-" are identifiers, commands, or URLs that translators must never see.

type CommandInfo struct {
	Command []string `toml:"Command" i18n:"-"`
	Missing string   `toml:"Missing"`
}

type WelcomeCard struct {
	Title  string `toml:"Title"`
	Body   string `toml:"Body"`
	Button string `toml:"Button"`
	Action string `toml:"Action" i18n:"-"`
	URL    string `toml:"URL" i18n:"-"`
}

type LinkCard struct {
	Title string `toml:"Title"`
	Body  string `toml:"Body"`
	URL   string `toml:"URL" i18n:"-"`
}

type DesktopAction struct {
	Title   string   `toml:"Title"`
	Body    string   `toml:"Body"`
	Command []string `toml:"Command" i18n:"-"`
}

// DriverHelp configures the NVIDIA driver card: Command is tried first and URL is the fallback.
type DriverHelp struct {
	Title   string   `toml:"Title"`
	Body    string   `toml:"Body"`
	Button  string   `toml:"Button"`
	Command []string `toml:"Command" i18n:"-"`
	URL     string   `toml:"URL" i18n:"-"`
}

type NavItem struct {
	Key   string `toml:"Key" i18n:"-"`
	Label string `toml:"Label"`
}

type UIStrings struct {
	WindowTitle          string `toml:"WindowTitle"`
	SidebarTitle         string `toml:"SidebarTitle"`
	Tagline              string `toml:"Tagline"`
	ShowOnStartup        string `toml:"ShowOnStartup"`
	Close                string `toml:"Close"`
	ThemeLabel           string `toml:"ThemeLabel"`
	ThemeSystem          string `toml:"ThemeSystem"`
	ThemeLight           string `toml:"ThemeLight"`
	ThemeDark            string `toml:"ThemeDark"`
	WelcomeTitle         string `toml:"WelcomeTitle"`
	WelcomeSubtitle      string `toml:"WelcomeSubtitle"`
	GettingStarted       string `toml:"GettingStarted"`
	SolusLinux           string `toml:"SolusLinux"`
	EditionSuffix        string `toml:"EditionSuffix"`
	UnknownDesktop       string `toml:"UnknownDesktop"`
	Bits64               string `toml:"Bits64"`
	MemoryFormat         string `toml:"MemoryFormat"`
	CheckForUpdates      string `toml:"CheckForUpdates"`
	CheckingUpdates      string `toml:"CheckingUpdates"`
	UnableToCheckUpdates string `toml:"UnableToCheckUpdates"`
	UpdateTimedOut       string `toml:"UpdateTimedOut"`
	UpdateCheckFailed    string `toml:"UpdateCheckFailed"`
	SystemUpToDate       string `toml:"SystemUpToDate"`
	NoDesktopSettings    string `toml:"NoDesktopSettings"`
	OpenButton           string `toml:"OpenButton"`
	NotInstalled         string `toml:"NotInstalled"`
	LearnMoreButton      string `toml:"LearnMoreButton"`
	NoTerminal           string `toml:"NoTerminal"`
	NoAutostartDir       string `toml:"NoAutostartDir"`
	PressEnterToClose    string `toml:"PressEnterToClose"`
	CopySystemInfo       string `toml:"CopySystemInfo"`
	Copied               string `toml:"Copied"`
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
	Drivers                DriverHelp                 `toml:"Drivers"`
	UI                     UIStrings                  `toml:"UI"`
}

var errBaseContent = errors.New("base content unavailable")

// NavLabel returns the configured label for a navigation key, or the key itself if it is not configured.
func (c *AppContent) NavLabel(key string) string {
	for _, item := range c.NavItems {
		if item.Key == key {
			return item.Label
		}
	}
	return key
}

// commandFor picks the command for a desktop, falling back to the configured default entry.
func commandFor(commands map[string]CommandInfo, desktopKey string) CommandInfo {
	if command, ok := commands[desktopKey]; ok {
		return command
	}
	return commands[defaultDesktopKey]
}

// commandGroups names the per-desktop command tables so they can be validated uniformly.
func (c *AppContent) commandGroups() map[string]map[string]CommandInfo {
	return map[string]map[string]CommandInfo{
		"software":  c.SoftwareCommands,
		"customise": c.CustomiseCommands,
		"settings":  c.SystemSettingsCommands,
	}
}

// loadContent returns the system override when it is valid and the embedded configuration otherwise.
// The returned source names the file used; an error is only returned when the embedded configuration is unusable.
func loadContent(overridePath string) (content AppContent, source string, err error) {
	if overridePath != "" {
		data, readErr := os.ReadFile(overridePath)
		switch {
		case readErr == nil:
			override, decodeErr := decodeContent(data)
			if decodeErr == nil {
				return override, overridePath, nil
			}
			log.Printf("warning: ignoring %s: %v", overridePath, decodeErr)
		case !errors.Is(readErr, os.ErrNotExist):
			log.Printf("warning: ignoring %s: %v", overridePath, readErr)
		}
	}

	data, err := assets.FS.ReadFile("config.toml")
	if err != nil {
		return AppContent{}, "", errors.Join(errBaseContent, err)
	}
	content, err = decodeContent(data)
	if err != nil {
		return AppContent{}, "", errors.Join(errBaseContent, err)
	}
	return content, "embedded", nil
}

// decodeContent parses and validates a configuration document, rejecting unknown keys.
func decodeContent(data []byte) (AppContent, error) {
	var content AppContent
	metadata, err := toml.Decode(string(data), &content)
	if err != nil {
		return AppContent{}, fmt.Errorf("decode configuration: %w", err)
	}
	if keys := metadata.Undecoded(); len(keys) > 0 {
		return AppContent{}, fmt.Errorf("unknown configuration keys: %v", keys)
	}
	if err := validateContent(content); err != nil {
		return AppContent{}, err
	}
	return content, nil
}

// validURL reports whether raw is an absolute http(s) URL.
func validURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Host != ""
}

// validateContent rejects configuration that would produce missing controls, silent actions, or unsafe links.
func validateContent(content AppContent) error {
	if len(content.NavItems) == 0 {
		return errors.New("configuration has no navigation items")
	}
	seenNav := make(map[string]bool, len(content.NavItems))
	for _, item := range content.NavItems {
		if _, ok := pages[item.Key]; !ok || strings.TrimSpace(item.Label) == "" {
			return fmt.Errorf("invalid navigation item %q", item.Key)
		}
		if seenNav[item.Key] {
			return fmt.Errorf("duplicate navigation item %q", item.Key)
		}
		seenNav[item.Key] = true
	}

	seenActions := make(map[string]bool, len(content.WelcomeCards))
	for _, card := range content.WelcomeCards {
		spec, ok := actions[card.Action]
		if !ok || spec.internal || strings.TrimSpace(card.Title) == "" || strings.TrimSpace(card.Button) == "" {
			return fmt.Errorf("invalid welcome action %q", card.Action)
		}
		if spec.needsURL && !validURL(card.URL) {
			return fmt.Errorf("invalid URL %q for welcome action %q", card.URL, card.Action)
		}
		if seenActions[card.Action] {
			return fmt.Errorf("duplicate welcome action %q", card.Action)
		}
		seenActions[card.Action] = true
	}

	for group, commands := range content.commandGroups() {
		if _, ok := commands[defaultDesktopKey]; !ok {
			return fmt.Errorf("missing default %s command", group)
		}
		for desktop, command := range commands {
			if (desktop != defaultDesktopKey && !knownDesktop(desktop)) || len(command.Command) == 0 ||
				strings.TrimSpace(command.Command[0]) == "" || strings.TrimSpace(command.Missing) == "" {
				return fmt.Errorf("invalid %s command for %q", group, desktop)
			}
		}
	}
	for desktop, desktopActions := range content.DesktopActions {
		if !knownDesktop(desktop) {
			return fmt.Errorf("invalid desktop action group %q", desktop)
		}
		for _, action := range desktopActions {
			if strings.TrimSpace(action.Title) == "" || len(action.Command) == 0 || strings.TrimSpace(action.Command[0]) == "" {
				return fmt.Errorf("invalid desktop action for %q", desktop)
			}
		}
	}

	for _, links := range [][]LinkCard{content.SupportLinks, content.ContributeLinks} {
		for _, link := range links {
			if !validURL(link.URL) {
				return fmt.Errorf("invalid link URL %q", link.URL)
			}
		}
	}

	drivers := content.Drivers
	if strings.TrimSpace(drivers.Title) == "" || strings.TrimSpace(drivers.Body) == "" || strings.TrimSpace(drivers.Button) == "" ||
		len(drivers.Command) == 0 || strings.TrimSpace(drivers.Command[0]) == "" {
		return errors.New("invalid driver helper configuration")
	}
	if !validURL(drivers.URL) {
		return fmt.Errorf("invalid driver helper URL %q", drivers.URL)
	}

	ui := reflect.ValueOf(content.UI)
	for i := range ui.NumField() {
		if strings.TrimSpace(ui.Field(i).String()) == "" {
			return fmt.Errorf("missing UI field %s", ui.Type().Field(i).Name)
		}
	}
	formats := []struct {
		name, format string
		sample       any
	}{
		{"NotInstalled", content.UI.NotInstalled, "application"},
		{"MemoryFormat", content.UI.MemoryFormat, 1.5},
	}
	for _, f := range formats {
		if !strings.Contains(f.format, "%") || strings.Contains(fmt.Sprintf(f.format, f.sample), "%!") {
			return fmt.Errorf("UI.%s has invalid formatting", f.name)
		}
	}
	return nil
}
