package bubbletea

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lvcasx1/quikgit/internal/github"
	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

type CloningModel struct {
	app          *Application
	repositories []*ghClient.Repository
	progressBars map[string]progress.Model
	cloneManager *github.CloneManager
	ctx          context.Context
	cancel       context.CancelFunc

	// Progress tracking
	completed    map[string]bool
	errors       map[string]error
	statuses     map[string]string
	allCompleted bool
	successCount int
	errorCount   int

	// UI state
	targetDir    string
	started      bool
	cloneStarted bool // New flag to prevent multiple clone starts
	autoSubdirs  bool // Track if subdirectories were automatically enabled

	// Animation enhancement
	startTime         time.Time
	minDuration       time.Duration
	animationTickers  map[string]*time.Ticker
	animationProgress map[string]float64
}

func NewCloningModel(app *Application) *CloningModel {
	ctx, cancel := context.WithCancel(context.Background())

	model := &CloningModel{
		app:          app,
		repositories: app.selectedRepos,
		progressBars: make(map[string]progress.Model),
		ctx:          ctx,
		cancel:       cancel,
		completed:    make(map[string]bool),
		errors:       make(map[string]error),
		statuses:     make(map[string]string),
		started:      false,
		cloneStarted: false,
		// Animation enhancement
		minDuration:       2 * time.Second, // Minimum 2 seconds for rich experience
		animationTickers:  make(map[string]*time.Ticker),
		animationProgress: make(map[string]float64),
	}

	// Initialize progress bars for each repository
	for _, repo := range model.repositories {
		progressBar := progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(50),
		)
		model.progressBars[repo.FullName] = progressBar
		model.statuses[repo.FullName] = "󰔟 Preparing..."
		model.animationProgress[repo.FullName] = 0.0
	}

	return model
}

func (m *CloningModel) Init() tea.Cmd {
	// Return command to start the cloning process
	return m.startCloningProcess()
}

func (m *CloningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if m.successCount > 0 {
					// Set successful paths for installation
					m.app.clonedPaths = m.getSuccessfullyClonedPaths()
					m.app.message = fmt.Sprintf("Successfully cloned %d repositories", m.successCount)
					return m, m.app.NavigateTo(StateInstalling)
				} else {
					return m, m.app.NavigateTo(StateMainMenu)
				}
			}
		case "esc", "q":
			if m.allCompleted {
				return m, m.app.NavigateTo(StateMainMenu)
			}
		}

	case CloneProgressMsg:
		return m.handleProgressUpdate(msg)

	case CloneCompleteMsg:
		m.allCompleted = true
		return m, nil

	case CloneStartMsg:
		if !m.started {
			m.started = true
			m.startTime = time.Now() // Record start time for minimum duration
			return m, tea.Batch(m.monitorProgress(), m.startAnimationTickers())
		}
		// Ignore duplicate start messages
		return m, nil

	case AnimationTickMsg:
		return m.handleAnimationTick(msg)

	case DurationCheckMsg:
		return m.handleDurationCheck()
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

func (m *CloningModel) View() string {
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
	title := fmt.Sprintf("󰓁 Cloning Repositories (%d)", len(m.repositories))
	titleStyle := TitleStyle.Width(width - 20)
	sections = append(sections, titleStyle.Render(title))

	// Show subdirectory info if automatically enabled
	if m.autoSubdirs {
		subdirInfo := "󰉋 Using owner/repo subdirectories due to name conflicts"
		subdirStyle := InfoStyle.Copy().
			Align(lipgloss.Center).
			Italic(true).
			MarginBottom(1)
		sections = append(sections, subdirStyle.Render(subdirInfo))
	}

	// Progress section
	progressSection := m.renderProgress()
	sections = append(sections, progressSection)

	// Summary and instructions
	summarySection := m.renderSummary()
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

func (m *CloningModel) renderProgress() string {
	// Calculate responsive width
	width := m.app.width
	if width == 0 {
		width = 120
	}
	progressWidth := width - 20
	if progressWidth < 80 {
		progressWidth = 80
	}

	progressStyle := lipgloss.NewStyle().
		Padding(2, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Width(progressWidth)

	var progressItems []string

	for i, repo := range m.repositories {
		var itemParts []string

		// Repository name with icon
		repoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
		itemParts = append(itemParts, repoStyle.Render("󰉋 "+repo.FullName))

		// Progress bar with enhanced animation
		if bar, exists := m.progressBars[repo.FullName]; exists {
			progressPercent := 0.0
			if m.completed[repo.FullName] {
				progressPercent = 1.0
			} else if m.errors[repo.FullName] != nil {
				progressPercent = 0.0 // Reset on error
			} else {
				// Use smooth animation progress instead of stepped progress
				animProgress := m.animationProgress[repo.FullName]
				if animProgress > 0 {
					progressPercent = animProgress
				} else {
					// Fallback to status-based progress for very start
					status := m.statuses[repo.FullName]
					switch {
					case strings.Contains(status, "Initializing"):
						progressPercent = 0.05
					case strings.Contains(status, "Waiting"):
						progressPercent = 0.01
					default:
						progressPercent = 0.1
					}
				}
			}

			progressView := bar.ViewAs(progressPercent)
			itemParts = append(itemParts, progressView)
		}

		// Status or error
		if err, hasError := m.errors[repo.FullName]; hasError {
			errorStyle := ErrorStyle.Copy().Width(80)
			itemParts = append(itemParts, errorStyle.Render("󰅖 Error: "+err.Error()))
		} else if m.completed[repo.FullName] {
			successStyle := SuccessStyle
			itemParts = append(itemParts, successStyle.Render("󰄬 Completed"))
		} else {
			statusStyle := InfoStyle
			status := m.statuses[repo.FullName]
			// Don't add extra icon if status already has one
			if strings.HasPrefix(status, "󰔟") || strings.HasPrefix(status, "󰦖") ||
				strings.HasPrefix(status, "󰓂") || strings.HasPrefix(status, "󰇚") ||
				strings.HasPrefix(status, "󰇘") || strings.HasPrefix(status, "󰧑") {
				itemParts = append(itemParts, statusStyle.Render(status))
			} else {
				itemParts = append(itemParts, statusStyle.Render("󰔟 "+status))
			}
		}

		// Join this repository's info
		repoItem := lipgloss.JoinVertical(lipgloss.Left, itemParts...)
		progressItems = append(progressItems, repoItem)

		// Add separator between repos only if not the last one
		if i < len(m.repositories)-1 {
			progressItems = append(progressItems, "") // Empty line between repos
		}
	}

	progressContent := lipgloss.JoinVertical(lipgloss.Left, progressItems...)
	return progressStyle.Render(progressContent)
}

func (m *CloningModel) renderSummary() string {
	// Calculate responsive width
	width := m.app.width
	if width == 0 {
		width = 120
	}
	var summaryParts []string

	if m.allCompleted {
		// Final summary
		summaryWidth := width - 60
		if summaryWidth < 40 {
			summaryWidth = 40
		}

		summaryStyle := lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Width(summaryWidth).
			Align(lipgloss.Center).
			MarginTop(1)

		summary := fmt.Sprintf("Summary: %d successful, %d failed", m.successCount, m.errorCount)
		summaryParts = append(summaryParts, summaryStyle.Render(summary))

		// Instructions
		if m.successCount > 0 {
			instructionStyle := SuccessStyle.Copy().
				Align(lipgloss.Center).
				MarginTop(1)
			summaryParts = append(summaryParts, instructionStyle.Render("󰊤 Press Enter to install dependencies"))
		} else {
			instructionStyle := ErrorStyle.Copy().
				Align(lipgloss.Center).
				MarginTop(1)
			summaryParts = append(summaryParts, instructionStyle.Render("󰅖 No repositories cloned. Press Enter to return."))
		}
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

func (m *CloningModel) startCloningProcess() tea.Cmd {
	return func() tea.Msg {
		// Determine target directory
		var targetDir string
		if m.app.config.Clone.UseCurrentDir {
			wd, err := os.Getwd()
			if err != nil {
				return CloneProgressMsg{
					Repository: "system",
					Error:      fmt.Errorf("failed to get working directory: %w", err),
				}
			}
			targetDir = wd
		} else if m.app.config.Clone.DefaultPath != "" {
			targetDir = m.app.config.Clone.DefaultPath
		} else {
			wd, _ := os.Getwd()
			targetDir = wd
		}

		// Ensure target directory exists
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return CloneProgressMsg{
				Repository: "system",
				Error:      fmt.Errorf("failed to create target directory: %w", err),
			}
		}

		m.targetDir = targetDir

		// Get authentication token
		token := ""
		if m.app.authManager != nil {
			token = m.app.authManager.GetToken()
		}
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		// Check for repository name conflicts
		createSubdirs := m.app.config.Clone.CreateSubdirs
		if !createSubdirs {
			// Detect name conflicts - if multiple repos have the same name, enable subdirectories
			repoNames := make(map[string]bool)
			for _, repo := range m.repositories {
				name := repo.Name
				if repoNames[name] {
					// Found duplicate name, enable subdirectories
					createSubdirs = true
					m.autoSubdirs = true // Track that this was automatically enabled
					break
				}
				repoNames[name] = true
			}
		}

		// Create clone manager
		m.cloneManager = github.NewCloneManager(token, targetDir)
		m.cloneManager.SetCreateSubdirs(createSubdirs)

		return CloneStartMsg{}
	}
}

func (m *CloningModel) monitorProgress() tea.Cmd {
	return func() tea.Msg {
		// Start the actual cloning in a separate goroutine only once
		if m.cloneManager != nil && !m.cloneStarted {
			m.cloneStarted = true
			go m.cloneManager.CloneRepositories(m.ctx, m.repositories, 3)
		}

		// Get the progress channel
		if m.cloneManager != nil {
			progressChan := m.cloneManager.GetProgressChannel()

			// Wait for the next progress update
			select {
			case progress, ok := <-progressChan:
				if !ok {
					// Channel closed, all done
					return CloneCompleteMsg{}
				}

				return CloneProgressMsg{
					Repository: progress.Repository,
					Status:     progress.Status,
					Progress:   progress.Progress,
					Error:      progress.Error,
					Completed:  progress.Completed,
				}

			case <-m.ctx.Done():
				// Context cancelled
				return CloneCompleteMsg{}

			case <-time.After(10 * time.Second):
				// Longer timeout to prevent premature completion
				return CloneProgressMsg{
					Repository: "system",
					Error:      fmt.Errorf("cloning timeout - no progress received"),
				}
			}
		}

		return CloneCompleteMsg{}
	}
}

func (m *CloningModel) handleProgressUpdate(msg CloneProgressMsg) (tea.Model, tea.Cmd) {
	// Update status
	if msg.Repository != "system" {
		m.statuses[msg.Repository] = msg.Status
	}

	// Handle completion
	if msg.Completed {
		if msg.Repository != "system" {
			m.completed[msg.Repository] = true

			// Handle "already exists" as success, not error
			if msg.Error != nil {
				if strings.Contains(msg.Error.Error(), "already exists") {
					m.statuses[msg.Repository] = "Already exists (skipped)"
					m.successCount++
					// Don't add to errors map - treat as success for installation
				} else {
					m.errors[msg.Repository] = msg.Error
					m.errorCount++
				}
			} else {
				m.successCount++
			}
		}

		// Check if all repositories are done
		if m.successCount+m.errorCount >= len(m.repositories) {
			// Check if minimum duration has been met
			elapsed := time.Since(m.startTime)
			if elapsed < m.minDuration {
				// Wait for remaining duration before allowing completion
				remaining := m.minDuration - elapsed
				return m, tea.Tick(remaining, func(t time.Time) tea.Msg {
					return CloneCompleteMsg{}
				})
			}
			return m, func() tea.Msg { return CloneCompleteMsg{} }
		}
	}

	// Handle system-level errors
	if msg.Repository == "system" && msg.Error != nil {
		m.allCompleted = true
		// Don't set app.error as it causes uncentered error display
		return m, nil
	}

	// Only continue monitoring if not completed and not a system error
	if !m.allCompleted && msg.Repository != "system" {
		return m, m.monitorProgress()
	}

	return m, nil
}

func (m *CloningModel) getSuccessfullyClonedPaths() []string {
	var paths []string

	for _, repo := range m.repositories {
		if m.completed[repo.FullName] && m.errors[repo.FullName] == nil {
			var repoPath string
			if m.app.config.Clone.CreateSubdirs {
				repoPath = filepath.Join(m.targetDir, repo.Owner, repo.Name)
			} else {
				repoPath = filepath.Join(m.targetDir, repo.Name)
			}
			paths = append(paths, repoPath)
		}
	}

	return paths
}

// startAnimationTickers starts smooth animation tickers for each repository
func (m *CloningModel) startAnimationTickers() tea.Cmd {
	var cmds []tea.Cmd
	for _, repo := range m.repositories {
		// Initialize animation progress
		m.animationProgress[repo.FullName] = 0.0
		// Start animation ticker for each repo
		cmds = append(cmds, m.animationTickCmd(repo.FullName))
	}
	return tea.Batch(cmds...)
}

// animationTickCmd creates a command that sends animation ticks
func (m *CloningModel) animationTickCmd(repoName string) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return AnimationTickMsg{Repository: repoName}
	})
}

// handleAnimationTick processes animation tick messages for smooth progress
func (m *CloningModel) handleAnimationTick(msg AnimationTickMsg) (tea.Model, tea.Cmd) {
	// Don't animate if already completed or if we don't have this repo
	if m.allCompleted || m.completed[msg.Repository] {
		return m, nil
	}

	// Get current status to determine appropriate animation speed
	status := m.statuses[msg.Repository]
	increment := 0.02 // Base increment per tick (100ms = 0.02 per second)

	// Adjust animation speed based on status
	switch {
	case strings.Contains(status, "Preparing"):
		increment = 0.01 // Slow for preparing
	case strings.Contains(status, "Initializing"):
		increment = 0.015 // Slower for initializing
	case strings.Contains(status, "Cloning"):
		increment = 0.03 // Faster for active cloning
	case strings.Contains(status, "Fetching"):
		increment = 0.025 // Medium speed for fetching
	case strings.Contains(status, "Waiting"):
		increment = 0.005 // Very slow for waiting
	}

	// Update animation progress
	currentProgress := m.animationProgress[msg.Repository]
	newProgress := currentProgress + increment

	// Enhance status progression based on progress
	m.updateStatusProgression(msg.Repository, newProgress)

	// Cap progress to not exceed 0.95 (leave room for completion)
	if newProgress > 0.95 {
		newProgress = 0.95
	}

	m.animationProgress[msg.Repository] = newProgress

	// Continue animation if not completed
	if !m.completed[msg.Repository] {
		return m, m.animationTickCmd(msg.Repository)
	}

	return m, nil
}

// updateStatusProgression updates status messages based on progress for richer UX
func (m *CloningModel) updateStatusProgression(repository string, progress float64) {
	// Don't update if we have a real status from the clone manager
	if m.completed[repository] || m.errors[repository] != nil {
		return
	}

	// Update status based on progress to create a rich animation
	switch {
	case progress < 0.1:
		m.statuses[repository] = "󰔟 Preparing..."
	case progress < 0.2:
		m.statuses[repository] = "󰦖 Initializing repository..."
	case progress < 0.4:
		m.statuses[repository] = "󰓂 Connecting to GitHub..."
	case progress < 0.6:
		m.statuses[repository] = "󰇚 Cloning objects..."
	case progress < 0.8:
		m.statuses[repository] = "󰇘 Fetching files..."
	case progress < 0.95:
		m.statuses[repository] = "󰧑 Finalizing..."
	}
}

// handleDurationCheck ensures minimum duration is met before allowing completion
func (m *CloningModel) handleDurationCheck() (tea.Model, tea.Cmd) {
	elapsed := time.Since(m.startTime)

	// If minimum duration hasn't been met, wait a bit more
	if elapsed < m.minDuration {
		remaining := m.minDuration - elapsed
		return m, tea.Tick(remaining, func(t time.Time) tea.Msg {
			return DurationCheckMsg{}
		})
	}

	// Minimum duration met, allow completion
	m.allCompleted = true
	return m, nil
}

// Message types for cloning process
type CloneStartMsg struct{}

type CloneProgressMsg struct {
	Repository string
	Status     string
	Progress   float64
	Error      error
	Completed  bool
}

type CloneCompleteMsg struct{}

// Animation tick message for smooth progress updates
type AnimationTickMsg struct {
	Repository string
}

// Duration check message
type DurationCheckMsg struct{}
