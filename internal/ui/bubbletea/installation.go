package bubbletea

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lvcasx1/quikgit/internal/install"
)

type InstallationModel struct {
	app          *Application
	repositories []string // Paths to cloned repositories
	progressBars map[string]progress.Model
	installMgr   *install.Manager
	ctx          context.Context
	cancel       context.CancelFunc

	// Progress tracking
	completed    map[string]bool
	errors       map[string]error
	statuses     map[string]string
	allCompleted bool
	successCount int
	errorCount   int
	started      bool
}

func NewInstallationModel(app *Application) *InstallationModel {
	ctx, cancel := context.WithCancel(context.Background())

	// Use the cloned paths from successful cloning
	repoPaths := app.clonedPaths

	model := &InstallationModel{
		app:          app,
		repositories: repoPaths,
		progressBars: make(map[string]progress.Model),
		ctx:          ctx,
		cancel:       cancel,
		completed:    make(map[string]bool),
		errors:       make(map[string]error),
		statuses:     make(map[string]string),
		started:      false,
	}

	// Initialize progress bars for each repository
	for _, repoPath := range model.repositories {
		repoName := filepath.Base(repoPath)
		progressBar := progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(50),
		)
		model.progressBars[repoName] = progressBar
		model.statuses[repoName] = "Waiting..."
	}

	return model
}

func (m *InstallationModel) Init() tea.Cmd {
	// Return command to start the installation process
	return m.startInstallationProcess()
}

func (m *InstallationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if !m.allCompleted {
				m.cancel()
				return m, m.app.NavigateTo(StateMainMenu)
			}
		case "enter":
			if m.allCompleted {
				// Installation complete, return to main menu
				m.app.message = fmt.Sprintf("Installation completed! %d successful, %d failed", m.successCount, m.errorCount)
				return m, m.app.NavigateTo(StateMainMenu)
			}
		case "esc", "q":
			if m.allCompleted {
				return m, m.app.NavigateTo(StateMainMenu)
			}
		}

	case InstallProgressMsg:
		return m.handleProgressUpdate(msg)

	case InstallCompleteMsg:
		m.allCompleted = true
		return m, nil

	case InstallStartMsg:
		if !m.started {
			m.started = true
			return m, m.monitorProgress()
		}
	}

	// Update progress bars
	var cmds []tea.Cmd
	for repoName, bar := range m.progressBars {
		updatedBar, cmd := bar.Update(msg)
		if updatedBar != nil {
			if progressBar, ok := updatedBar.(progress.Model); ok {
				m.progressBars[repoName] = progressBar
			}
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *InstallationModel) View() string {
	// Use full screen dimensions with fallback
	width := m.app.width
	height := m.app.height - 3
	if width == 0 {
		width = 120
	}
	if height <= 0 {
		height = 30
	}

	var sections []string

	// Title
	title := fmt.Sprintf("󰏖 Installing Dependencies (%d repositories)", len(m.repositories))
	titleStyle := TitleStyle.Width(width - 20)
	sections = append(sections, titleStyle.Render(title))

	// Progress section
	progressSection := m.renderProgress(width)
	sections = append(sections, progressSection)

	// Summary and instructions
	summarySection := m.renderSummary(width)
	sections = append(sections, summarySection)

	// Join all sections
	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m *InstallationModel) renderProgress(width int) string {
	progressStyle := lipgloss.NewStyle().
		Padding(2, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Width(width - 20)

	if len(m.repositories) == 0 {
		noReposStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Align(lipgloss.Center)
		return progressStyle.Render(noReposStyle.Render("No repositories to install dependencies for"))
	}

	var progressItems []string

	for _, repoPath := range m.repositories {
		repoName := filepath.Base(repoPath)
		var itemParts []string

		// Repository name with icon
		repoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
		itemParts = append(itemParts, repoStyle.Render("󰏖 "+repoName))

		// Progress bar with calculated progress
		if bar, exists := m.progressBars[repoName]; exists {
			progressPercent := 0.0
			if m.completed[repoName] {
				progressPercent = 1.0
			} else if m.errors[repoName] != nil {
				progressPercent = 0.0 // Reset on error
			} else {
				// Calculate progress based on status
				status := m.statuses[repoName]
				switch {
				case strings.Contains(status, "Detecting"):
					progressPercent = 0.1
				case strings.Contains(status, "Installing"):
					progressPercent = 0.5
				case strings.Contains(status, "Running"):
					progressPercent = 0.7
				case strings.Contains(status, "Completing"):
					progressPercent = 0.9
				case strings.Contains(status, "Waiting"):
					progressPercent = 0.0
				default:
					progressPercent = 0.2
				}
			}

			progressView := bar.ViewAs(progressPercent)
			itemParts = append(itemParts, progressView)
		}

		// Status or error
		if err, hasError := m.errors[repoName]; hasError {
			errorStyle := ErrorStyle.Copy().Width(width - 40)
			itemParts = append(itemParts, errorStyle.Render("󰅖 Error: "+err.Error()))
		} else if m.completed[repoName] {
			successStyle := SuccessStyle
			itemParts = append(itemParts, successStyle.Render("󰄬 Dependencies installed"))
		} else {
			statusStyle := InfoStyle
			status := m.statuses[repoName]
			itemParts = append(itemParts, statusStyle.Render("🔄 "+status))
		}

		// Join this repository's info
		repoItem := lipgloss.JoinVertical(lipgloss.Left, itemParts...)
		progressItems = append(progressItems, repoItem)
		progressItems = append(progressItems, "") // Empty line between repos
	}

	progressContent := lipgloss.JoinVertical(lipgloss.Left, progressItems...)
	return progressStyle.Render(progressContent)
}

func (m *InstallationModel) renderSummary(width int) string {
	var summaryParts []string

	if m.allCompleted {
		// Final summary
		summaryStyle := lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Width(width - 60).
			Align(lipgloss.Center).
			MarginTop(1)

		summary := fmt.Sprintf("Summary: %d successful, %d failed", m.successCount, m.errorCount)
		summaryParts = append(summaryParts, summaryStyle.Render(summary))

		// Instructions
		instructionStyle := SuccessStyle.Copy().
			Align(lipgloss.Center).
			MarginTop(1)
		summaryParts = append(summaryParts, instructionStyle.Render("🎉 Press Enter to continue"))
	} else {
		// In-progress summary
		progressText := fmt.Sprintf("Progress: %d/%d repositories processed",
			m.successCount+m.errorCount, len(m.repositories))

		progressStyle := InfoStyle.Copy().
			Align(lipgloss.Center).
			MarginTop(1)
		summaryParts = append(summaryParts, progressStyle.Render(progressText))

		// Cancel instruction
		cancelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Align(lipgloss.Center)
		summaryParts = append(summaryParts, cancelStyle.Render("Press Ctrl+C to cancel"))
	}

	return lipgloss.JoinVertical(lipgloss.Center, summaryParts...)
}

func (m *InstallationModel) startInstallationProcess() tea.Cmd {
	return func() tea.Msg {
		if len(m.repositories) == 0 {
			return InstallProgressMsg{
				Repository: "system",
				Error:      fmt.Errorf("no repositories to install dependencies for"),
			}
		}

		// Create installation manager
		m.installMgr = install.NewManager(3, 10*time.Minute)
		m.installMgr.SetSkipOnError(true)

		return InstallStartMsg{}
	}
}

func (m *InstallationModel) monitorProgress() tea.Cmd {
	return func() tea.Msg {
		// Start the actual installation in a separate goroutine
		go func() {
			m.installMgr.InstallDependencies(m.ctx, m.repositories)
		}()

		// Monitor the progress channel
		progressChan := m.installMgr.GetProgressChannel()

		// Wait for the first progress update and return it
		select {
		case progress, ok := <-progressChan:
			if !ok {
				// Channel closed, all done
				return InstallCompleteMsg{}
			}

			return InstallProgressMsg{
				Repository:  progress.Repository,
				ProjectType: progress.ProjectType,
				Status:      progress.Status,
				Error:       progress.Error,
				Completed:   progress.Completed,
			}

		case <-m.ctx.Done():
			// Context cancelled
			return InstallCompleteMsg{}

		case <-time.After(5 * time.Second):
			// Timeout - something went wrong
			return InstallProgressMsg{
				Repository: "system",
				Error:      fmt.Errorf("installation timeout - no progress received"),
			}
		}
	}
}

func (m *InstallationModel) handleProgressUpdate(msg InstallProgressMsg) (tea.Model, tea.Cmd) {
	// Update status
	if msg.Repository != "system" {
		repoName := filepath.Base(msg.Repository)
		m.statuses[repoName] = msg.Status
	}

	// Handle completion
	if msg.Completed {
		if msg.Repository != "system" {
			repoName := filepath.Base(msg.Repository)
			m.completed[repoName] = true

			if msg.Error != nil {
				m.errors[repoName] = msg.Error
				m.errorCount++
			} else {
				m.successCount++
			}
		}

		// Check if all repositories are done
		if m.successCount+m.errorCount >= len(m.repositories) {
			return m, func() tea.Msg { return InstallCompleteMsg{} }
		}
	}

	// Handle system-level errors
	if msg.Repository == "system" && msg.Error != nil {
		m.allCompleted = true
		// Don't set app.error as it causes uncentered error display
		return m, nil
	}

	// Continue monitoring for more progress updates
	return m, m.monitorProgress()
}

// Message types for installation process
type InstallStartMsg struct{}

type InstallProgressMsg struct {
	Repository  string
	ProjectType string
	Status      string
	Error       error
	Completed   bool
}

type InstallCompleteMsg struct{}
