package components

import (
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

// Spinner represents a loading spinner with animation
type Spinner struct {
	frames   []string
	current  int
	active   bool
	spring   harmonica.Spring
	lastTick time.Time
	style    lipgloss.Style
}

// SpinnerTickMsg is sent on each animation frame
type SpinnerTickMsg struct{}

// NewSpinner creates a new spinner component
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		spring: harmonica.NewSpring(harmonica.FPS(60), 10.0, 1.0),
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),
	}
}

// NewDotsSpinner creates a dots-style spinner
func NewDotsSpinner() *Spinner {
	return &Spinner{
		frames: []string{"   ", ".  ", ".. ", "...", " ..", "  .", "   "},
		spring: harmonica.NewSpring(harmonica.FPS(30), 8.0, 1.0),
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),
	}
}

// NewLineSpinner creates a line-style spinner
func NewLineSpinner() *Spinner {
	return &Spinner{
		frames: []string{"|", "/", "-", "\\"},
		spring: harmonica.NewSpring(harmonica.FPS(40), 12.0, 1.0),
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() tea.Cmd {
	s.active = true
	s.lastTick = time.Now()
	return s.Tick()
}

// Stop ends the spinner animation
func (s *Spinner) Stop() {
	s.active = false
}

// Tick returns a command that sends a tick message
func (s *Spinner) Tick() tea.Cmd {
	if !s.active {
		return nil
	}

	return tea.Tick(time.Millisecond*100, func(time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

// Update handles spinner messages
func (s *Spinner) Update(msg tea.Msg) (*Spinner, tea.Cmd) {
	switch msg.(type) {
	case SpinnerTickMsg:
		if s.active {
			s.current = (s.current + 1) % len(s.frames)
			return s, s.Tick()
		}
	}
	return s, nil
}

// View renders the spinner
func (s *Spinner) View() string {
	if !s.active || len(s.frames) == 0 {
		return ""
	}
	return s.style.Render(s.frames[s.current])
}

// SetStyle updates the spinner's style
func (s *Spinner) SetStyle(style lipgloss.Style) {
	s.style = style
}

// IsActive returns whether the spinner is currently active
func (s *Spinner) IsActive() bool {
	return s.active
}
