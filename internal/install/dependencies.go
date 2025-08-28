package install

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/lvcasx1/quikgit/internal/detect"
)

type InstallProgress struct {
	Repository  string
	ProjectType string
	Command     string
	Status      string
	Output      string
	Error       error
	Completed   bool
	Duration    time.Duration
}

type InstallResult struct {
	Repository  string
	ProjectType string
	Success     bool
	Commands    []CommandResult
	Duration    time.Duration
	Error       error
}

type CommandResult struct {
	Command  string
	Success  bool
	Output   string
	Error    error
	Duration time.Duration
	ExitCode int
}

type Manager struct {
	progress    chan InstallProgress
	concurrent  int
	timeout     time.Duration
	skipOnError bool
}

func NewManager(concurrent int, timeout time.Duration) *Manager {
	if concurrent <= 0 {
		concurrent = 3
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}

	return &Manager{
		progress:   make(chan InstallProgress, 100),
		concurrent: concurrent,
		timeout:    timeout,
	}
}

func (m *Manager) SetSkipOnError(skip bool) {
	m.skipOnError = skip
}

func (m *Manager) GetProgressChannel() <-chan InstallProgress {
	return m.progress
}

func (m *Manager) InstallDependencies(ctx context.Context, repositories []string) ([]InstallResult, error) {
	if len(repositories) == 0 {
		close(m.progress)
		return nil, nil
	}

	results := make([]InstallResult, len(repositories))
	semaphore := make(chan struct{}, m.concurrent)
	var wg sync.WaitGroup

	for i, repo := range repositories {
		wg.Add(1)
		go func(index int, repoPath string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			results[index] = m.installForRepository(ctx, repoPath)
		}(i, repo)
	}

	wg.Wait()
	close(m.progress)

	return results, nil
}

func (m *Manager) InstallForRepository(ctx context.Context, repositoryPath string) InstallResult {
	return m.installForRepository(ctx, repositoryPath)
}

func (m *Manager) installForRepository(ctx context.Context, repositoryPath string) InstallResult {
	start := time.Now()

	repoName := filepath.Base(repositoryPath)
	result := InstallResult{
		Repository: repoName,
		Commands:   []CommandResult{},
	}

	m.sendProgress(InstallProgress{
		Repository: repoName,
		Status:     "Detecting project type",
	})

	// Detect project type
	detector := detect.NewDetector(repositoryPath)
	projectType, err := detector.DetectPrimaryProject()
	if err != nil {
		result.Error = fmt.Errorf("failed to detect project type: %w", err)
		result.Duration = time.Since(start)

		m.sendProgress(InstallProgress{
			Repository: repoName,
			Status:     "Failed",
			Error:      result.Error,
			Completed:  true,
		})

		return result
	}

	if projectType == nil {
		result.Error = fmt.Errorf("no supported project type detected")
		result.Duration = time.Since(start)

		m.sendProgress(InstallProgress{
			Repository: repoName,
			Status:     "Skipped - unsupported project type",
			Completed:  true,
		})

		return result
	}

	result.ProjectType = projectType.Name

	m.sendProgress(InstallProgress{
		Repository:  repoName,
		ProjectType: projectType.Name,
		Status:      fmt.Sprintf("Installing dependencies for %s project", projectType.Name),
	})

	// Execute commands
	allSuccessful := true
	for _, cmd := range projectType.Commands {
		if ctx.Err() != nil {
			break
		}

		cmdResult := m.executeCommand(ctx, repositoryPath, repoName, projectType.Name, cmd)
		result.Commands = append(result.Commands, cmdResult)

		if !cmdResult.Success {
			allSuccessful = false
			if cmd.Required && !m.skipOnError {
				break
			}
		}
	}

	result.Success = allSuccessful && len(result.Commands) > 0
	result.Duration = time.Since(start)

	status := "Completed"
	if !result.Success {
		status = "Failed"
	}

	m.sendProgress(InstallProgress{
		Repository:  repoName,
		ProjectType: projectType.Name,
		Status:      status,
		Completed:   true,
		Duration:    result.Duration,
		Error:       result.Error,
	})

	return result
}

func (m *Manager) executeCommand(ctx context.Context, repoPath, repoName, projectType string, cmd detect.Command) CommandResult {
	start := time.Now()

	cmdStr := fmt.Sprintf("%s %s", cmd.Command, strings.Join(cmd.Args, " "))

	m.sendProgress(InstallProgress{
		Repository:  repoName,
		ProjectType: projectType,
		Command:     cmdStr,
		Status:      fmt.Sprintf("Running: %s", cmdStr),
	})

	result := CommandResult{
		Command: cmdStr,
	}

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	execCmd := exec.CommandContext(cmdCtx, cmd.Command, cmd.Args...)
	execCmd.Dir = repoPath

	// Set up environment
	execCmd.Env = os.Environ()

	// Create pipes for stdout and stderr
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		result.Error = fmt.Errorf("failed to create stdout pipe: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	stderr, err := execCmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Errorf("failed to create stderr pipe: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Start the command
	if err := execCmd.Start(); err != nil {
		result.Error = fmt.Errorf("failed to start command: %w", err)
		result.Duration = time.Since(start)
		return result
	}

	// Read output in separate goroutines
	var outputBuilder strings.Builder
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString(line + "\n")

			// Send more detailed progress updates based on output
			status := "Running"
			if strings.Contains(strings.ToLower(line), "installing") {
				status = "Installing packages"
			} else if strings.Contains(strings.ToLower(line), "downloading") {
				status = "Downloading packages"
			} else if strings.Contains(strings.ToLower(line), "resolving") {
				status = "Resolving dependencies"
			} else if strings.Contains(strings.ToLower(line), "building") {
				status = "Building packages"
			} else if strings.Contains(strings.ToLower(line), "fetching") {
				status = "Fetching packages"
			}

			m.sendProgress(InstallProgress{
				Repository:  repoName,
				ProjectType: projectType,
				Command:     cmdStr,
				Status:      status,
				Output:      line,
			})
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			outputBuilder.WriteString("STDERR: " + line + "\n")
		}
	}()

	// Send periodic updates while command is running
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.sendProgress(InstallProgress{
					Repository:  repoName,
					ProjectType: projectType,
					Command:     cmdStr,
					Status:      "Running...",
				})
			case <-done:
				return
			}
		}
	}()

	// Wait for the command to finish
	err = execCmd.Wait()
	close(done) // Stop the periodic updates
	wg.Wait()

	result.Output = outputBuilder.String()
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
	} else {
		result.Success = true
	}

	return result
}

func (m *Manager) sendProgress(progress InstallProgress) {
	select {
	case m.progress <- progress:
	case <-time.After(100 * time.Millisecond):
		// Channel is full or blocked, try to make space
		select {
		case m.progress <- progress:
		default:
			// Still can't send, skip this update
		}
	}
}

// CheckCommandAvailability checks if required commands are available on the system
func CheckCommandAvailability(projectType *detect.ProjectType) map[string]bool {
	availability := make(map[string]bool)

	for _, cmd := range projectType.Commands {
		_, err := exec.LookPath(cmd.Command)
		availability[cmd.Command] = err == nil
	}

	return availability
}

// GetInstallationSuggestions provides suggestions for installing missing commands
func GetInstallationSuggestions(command string) []string {
	suggestions := map[string][]string{
		"go": {
			"Visit https://golang.org/dl/ to download and install Go",
			"On macOS: brew install go",
			"On Ubuntu/Debian: sudo apt install golang-go",
		},
		"node": {
			"Visit https://nodejs.org/ to download Node.js",
			"On macOS: brew install node",
			"On Ubuntu/Debian: curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs",
		},
		"npm": {
			"npm comes with Node.js - install Node.js first",
		},
		"yarn": {
			"npm install -g yarn",
			"On macOS: brew install yarn",
		},
		"pip": {
			"pip comes with Python - install Python first",
			"On macOS: brew install python",
			"On Ubuntu/Debian: sudo apt install python3-pip",
		},
		"pipenv": {
			"pip install pipenv",
		},
		"poetry": {
			"curl -sSL https://install.python-poetry.org | python3 -",
			"pip install poetry",
		},
		"bundle": {
			"gem install bundler",
			"Ruby and RubyGems must be installed first",
		},
		"cargo": {
			"Visit https://rustup.rs/ to install Rust and Cargo",
			"curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
		},
		"composer": {
			"Visit https://getcomposer.org/download/",
			"php -r \"copy('https://getcomposer.org/installer', 'composer-setup.php');\" && php composer-setup.php",
		},
		"mvn": {
			"Visit https://maven.apache.org/install.html",
			"On macOS: brew install maven",
			"On Ubuntu/Debian: sudo apt install maven",
		},
		"gradle": {
			"Visit https://gradle.org/install/",
			"On macOS: brew install gradle",
		},
		"cmake": {
			"Visit https://cmake.org/download/",
			"On macOS: brew install cmake",
			"On Ubuntu/Debian: sudo apt install cmake",
		},
		"dotnet": {
			"Visit https://dotnet.microsoft.com/download",
			"On macOS: brew install --cask dotnet",
		},
		"swift": {
			"Swift comes with Xcode on macOS",
			"On Linux: visit https://swift.org/download/",
		},
		"flutter": {
			"Visit https://flutter.dev/docs/get-started/install",
			"On macOS: brew install --cask flutter",
		},
	}

	if suggestions, exists := suggestions[command]; exists {
		return suggestions
	}

	return []string{fmt.Sprintf("Please install %s manually", command)}
}
