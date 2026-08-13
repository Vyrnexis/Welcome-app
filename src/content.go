package main

type CommandInfo struct {
	Command []string
	Missing string
}

type WelcomeCard struct {
	Title  string
	Body   string
	Button string
	Action welcomeAction
}

type LinkCard struct {
	Title string
	Body  string
	URL   string
}

type DesktopAction struct {
	Title   string
	Body    string
	Command []string
}

type NavItem struct {
	Key   string
	Label string
}

var navItems = []NavItem{
	{"welcome", "Welcome"},
	{"customise", "Customise"},
	{"support", "Support"},
	{"contribute", "Contribute"},
}

type welcomeAction int

const (
	actionUpdates welcomeAction = iota
	actionSoftware
	actionCustomise
	actionLearn
	actionSettings
	actionDonate
)

var welcomeCards = []WelcomeCard{
	{
		Title:  "Update your system",
		Body:   "Keep your system secure and up to date with the latest packages.",
		Button: "Check for Updates",
		Action: actionUpdates,
	},
	{
		Title:  "Install common apps",
		Body:   "Browse and install apps like browsers, codecs, Steam, Discord, and more.",
		Button: "Open Software Centre",
		Action: actionSoftware,
	},
	{
		Title:  "Customise your desktop",
		Body:   "Make Solus feel like home. Adjust themes, layout, appearance, and more.",
		Button: "Open System Settings",
		Action: actionCustomise,
	},
	{
		Title:  "Learn Solus",
		Body:   "Find help, documentation, and community support resources.",
		Button: "Visit Help Centre",
		Action: actionLearn,
	},
	{
		Title:  "System Settings",
		Body:   "Configure displays, printers, Bluetooth, audio, power, and more.",
		Button: "Open Settings",
		Action: actionSettings,
	},
	{
		Title:  "Donate to Solus",
		Body:   "Support Solus infrastructure, hosting, development, and community work.",
		Button: "Donate",
		Action: actionDonate,
	},
}

var softwareCommands = map[string]CommandInfo{
	"gnome":  {[]string{"gnome-software"}, "GNOME Software is not installed."},
	"budgie": {[]string{"plasma-discover"}, "Discover is not installed."},
	"kde":    {[]string{"plasma-discover"}, "Discover is not installed."},
}

var customiseCommands = map[string]CommandInfo{
	"budgie": {[]string{"budgie-desktop-settings"}, "Budgie Desktop Settings is not installed."},
	"kde":    {[]string{"kcmshell6", "kcm_lookandfeel"}, "KDE appearance settings are not installed."},
	"gnome":  {[]string{"gnome-control-center", "background"}, "GNOME appearance settings are not installed."},
	"xfce":   {[]string{"xfce4-appearance-settings"}, "Xfce Appearance settings are not installed."},
}

var systemSettingsCommands = map[string]CommandInfo{
	"budgie": {[]string{"budgie-control-center"}, "Budgie Control Center is not installed."},
	"kde":    {[]string{"systemsettings"}, "KDE System Settings is not installed."},
	"gnome":  {[]string{"gnome-control-center", "display"}, "GNOME display settings are not installed."},
	"xfce":   {[]string{"xfce4-settings-manager"}, "Xfce Settings Manager is not installed."},
}

var desktopActions = map[string][]DesktopAction{
	"kde": {
		{"System Settings", "Launch the KDE Plasma settings centre.", []string{"systemsettings"}},
		{"Appearance", "Change Plasma style, colours, icons, and global theme.", []string{"kcmshell6", "kcm_lookandfeel"}},
		{"Colours", "Adjust KDE colour schemes.", []string{"kcmshell6", "kcm_colors"}},
		{"Displays", "Configure monitors, scale, refresh rate, and layout.", []string{"kcmshell6", "kcm_kscreen"}},
	},
	"budgie": {
		{"Budgie Desktop Settings", "Adjust panels, applets, Raven, desktop icons, and style.", []string{"budgie-desktop-settings"}},
		{"System Settings", "Open Budgie Control Center for system-wide preferences.", []string{"budgie-control-center"}},
		{"Wallpapers", "Choose desktop backgrounds.", []string{"budgie-control-center", "background"}},
		{"Notifications", "Manage notification behaviour.", []string{"budgie-control-center", "notifications"}},
		{"Sound", "Configure speakers, microphones, and alert sounds.", []string{"budgie-control-center", "sound"}},
		{"Printers", "Add and manage printers.", []string{"budgie-control-center", "printers"}},
		{"Power", "Configure power saving and suspend behaviour.", []string{"budgie-control-center", "power"}},
		{"Keyboard and shortcuts", "Adjust typing settings and keyboard shortcuts.", []string{"budgie-control-center", "keyboard"}},
	},
	"gnome": {
		{"Settings", "Launch GNOME Settings.", []string{"gnome-control-center"}},
		{"Appearance", "Change style, background, and dock options.", []string{"gnome-control-center", "background"}},
		{"Displays", "Configure monitors, scale, refresh rate, and layout.", []string{"gnome-control-center", "display"}},
		{"Sound", "Configure speakers, microphones, and alert sounds.", []string{"gnome-control-center", "sound"}},
		{"Power", "Configure power saving and suspend behaviour.", []string{"gnome-control-center", "power"}},
		{"Extensions", "Manage GNOME Shell extensions.", []string{"gnome-extensions-app"}},
	},
	"xfce": {
		{"Settings Manager", "Launch the Xfce settings manager.", []string{"xfce4-settings-manager"}},
		{"Appearance", "Change GTK theme, icons, fonts, and toolbar style.", []string{"xfce4-appearance-settings"}},
		{"Displays", "Configure monitors, scale, refresh rate, and layout.", []string{"xfce4-display-settings"}},
		{"Window Manager", "Configure window decorations, focus, and keyboard shortcuts.", []string{"xfwm4-settings"}},
		{"Panel", "Configure Xfce panels, items, and layout.", []string{"xfce4-panel", "--preferences"}},
		{"Power Manager", "Configure power saving and display blanking.", []string{"xfce4-power-manager-settings"}},
	},
}

var supportLinks = []LinkCard{
	{"Solus Help Centre", "Start with official help and troubleshooting.", "https://help.getsol.us/"},
	{"Forums", "Ask questions and share solutions with the community.", "https://discuss.getsol.us/"},
	{"Matrix Community", "Join the Solus Matrix room for real-time community discussion.", "https://matrix.to/#/#solus:matrix.org"},
	{"Mastodon", "Follow Solus project updates on Mastodon.", "https://floss.social/@getsolus"},
	{"Bluesky", "Follow Solus project updates on Bluesky.", "https://bsky.app/profile/getsolus.bsky.social"},
}

var contributeLinks = []LinkCard{
	{"Report Bugs", "File high-quality reports with logs and clear reproduction steps.", "https://getsol.us/contribute/report-bugs"},
	{"Documentation", "Improve guides for new and experienced Solus users.", "https://getsol.us/contribute/documentation"},
	{"Packaging", "Help maintain software packages for the repository.", "https://getsol.us/contribute/packaging"},
	{"Development", "Contribute to Solus tooling and desktop integration.", "https://getsol.us/contribute/development"},
	{"Donations", "Support project infrastructure and ongoing work.", "https://getsol.us/contribute/donations"},
}
