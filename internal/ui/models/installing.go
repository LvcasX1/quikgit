package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/quikgit/internal/install"
)

type InstallingModel struct {
	repositories []string
	progress     map[string]*progress.Model
	statuses     map[string]string
	outputs      map[string][]string
	errors       map[string]error
	completed    map[string]bool
	results      map[string]install.InstallResult

	installManager *install.Manager
	ctx            context.Context
	cancel         context.CancelFunc

	allCompleted bool
	successCount int
	errorCount   int

	// UI state
	showDetails  bool
	selectedRepo string

	// Styles
	titleStyle    lipgloss.Style
	successStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	progressStyle lipgloss.Style
	detailStyle   lipgloss.Style
}

func NewInstallingModel() *InstallingModel {
	return &InstallingModel{
		progress:  make(map[string]*progress.Model),
		statuses:  make(map[string]string),
		outputs:   make(map[string][]string),
		errors:    make(map[string]error),
		completed: make(map[string]bool),
		results:   make(map[string]install.InstallResult),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Align(lipgloss.Center),
		successStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
		progressStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")),
		detailStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			MarginTop(1),
	}
}

func (m *InstallingModel) SetRepositories(repos []string) {
	m.repositories = repos
	m.progress = make(map[string]*progress.Model)
	m.statuses = make(map[string]string)
	m.outputs = make(map[string][]string)
	m.errors = make(map[string]error)
	m.completed = make(map[string]bool)
	m.results = make(map[string]install.InstallResult)
	m.allCompleted = false
	m.successCount = 0
	m.errorCount = 0

	// Initialize progress bars for each repository
	for _, repo := range repos {
		p := progress.New(progress.WithDefaultGradient())
		p.Width = 50
		repoName := getRepoName(repo)
		m.progress[repoName] = &p
		m.statuses[repoName] = "Detecting project type"
		m.outputs[repoName] = []string{}
	}
}

func (m *InstallingModel) Init() tea.Cmd {
	return m.startInstalling()
}

func (m *InstallingModel) Update(msg tea.Msg) (*InstallingModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if !m.allCompleted && m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit

		case "enter":
			if m.allCompleted {
				return m, func() tea.Msg {
					return AppMsg{Type: "install_completed", Payload: m.results}
				}
			}

		case "d":
			m.showDetails = !m.showDetails

		case "j", "down":
			if m.showDetails && len(m.repositories) > 0 {
				currentIndex := m.getSelectedRepoIndex()
				if currentIndex < len(m.repositories)-1 {
					m.selectedRepo = getRepoName(m.repositories[currentIndex+1])
				}
			}

		case "k", "up":
			if m.showDetails && len(m.repositories) > 0 {
				currentIndex := m.getSelectedRepoIndex()
				if currentIndex > 0 {
					m.selectedRepo = getRepoName(m.repositories[currentIndex-1])
				}
			}
		}

	case install.InstallProgress:
		repoName := msg.Repository

		if prog, exists := m.progress[repoName]; exists {
			// Update progress based on status
			var percent float64
			switch msg.Status {
			case "Detecting project type":
				percent = 0.1
			case "Installing dependencies":
				percent = 0.3
			case "Running commands":
				percent = 0.6
			case "Completed":
				percent = 1.0
			case "Failed":
				percent = 1.0
			default:
				if strings.Contains(msg.Status, "Running:") {
					percent = 0.7
				}
			}

			cmd := prog.SetPercent(percent)
			cmds = append(cmds, cmd)
		}

		m.statuses[repoName] = msg.Status

		if msg.Output != "" {
			m.outputs[repoName] = append(m.outputs[repoName], msg.Output)
		}

		if msg.Completed {
			m.completed[repoName] = true
			if msg.Error != nil {
				m.errors[repoName] = msg.Error
				m.errorCount++
			} else {
				m.successCount++
			}

			// Check if all are completed
			if m.successCount+m.errorCount == len(m.repositories) {
				m.allCompleted = true
			}
		}
	}

	// Update all progress bars
	for _, prog := range m.progress {
		newProg, newCmd := prog.Update(msg)
		*prog = newProg.(progress.Model)
		if newCmd != nil {
			cmds = append(cmds, newCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *InstallingModel) View() string {
	var content strings.Builder

	content.WriteString(m.titleStyle.Render("Installing Dependencies"))
	content.WriteString("\n\n")

	for _, repoPath := range m.repositories {
		repoName := getRepoName(repoPath)

		content.WriteString(fmt.Sprintf("📦 %s\n", repoName))

		if prog, exists := m.progress[repoName]; exists {
			content.WriteString(m.progressStyle.Render(prog.View()))
		}

		status := m.statuses[repoName]
		if err, hasError := m.errors[repoName]; hasError {
			content.WriteString("\n")
			content.WriteString(m.errorStyle.Render(fmt.Sprintf("❌ %s: %s", status, err.Error())))
		} else if m.completed[repoName] {
			content.WriteString("\n")
			content.WriteString(m.successStyle.Render(fmt.Sprintf("✅ %s", status)))
		} else {
			content.WriteString(fmt.Sprintf("\n🔄 %s", status))
		}

		content.WriteString("\n")

		// Show details if enabled and this is the selected repo
		if m.showDetails && (m.selectedRepo == repoName || m.selectedRepo == "") {
			m.selectedRepo = repoName
			content.WriteString(m.renderDetails(repoName))
		}

		content.WriteString("\n")
	}

	// Summary
	if m.allCompleted {
		content.WriteString("─────────────────────────────────────\n")
		content.WriteString(fmt.Sprintf("Summary: %d successful, %d failed\n", m.successCount, m.errorCount))
		content.WriteString("\n")
		content.WriteString(m.successStyle.Render("Press Enter to continue"))
	} else {
		content.WriteString("Press 'd' to toggle details • Ctrl+C to cancel")
	}

	return content.String()
}

func (m *InstallingModel) renderDetails(repoName string) string {
	var details strings.Builder

	// Show recent output
	outputs := m.outputs[repoName]
	if len(outputs) > 0 {
		details.WriteString("Recent output:\n")
		start := len(outputs) - 5
		if start < 0 {
			start = 0
		}

		for i := start; i < len(outputs); i++ {
			line := outputs[i]
			if len(line) > 80 {
				line = line[:77] + "..."
			}
			details.WriteString(fmt.Sprintf("  %s\n", line))
		}
	}

	// Show result if completed
	if result, exists := m.results[repoName]; exists {
		details.WriteString(fmt.Sprintf("\nProject type: %s\n", result.ProjectType))
		details.WriteString(fmt.Sprintf("Duration: %v\n", result.Duration))

		for _, cmd := range result.Commands {
			if cmd.Success {
				details.WriteString(fmt.Sprintf("✅ %s (%v)\n", cmd.Command, cmd.Duration))
			} else {
				details.WriteString(fmt.Sprintf("❌ %s: %s\n", cmd.Command, cmd.Error))
			}
		}
	}

	return m.detailStyle.Render(details.String())
}

func (m *InstallingModel) getSelectedRepoIndex() int {
	for i, repoPath := range m.repositories {
		if getRepoName(repoPath) == m.selectedRepo {
			return i
		}
	}
	return 0
}

func (m *InstallingModel) startInstalling() tea.Cmd {
	return func() tea.Msg {
		m.installManager = install.NewManager(3, 10*time.Minute)
		m.installManager.SetSkipOnError(true)
		m.ctx, m.cancel = context.WithCancel(context.Background())

		// Start installation process
		go func() {
			results, err := m.installManager.InstallDependencies(m.ctx, m.repositories)
			if err != nil {
				// Handle error
				return
			}

			// Store results
			for _, result := range results {
				m.results[result.Repository] = result
			}
		}()

		return nil
	}
}

func getRepoName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
