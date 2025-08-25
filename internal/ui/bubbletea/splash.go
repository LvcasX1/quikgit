package bubbletea

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SplashModel struct {
	app *Application
}

// SplashTimeoutMsg is sent when the splash screen timeout occurs
type SplashTimeoutMsg struct{}

func NewSplashModel(app *Application) *SplashModel {
	return &SplashModel{
		app: app,
	}
}

func (m *SplashModel) Init() tea.Cmd {
	// Set a timer for 1 second to transition to main menu
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return SplashTimeoutMsg{}
	})
}

func (m *SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		// Allow skipping splash screen with any key
		return m, m.app.NavigateTo(StateMainMenu)
	case SplashTimeoutMsg:
		// Timeout reached, go to main menu
		return m, m.app.NavigateTo(StateMainMenu)
	}
	return m, nil
}

func (m *SplashModel) View() string {
	// Get screen dimensions (if available)
	width := m.app.width
	height := m.app.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// ASCII art for QuikGit
	asciiArt := `
 ██████╗ ██╗   ██╗██╗██╗  ██╗ ██████╗ ██╗████████╗
██╔═══██╗██║   ██║██║██║ ██╔╝██╔════╝ ██║╚══██╔══╝
██║   ██║██║   ██║██║█████╔╝ ██║  ███╗██║   ██║   
██║   ██║██║   ██║██║██╔═██╗ ██║   ██║██║   ██║   
╚██████╔╝╚██████╔╝██║██║  ██╗╚██████╔╝██║   ██║   
 ╚══▀▀═╝  ╚═════╝ ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝   ╚═╝   
                                                   
            GitHub Repository Manager              `

	// Style the ASCII art
	artStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center)

	styledArt := artStyle.Render(asciiArt)

	// Loading indicator
	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		MarginTop(3)

	loading := loadingStyle.Render("Loading... (press any key to skip)")

	// Combine all elements
	content := lipgloss.JoinVertical(lipgloss.Center, styledArt, loading)

	// Center on screen
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
