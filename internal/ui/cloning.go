package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/lvcasx1/quikgit/internal/github"
	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

// CloningProgress represents the progress of a cloning operation
type CloningProgress struct {
	Repository string
	Status     string
	Progress   float64
	Error      error
	Completed  bool
}

// CloningManager handles the cloning process with proper progress tracking
type CloningManager struct {
	app          *Application
	repositories []*ghClient.Repository
	progressView *tview.TextView
	progressBars map[string]*ProgressBar
	cloneManager *github.CloneManager
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.Mutex
	targetDir    string // Store the target directory used for cloning

	// Progress tracking
	completed    map[string]bool
	errors       map[string]error
	statuses     map[string]string
	allCompleted bool
	successCount int
	errorCount   int
}

// ProgressBar represents a simple text-based progress bar
type ProgressBar struct {
	width    int
	progress float64
	status   string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{
		width:    width,
		progress: 0.0,
		status:   "Waiting",
	}
}

// Render renders the progress bar as text
func (pb *ProgressBar) Render() string {
	filled := int(pb.progress * float64(pb.width))
	if filled > pb.width {
		filled = pb.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("▒", pb.width-filled)
	percentage := int(pb.progress * 100)

	return fmt.Sprintf("[green]%s[white] %3d%% [yellow]%s[white]", bar, percentage, pb.status)
}

// Update updates the progress bar
func (pb *ProgressBar) Update(progress float64, status string) {
	pb.progress = progress
	pb.status = status
}

// startCloning initiates the cloning process
func (a *Application) startCloning() {
	if len(a.selectedRepos) == 0 {
		a.error = fmt.Errorf("no repositories selected for cloning")
		a.showMainMenu()
		return
	}

	// Create cloning manager
	ctx, cancel := context.WithCancel(a.ctx)

	cm := &CloningManager{
		app:          a,
		repositories: a.selectedRepos,
		ctx:          ctx,
		cancel:       cancel,
		progressBars: make(map[string]*ProgressBar),
		completed:    make(map[string]bool),
		errors:       make(map[string]error),
		statuses:     make(map[string]string),
	}

	// Initialize progress bars
	for _, repo := range a.selectedRepos {
		cm.progressBars[repo.FullName] = NewProgressBar(40)
		cm.statuses[repo.FullName] = "Waiting"
	}

	// Create progress view
	cm.createProgressView()

	// Start cloning in background
	go cm.performCloning()

	a.currentState = StateCloning
}

// createProgressView creates the progress display
func (cm *CloningManager) createProgressView() {
	cm.progressView = tview.NewTextView()
	cm.progressView.SetBorder(true)
	cm.progressView.SetTitle(fmt.Sprintf(" %s Cloning Repositories (%d) ", IconMap["📁"], len(cm.repositories)))
	cm.progressView.SetDynamicColors(true)
	cm.progressView.SetScrollable(true)

	// Initial content
	cm.updateProgressView()

	// Set up input capture
	cm.progressView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC && !cm.allCompleted {
			cm.cancel()
			cm.app.app.QueueUpdateDraw(func() {
				cm.app.message = "Cloning cancelled by user"
				cm.app.showMainMenu()
			})
			return nil
		}

		if event.Key() == tcell.KeyEnter && cm.allCompleted {
			// Proceed to installation if any repos were cloned successfully
			if cm.successCount > 0 {
				cm.app.app.QueueUpdateDraw(func() {
					cm.app.startInstallation(cm.getSuccessfullyClonedPaths())
				})
			} else {
				cm.app.app.QueueUpdateDraw(func() {
					cm.app.showMainMenu()
				})
			}
			return nil
		}

		return event
	})

	// Create centered layout
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(cm.progressView, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 4, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	cm.app.pages.AddAndSwitchToPage("cloning", flex, true)
	cm.app.currentState = StateCloning
	cm.app.app.SetFocus(cm.progressView)
}

// updateProgressView updates the progress display
func (cm *CloningManager) updateProgressView() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var content strings.Builder

	for _, repo := range cm.repositories {
		content.WriteString(fmt.Sprintf("%s [white]%s[white]\n", IconMap["📁"], repo.FullName))

		if bar, exists := cm.progressBars[repo.FullName]; exists {
			content.WriteString(bar.Render())
		}

		// Show status or error
		if err, hasError := cm.errors[repo.FullName]; hasError {
			content.WriteString(fmt.Sprintf("\n[red]%s Error: %s[white]\n\n", IconMap["❌"], err.Error()))
		} else if cm.completed[repo.FullName] {
			content.WriteString(fmt.Sprintf("\n[green]%s Completed[white]\n\n", IconMap["✅"]))
		} else {
			content.WriteString(fmt.Sprintf("\n[yellow]%s %s[white]\n\n", IconMap["🔄"], cm.statuses[repo.FullName]))
		}
	}

	// Add summary
	if cm.allCompleted {
		content.WriteString("─────────────────────────────────────\n")
		content.WriteString(fmt.Sprintf("[yellow]Summary:[white] %d successful, %d failed\n", cm.successCount, cm.errorCount))

		if cm.successCount > 0 {
			content.WriteString(fmt.Sprintf("\n[green]%s Press Enter to install dependencies[white]", IconMap["🚀"]))
		} else {
			content.WriteString("\n[red]No repositories were successfully cloned. Press Enter to return to main menu.[white]")
		}
	} else {
		content.WriteString(fmt.Sprintf("[yellow]Progress:[white] %d/%d repositories processed\n", cm.successCount+cm.errorCount, len(cm.repositories)))
		content.WriteString("[yellow]Press Ctrl+C to cancel cloning[white]")
	}

	cm.progressView.SetText(content.String())
}

// performCloning executes the actual cloning process
func (cm *CloningManager) performCloning() {
	// Determine target directory based on configuration
	var targetDir string

	if cm.app.config.Clone.UseCurrentDir {
		// Use current working directory
		wd, err := os.Getwd()
		if err != nil {
			cm.app.app.QueueUpdateDraw(func() {
				cm.app.error = fmt.Errorf("failed to get working directory: %w", err)
				cm.app.showMainMenu()
			})
			return
		}
		targetDir = wd
	} else if cm.app.config.Clone.DefaultPath != "" {
		// Use configured default path
		targetDir = cm.app.config.Clone.DefaultPath
	} else {
		// Fallback to current directory
		wd, err := os.Getwd()
		if err != nil {
			cm.app.app.QueueUpdateDraw(func() {
				cm.app.error = fmt.Errorf("failed to get working directory: %w", err)
				cm.app.showMainMenu()
			})
			return
		}
		targetDir = wd
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		cm.app.app.QueueUpdateDraw(func() {
			cm.app.error = fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
			cm.app.showMainMenu()
		})
		return
	}

	// Store the target directory for later use
	cm.targetDir = targetDir

	// Get token from auth manager
	token := ""
	if cm.app.authManager != nil {
		token = cm.app.authManager.GetToken()
	}
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	// Create clone manager with the determined target directory
	cm.cloneManager = github.NewCloneManager(token, targetDir)

	// Configure subdirectory creation based on config
	cm.cloneManager.SetCreateSubdirs(cm.app.config.Clone.CreateSubdirs)

	// Start progress monitoring
	go cm.monitorProgress()

	// Start cloning repositories (don't wait here - let monitorProgress handle completion)
	cm.cloneManager.CloneRepositories(cm.ctx, cm.repositories, 3)
}

// monitorProgress monitors the cloning progress
func (cm *CloningManager) monitorProgress() {
	progressChan := cm.cloneManager.GetProgressChannel()

	for progress := range progressChan {
		cm.mutex.Lock()

		// Update progress bar
		if bar, exists := cm.progressBars[progress.Repository]; exists {
			bar.Update(progress.Progress, progress.Status)
		}

		// Update status
		cm.statuses[progress.Repository] = progress.Status

		// Check if completed
		if progress.Completed {
			cm.completed[progress.Repository] = true
			if progress.Error != nil {
				cm.errors[progress.Repository] = progress.Error
				cm.errorCount++
			} else {
				cm.successCount++
			}
		}

		// Check if all completed
		if cm.successCount+cm.errorCount >= len(cm.repositories) {
			cm.allCompleted = true
		}

		cm.mutex.Unlock()

		// Update UI
		cm.app.app.QueueUpdateDraw(func() {
			cm.updateProgressView()
		})
	}
}

// getSuccessfullyClonedPaths returns paths of successfully cloned repositories
func (cm *CloningManager) getSuccessfullyClonedPaths() []string {
	var paths []string

	for _, repo := range cm.repositories {
		if cm.completed[repo.FullName] && cm.errors[repo.FullName] == nil {
			var repoPath string
			if cm.app.config.Clone.CreateSubdirs {
				// Path is targetDir/owner/repo
				repoPath = filepath.Join(cm.targetDir, repo.Owner, repo.Name)
			} else {
				// Path is targetDir/repo
				repoPath = filepath.Join(cm.targetDir, repo.Name)
			}
			paths = append(paths, repoPath)
		}
	}

	return paths
}

// performQuickCloneEnhanced handles quick clone functionality
func (a *Application) performQuickCloneEnhanced() {
	repoInput := strings.TrimSpace(a.quickCloneForm.GetFormItemByLabel("Repository").(*tview.InputField).GetText())
	if repoInput == "" {
		a.error = fmt.Errorf("repository cannot be empty")
		a.showQuickClone()
		return
	}

	// Parse input to extract owner/repo
	var owner, repo string

	if strings.HasPrefix(repoInput, "https://github.com/") {
		// Full GitHub URL format
		parts := strings.TrimPrefix(repoInput, "https://github.com/")
		parts = strings.TrimSuffix(parts, ".git")
		repoParts := strings.Split(parts, "/")
		if len(repoParts) >= 2 {
			owner = repoParts[0]
			repo = repoParts[1]
		}
	} else if strings.Contains(repoInput, "/") {
		// owner/repo format
		repoParts := strings.Split(repoInput, "/")
		if len(repoParts) >= 2 {
			owner = repoParts[0]
			repo = repoParts[1]
		}
	}

	if owner == "" || repo == "" {
		a.error = fmt.Errorf("invalid format. Use 'owner/repo' or full GitHub URL")
		a.showQuickClone()
		return
	}

	// Create repository object for cloning
	repository := &ghClient.Repository{
		Name:     repo,
		FullName: fmt.Sprintf("%s/%s", owner, repo),
		CloneURL: fmt.Sprintf("https://github.com/%s/%s.git", owner, repo),
		SSHURL:   fmt.Sprintf("git@github.com:%s/%s.git", owner, repo),
		Owner:    owner,
	}

	// Set up for cloning
	a.selectedRepos = []*ghClient.Repository{repository}
	a.startCloning()
}

// Installation is now handled by installation.go
