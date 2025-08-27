package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/lvcasx1/quikgit/internal/install"
)

// InstallationProgress represents the progress of a dependency installation
type InstallationProgress struct {
	Repository string
	Status     string
	Progress   float64
	Error      error
	Completed  bool
}

// InstallationManager handles the dependency installation process
type InstallationManager struct {
	app          *Application
	repositories []string // Paths to cloned repositories
	progressView *tview.TextView
	progressBars map[string]*ProgressBar
	installMgr   *install.Manager
	ctx          context.Context
	cancel       context.CancelFunc
	mutex        sync.Mutex

	// Progress tracking
	completed    map[string]bool
	errors       map[string]error
	statuses     map[string]string
	allCompleted bool
	successCount int
	errorCount   int
}

// startInstallation starts the dependency installation process
func (a *Application) startInstallation(clonedPaths []string) {
	if len(clonedPaths) == 0 {
		a.message = "No repositories to install dependencies for"
		a.showMainMenu()
		return
	}

	// Create installation manager
	ctx, cancel := context.WithCancel(a.ctx)

	im := &InstallationManager{
		app:          a,
		repositories: clonedPaths,
		ctx:          ctx,
		cancel:       cancel,
		progressBars: make(map[string]*ProgressBar),
		completed:    make(map[string]bool),
		errors:       make(map[string]error),
		statuses:     make(map[string]string),
	}

	// Initialize progress bars
	for _, repoPath := range clonedPaths {
		repoName := filepath.Base(repoPath)
		im.progressBars[repoName] = NewProgressBar(40)
		im.statuses[repoName] = "Waiting"
	}

	// Create progress view
	im.createProgressView()

	// Start installation in background
	go im.performInstallation()

	a.currentState = StateInstalling
}

// createProgressView creates the installation progress display
func (im *InstallationManager) createProgressView() {
	im.progressView = tview.NewTextView()
	im.progressView.SetBorder(true)
	im.progressView.SetTitle(fmt.Sprintf(" %s Installing Dependencies (%d) ", IconMap["🚀"], len(im.repositories)))
	im.progressView.SetDynamicColors(true)
	im.progressView.SetScrollable(true)
	im.progressView.SetBackgroundColor(tcell.ColorDefault)

	// Initial content
	im.updateProgressView()

	// Set up input capture
	im.progressView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC && !im.allCompleted {
			im.cancel()
			im.app.app.QueueUpdateDraw(func() {
				im.app.message = "Installation cancelled by user"
				im.app.showMainMenu()
			})
			return nil
		}

		if event.Key() == tcell.KeyEnter && im.allCompleted {
			im.app.app.QueueUpdateDraw(func() {
				im.app.message = fmt.Sprintf("Installation completed! %d successful, %d failed", im.successCount, im.errorCount)
				im.app.showMainMenu()
			})
			return nil
		}

		return event
	})

	// Create centered layout
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(im.progressView, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 4, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	im.app.pages.AddAndSwitchToPage("installing", flex, true)
	im.app.app.SetFocus(im.progressView)
}

// updateProgressView updates the installation progress display
func (im *InstallationManager) updateProgressView() {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	var content strings.Builder

	for _, repoPath := range im.repositories {
		repoName := filepath.Base(repoPath)
		content.WriteString(fmt.Sprintf("%s [-]%s[-]\n", IconMap["📁"], repoName))

		if bar, exists := im.progressBars[repoName]; exists {
			content.WriteString(bar.Render())
		}

		// Show status or error
		if err, hasError := im.errors[repoName]; hasError {
			content.WriteString(fmt.Sprintf("\n[red]%s Error: %s[white]\n\n", IconMap["❌"], err.Error()))
		} else if im.completed[repoName] {
			content.WriteString(fmt.Sprintf("\n[green]%s Completed[white]\n\n", IconMap["✅"]))
		} else {
			content.WriteString(fmt.Sprintf("\n[yellow]%s %s[white]\n\n", IconMap["🔄"], im.statuses[repoName]))
		}
	}

	// Add summary
	if im.allCompleted {
		content.WriteString("─────────────────────────────────────\n")
		content.WriteString(fmt.Sprintf("[-]Summary:[-] %d successful, %d failed\n", im.successCount, im.errorCount))
		content.WriteString(fmt.Sprintf("\n[green]%s Press Enter to continue[white]", IconMap["🚀"]))
	} else {
		content.WriteString(fmt.Sprintf("[-]Progress:[-] %d/%d repositories processed\n", im.successCount+im.errorCount, len(im.repositories)))
		content.WriteString("[-]Press Ctrl+C to cancel installation[-]")
	}

	im.progressView.SetText(content.String())
}

// performInstallation executes the actual dependency installation
func (im *InstallationManager) performInstallation() {
	// Create installation manager
	im.installMgr = install.NewManager(3, 10*time.Minute)
	im.installMgr.SetSkipOnError(true)

	// Start progress monitoring
	go im.monitorProgress()

	// Start installing dependencies for all repositories
	_, err := im.installMgr.InstallDependencies(im.ctx, im.repositories)
	if err != nil {
		// Handle error
		im.app.app.QueueUpdateDraw(func() {
			im.app.error = fmt.Errorf("installation failed: %w", err)
			im.allCompleted = true
			im.updateProgressView()
		})
		return
	}

	// Mark as completed
	im.app.app.QueueUpdateDraw(func() {
		im.allCompleted = true
		im.updateProgressView()
	})
}

// monitorProgress monitors the installation progress
func (im *InstallationManager) monitorProgress() {
	progressChan := im.installMgr.GetProgressChannel()

	for progress := range progressChan {
		im.mutex.Lock()

		repoName := filepath.Base(progress.Repository)

		// Update progress bar
		if bar, exists := im.progressBars[repoName]; exists {
			progressPercent := 0.0
			if progress.Completed {
				progressPercent = 1.0
			} else {
				// Estimate progress based on status
				switch progress.Status {
				case "Detecting language":
					progressPercent = 0.1
				case "Installing dependencies":
					progressPercent = 0.5
				case "Running post-install":
					progressPercent = 0.8
				default:
					progressPercent = 0.2
				}
			}
			bar.Update(progressPercent, progress.Status)
		}

		// Update status
		im.statuses[repoName] = progress.Status

		// Check if completed
		if progress.Completed {
			im.completed[repoName] = true
			if progress.Error != nil {
				im.errors[repoName] = progress.Error
				im.errorCount++
			} else {
				im.successCount++
			}
		}

		// Check if all completed
		if im.successCount+im.errorCount >= len(im.repositories) {
			im.allCompleted = true
		}

		im.mutex.Unlock()

		// Update UI
		im.app.app.QueueUpdateDraw(func() {
			im.updateProgressView()
		})
	}
}
