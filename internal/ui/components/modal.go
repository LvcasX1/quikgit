package components

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

// Modal represents a modal dialog component
type Modal struct {
	title         string
	content       string
	visible       bool
	width         int
	height        int
	spring        harmonica.Spring
	opacity       float64
	targetOpacity float64

	// Styles
	overlayStyle lipgloss.Style
	modalStyle   lipgloss.Style
	titleStyle   lipgloss.Style
	contentStyle lipgloss.Style

	// Animation state
	animating bool
}

// ModalAnimationMsg is sent during modal animations
type ModalAnimationMsg struct{}

// NewModal creates a new modal component
func NewModal(title, content string) *Modal {
	return &Modal{
		title:   title,
		content: content,
		width:   60,
		height:  20,
		spring:  harmonica.NewSpring(harmonica.FPS(60), 15.0, 0.8),
		overlayStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("0")).
			Foreground(lipgloss.Color("255")),
		modalStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Background(lipgloss.Color("0")).
			Padding(1, 2).
			MarginTop(2).
			MarginBottom(2),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Align(lipgloss.Center).
			MarginBottom(1),
		contentStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Align(lipgloss.Left),
	}
}

// Show displays the modal with animation
func (m *Modal) Show() tea.Cmd {
	if m.visible {
		return nil
	}

	m.visible = true
	m.animating = true
	m.targetOpacity = 1.0
	return m.animate()
}

// Hide conceals the modal with animation
func (m *Modal) Hide() tea.Cmd {
	if !m.visible {
		return nil
	}

	m.animating = true
	m.targetOpacity = 0.0
	return m.animate()
}

// SetContent updates the modal content
func (m *Modal) SetContent(title, content string) {
	m.title = title
	m.content = content
}

// SetSize updates the modal dimensions
func (m *Modal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsVisible returns whether the modal is visible
func (m *Modal) IsVisible() bool {
	return m.visible
}

// animate returns a command for the next animation frame
func (m *Modal) animate() tea.Cmd {
	if !m.animating {
		return nil
	}

	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return ModalAnimationMsg{}
	})
}

// Update handles modal messages
func (m *Modal) Update(msg tea.Msg) (*Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Adjust modal size based on window size
		maxWidth := msg.Width - 4
		maxHeight := msg.Height - 4

		if m.width > maxWidth {
			m.width = maxWidth
		}
		if m.height > maxHeight {
			m.height = maxHeight
		}

	case ModalAnimationMsg:
		if m.animating {
			// Update spring animation
			_, velocity := m.spring.Update(m.targetOpacity, m.opacity, 1.0/60.0)
			m.opacity += velocity * (1.0 / 60.0)

			// Check if animation is complete
			if abs(m.opacity-m.targetOpacity) < 0.01 {
				m.opacity = m.targetOpacity
				m.animating = false

				if m.targetOpacity == 0.0 {
					m.visible = false
				}
			} else {
				return m, m.animate()
			}
		}

	case tea.KeyMsg:
		if m.visible {
			switch msg.String() {
			case "esc", "q":
				return m, m.Hide()
			}
		}
	}

	return m, nil
}

// View renders the modal
func (m *Modal) View() string {
	if !m.visible {
		return ""
	}

	// Create the modal content
	titleStr := m.titleStyle.Width(m.width - 4).Render(m.title)

	// Wrap content to fit modal width
	contentLines := strings.Split(m.content, "\n")
	var wrappedContent strings.Builder

	for _, line := range contentLines {
		if len(line) <= m.width-6 {
			wrappedContent.WriteString(line + "\n")
		} else {
			// Simple word wrapping
			words := strings.Fields(line)
			currentLine := ""

			for _, word := range words {
				if len(currentLine)+len(word)+1 <= m.width-6 {
					if currentLine != "" {
						currentLine += " "
					}
					currentLine += word
				} else {
					if currentLine != "" {
						wrappedContent.WriteString(currentLine + "\n")
					}
					currentLine = word
				}
			}

			if currentLine != "" {
				wrappedContent.WriteString(currentLine + "\n")
			}
		}
	}

	contentStr := m.contentStyle.Width(m.width - 4).Render(strings.TrimSuffix(wrappedContent.String(), "\n"))

	// Combine title and content
	modalContent := titleStr + "\n" + contentStr

	// Apply modal styling
	modal := m.modalStyle.Width(m.width).Height(m.height).Render(modalContent)

	// Apply opacity effect (simplified for terminal)
	if m.opacity < 1.0 {
		// Could implement opacity effect here if needed
	}

	return modal
}

// Center centers the modal on screen
func (m *Modal) Center(screenWidth, screenHeight int) lipgloss.Style {
	leftMargin := (screenWidth - m.width) / 2
	topMargin := (screenHeight - m.height) / 2

	return lipgloss.NewStyle().
		MarginLeft(leftMargin).
		MarginTop(topMargin)
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Predefined modal types
func NewInfoModal(title, message string) *Modal {
	modal := NewModal(title, message)
	modal.modalStyle = modal.modalStyle.BorderForeground(lipgloss.Color("117"))
	return modal
}

func NewSuccessModal(title, message string) *Modal {
	modal := NewModal(title, message)
	modal.modalStyle = modal.modalStyle.BorderForeground(lipgloss.Color("82"))
	return modal
}

func NewErrorModal(title, message string) *Modal {
	modal := NewModal(title, message)
	modal.modalStyle = modal.modalStyle.BorderForeground(lipgloss.Color("196"))
	return modal
}

func NewWarningModal(title, message string) *Modal {
	modal := NewModal(title, message)
	modal.modalStyle = modal.modalStyle.BorderForeground(lipgloss.Color("214"))
	return modal
}
