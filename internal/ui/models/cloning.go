package models

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/quikgit/internal/github"
)

type CloningModel struct {
	repositories []*github.Repository
	progress     map[string]*progress.Model
	statuses     map[string]string
	errors       map[string]error
	completed    map[string]bool

	cloneManager *github.CloneManager
	ctx          context.Context
	cancel       context.CancelFunc

	allCompleted bool
	successCount int
	errorCount   int

	// Styles
	titleStyle    lipgloss.Style
	successStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	progressStyle lipgloss.Style
}

func NewCloningModel() *CloningModel {
	return &CloningModel{
		progress:  make(map[string]*progress.Model),
		statuses:  make(map[string]string),
		errors:    make(map[string]error),
		completed: make(map[string]bool),
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
	}
}

func (m *CloningModel) SetRepositories(repos []*github.Repository) {
	m.repositories = repos
	m.progress = make(map[string]*progress.Model)
	m.statuses = make(map[string]string)
	m.errors = make(map[string]error)
	m.completed = make(map[string]bool)
	m.allCompleted = false
	m.successCount = 0
	m.errorCount = 0

	// Initialize progress bars for each repository
	for _, repo := range repos {
		p := progress.New(progress.WithDefaultGradient())
		p.Width = 50
		m.progress[repo.FullName] = &p
		m.statuses[repo.FullName] = "Waiting"
	}
}

func (m *CloningModel) Init() tea.Cmd {
	return m.startCloning()
}

func (m *CloningModel) Update(msg tea.Msg) (*CloningModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" && !m.allCompleted {
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		}
		if msg.String() == "enter" && m.allCompleted {
			// Proceed to installation
			var clonedPaths []string
			for _, repo := range m.repositories {
				if m.completed[repo.FullName] && m.errors[repo.FullName] == nil {
					wd, _ := os.Getwd()
					clonedPaths = append(clonedPaths, filepath.Join(wd, repo.Name))
				}
			}
			return m, func() tea.Msg {
				return AppMsg{Type: "clone_completed", Payload: clonedPaths}
			}
		}

	case github.CloneProgress:
		if prog, exists := m.progress[msg.Repository]; exists {
			cmd := prog.SetPercent(msg.Progress)
			cmds = append(cmds, cmd)
		}

		m.statuses[msg.Repository] = msg.Status

		if msg.Completed {
			m.completed[msg.Repository] = true
			if msg.Error != nil {
				m.errors[msg.Repository] = msg.Error
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

func (m *CloningModel) View() string {
	var content strings.Builder

	content.WriteString(m.titleStyle.Render("Cloning Repositories"))
	content.WriteString("\n\n")

	for _, repo := range m.repositories {
		content.WriteString(fmt.Sprintf("📁 %s\n", repo.FullName))

		if prog, exists := m.progress[repo.FullName]; exists {
			content.WriteString(m.progressStyle.Render(prog.View()))
		}

		status := m.statuses[repo.FullName]
		if err, hasError := m.errors[repo.FullName]; hasError {
			content.WriteString("\n")
			content.WriteString(m.errorStyle.Render(fmt.Sprintf("❌ %s: %s", status, err.Error())))
		} else if m.completed[repo.FullName] {
			content.WriteString("\n")
			content.WriteString(m.successStyle.Render(fmt.Sprintf("✅ %s", status)))
		} else {
			content.WriteString(fmt.Sprintf("\n🔄 %s", status))
		}

		content.WriteString("\n\n")
	}

	// Summary
	if m.allCompleted {
		content.WriteString("─────────────────────────────────────\n")
		content.WriteString(fmt.Sprintf("Summary: %d successful, %d failed\n", m.successCount, m.errorCount))

		if m.successCount > 0 {
			content.WriteString("\n")
			content.WriteString(m.successStyle.Render("Press Enter to install dependencies"))
		}
	} else {
		content.WriteString("Press Ctrl+C to cancel cloning")
	}

	return content.String()
}

func (m *CloningModel) startCloning() tea.Cmd {
	return func() tea.Msg {
		wd, err := os.Getwd()
		if err != nil {
			return AppMsg{Type: "error", Payload: fmt.Errorf("failed to get working directory: %w", err)}
		}

		// TODO: Get token from auth manager
		m.cloneManager = github.NewCloneManager("", wd)
		m.ctx, m.cancel = context.WithCancel(context.Background())

		// Start cloning repositories concurrently
		go func() {
			m.cloneManager.CloneRepositories(m.ctx, m.repositories, 3)
			m.cloneManager.Wait()
		}()

		return nil
	}
}
