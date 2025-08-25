package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary colors
	Primary   = lipgloss.Color("86")  // Bright green
	Secondary = lipgloss.Color("63")  // Bright blue
	Accent    = lipgloss.Color("226") // Bright yellow

	// Status colors
	Success = lipgloss.Color("82")  // Green
	Warning = lipgloss.Color("214") // Orange
	Error   = lipgloss.Color("196") // Red
	Info    = lipgloss.Color("117") // Light blue

	// Neutral colors
	Foreground = lipgloss.Color("255") // White
	Background = lipgloss.Color("0")   // Black
	Muted      = lipgloss.Color("240") // Dark gray
	Subtle     = lipgloss.Color("244") // Gray
	Border     = lipgloss.Color("238") // Darker gray
)

// Base styles
var (
	// Typography
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Align(lipgloss.Center).
		MarginBottom(1)

	Subtitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Secondary).
			MarginBottom(1)

	Heading = lipgloss.NewStyle().
		Bold(true).
		Foreground(Foreground)

	Body = lipgloss.NewStyle().
		Foreground(Foreground)

	Caption = lipgloss.NewStyle().
		Foreground(Subtle).
		Italic(true)

	Help = lipgloss.NewStyle().
		Foreground(Muted)

	// Status styles
	SuccessText = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	WarningText = lipgloss.NewStyle().
			Foreground(Warning).
			Bold(true)

	ErrorText = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	InfoText = lipgloss.NewStyle().
			Foreground(Info)
)

// Layout styles
var (
	// Containers
	Container = lipgloss.NewStyle().
			Padding(1, 2).
			Margin(1, 0)

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1, 2).
		Margin(1, 0)

	ActivePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2).
			Margin(1, 0)

	Card = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(1, 2).
		Margin(0, 1, 1, 0)

	// Input styles
	Input = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Border).
		Padding(0, 1)

	ActiveInput = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(0, 1)

	// Button styles
	Button = lipgloss.NewStyle().
		Background(Secondary).
		Foreground(Background).
		Padding(0, 2).
		MarginRight(1).
		Bold(true)

	PrimaryButton = lipgloss.NewStyle().
			Background(Primary).
			Foreground(Background).
			Padding(0, 2).
			MarginRight(1).
			Bold(true)

	// List styles
	ListItem = lipgloss.NewStyle().
			Padding(0, 2)

	SelectedListItem = lipgloss.NewStyle().
				Background(Primary).
				Foreground(Background).
				Padding(0, 2).
				Bold(true)

	// Progress styles
	ProgressBar = lipgloss.NewStyle().
			Foreground(Primary)

	ProgressComplete = lipgloss.NewStyle().
				Foreground(Success)

	ProgressError = lipgloss.NewStyle().
			Foreground(Error)
)

// Icon styles and definitions
var (
	IconStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	// Common icons
	Icons = map[string]string{
		"success":      "✅",
		"error":        "❌",
		"warning":      "⚠️",
		"info":         "ℹ️",
		"loading":      "🔄",
		"search":       "🔍",
		"clone":        "📁",
		"install":      "📦",
		"github":       "🐙",
		"star":         "⭐",
		"fork":         "🍴",
		"language":     "💬",
		"calendar":     "📅",
		"user":         "👤",
		"organization": "🏢",
		"topic":        "🏷️",
		"link":         "🔗",
		"lock":         "🔒",
		"unlock":       "🔓",
		"code":         "💻",
		"gear":         "⚙️",
		"help":         "❓",
		"arrow_right":  "→",
		"arrow_left":   "←",
		"arrow_up":     "↑",
		"arrow_down":   "↓",
		"check":        "✓",
		"cross":        "✗",
		"plus":         "+",
		"minus":        "-",
		"bullet":       "•",
	}
)

// Themed components
type Theme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Accent     lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Muted      lipgloss.Color
}

var DefaultTheme = Theme{
	Primary:    Primary,
	Secondary:  Secondary,
	Background: Background,
	Foreground: Foreground,
	Accent:     Accent,
	Success:    Success,
	Warning:    Warning,
	Error:      Error,
	Muted:      Muted,
}

var DarkTheme = Theme{
	Primary:    lipgloss.Color("75"),  // Light blue
	Secondary:  lipgloss.Color("141"), // Light purple
	Background: lipgloss.Color("0"),   // Black
	Foreground: lipgloss.Color("255"), // White
	Accent:     lipgloss.Color("220"), // Light yellow
	Success:    lipgloss.Color("83"),  // Light green
	Warning:    lipgloss.Color("208"), // Orange
	Error:      lipgloss.Color("203"), // Light red
	Muted:      lipgloss.Color("243"), // Light gray
}

// Apply theme to styles
func ApplyTheme(theme Theme) {
	// Update global styles with theme colors
	Title = Title.Foreground(theme.Primary)
	Subtitle = Subtitle.Foreground(theme.Secondary)
	Heading = Heading.Foreground(theme.Foreground)
	Body = Body.Foreground(theme.Foreground)
	Help = Help.Foreground(theme.Muted)

	SuccessText = SuccessText.Foreground(theme.Success)
	WarningText = WarningText.Foreground(theme.Warning)
	ErrorText = ErrorText.Foreground(theme.Error)

	ActivePanel = ActivePanel.BorderForeground(theme.Primary)
	ActiveInput = ActiveInput.BorderForeground(theme.Primary)
	Button = Button.Background(theme.Secondary).Foreground(theme.Background)
	PrimaryButton = PrimaryButton.Background(theme.Primary).Foreground(theme.Background)

	SelectedListItem = SelectedListItem.Background(theme.Primary).Foreground(theme.Background)
	ProgressBar = ProgressBar.Foreground(theme.Primary)
	ProgressComplete = ProgressComplete.Foreground(theme.Success)
	ProgressError = ProgressError.Foreground(theme.Error)
	IconStyle = IconStyle.Foreground(theme.Primary)
}

// Utility functions
func WithIcon(icon, text string) string {
	if iconChar, exists := Icons[icon]; exists {
		return IconStyle.Render(iconChar) + " " + text
	}
	return text
}

func StatusIcon(success bool) string {
	if success {
		return Icons["success"]
	}
	return Icons["error"]
}

func LanguageIcon(language string) string {
	languageIcons := map[string]string{
		"Go":         "🐹",
		"JavaScript": "🟨",
		"TypeScript": "🔷",
		"Python":     "🐍",
		"Java":       "☕",
		"C++":        "⚡",
		"C":          "🔧",
		"C#":         "💜",
		"Ruby":       "💎",
		"PHP":        "🐘",
		"Swift":      "🍎",
		"Kotlin":     "🟠",
		"Rust":       "🦀",
		"Scala":      "🔺",
		"Shell":      "🐚",
		"HTML":       "🌐",
		"CSS":        "🎨",
		"Dart":       "🎯",
		"R":          "📊",
		"Lua":        "🌙",
	}

	if icon, exists := languageIcons[language]; exists {
		return icon
	}
	return Icons["code"]
}

// Responsive utilities
func AdaptiveWidth(width int) lipgloss.Style {
	if width < 60 {
		return lipgloss.NewStyle().Width(width - 4)
	} else if width < 100 {
		return lipgloss.NewStyle().Width(width - 8)
	}
	return lipgloss.NewStyle().Width(width - 12)
}

func AdaptiveHeight(height int) lipgloss.Style {
	if height < 20 {
		return lipgloss.NewStyle().Height(height - 4)
	}
	return lipgloss.NewStyle().Height(height - 6)
}
