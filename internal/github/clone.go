package github

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type CloneProgress struct {
	Repository string
	Status     string
	Progress   float64
	Error      error
	Completed  bool
}

type CloneManager struct {
	token         string
	sshKey        string
	targetDir     string
	progress      chan CloneProgress
	wg            sync.WaitGroup
	createSubdirs bool
}

func NewCloneManager(token, targetDir string) *CloneManager {
	if targetDir == "" {
		targetDir, _ = os.Getwd()
	}

	return &CloneManager{
		token:         token,
		targetDir:     targetDir,
		progress:      make(chan CloneProgress, 100),
		createSubdirs: false, // Default to false, can be set with SetCreateSubdirs
	}
}

// SetCreateSubdirs configures whether to create subdirectories organized by owner
func (cm *CloneManager) SetCreateSubdirs(createSubdirs bool) {
	cm.createSubdirs = createSubdirs
}

func (cm *CloneManager) SetSSHKey(keyPath string) {
	cm.sshKey = keyPath
}

func (cm *CloneManager) GetProgressChannel() <-chan CloneProgress {
	return cm.progress
}

func (cm *CloneManager) CloneRepository(ctx context.Context, repo *Repository) {
	cm.wg.Add(1)
	go cm.cloneWorker(ctx, repo)
}

func (cm *CloneManager) CloneRepositories(ctx context.Context, repos []*Repository, concurrent int) {
	if concurrent <= 0 {
		concurrent = 3
	}

	semaphore := make(chan struct{}, concurrent)

	for _, repo := range repos {
		cm.wg.Add(1)
		go func(r *Repository) {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
				cm.wg.Done()
			}()
			cm.cloneWorker(ctx, r)
		}(repo)
	}
}

func (cm *CloneManager) Wait() {
	cm.wg.Wait()
	close(cm.progress)
}

func (cm *CloneManager) cloneWorker(ctx context.Context, repo *Repository) {
	progress := CloneProgress{
		Repository: repo.FullName,
		Status:     "Starting",
		Progress:   0,
	}

	cm.sendProgress(progress)

	// Determine target path based on configuration
	var targetPath string
	if cm.createSubdirs {
		// Create subdirectory structure: targetDir/owner/repo
		ownerDir := filepath.Join(cm.targetDir, repo.Owner)
		if err := os.MkdirAll(ownerDir, 0755); err != nil {
			progress.Status = "Failed to create owner directory"
			progress.Error = fmt.Errorf("failed to create directory %s: %w", ownerDir, err)
			progress.Completed = true
			cm.sendProgress(progress)
			return
		}
		targetPath = filepath.Join(ownerDir, repo.Name)
	} else {
		// Just use repo name in target directory
		targetPath = filepath.Join(cm.targetDir, repo.Name)
	}

	// Check if directory already exists
	if _, err := os.Stat(targetPath); err == nil {
		progress.Status = "Directory exists"
		progress.Error = fmt.Errorf("directory %s already exists", targetPath)
		progress.Completed = true
		cm.sendProgress(progress)
		return
	}

	progress.Status = "Cloning"
	progress.Progress = 0.1
	cm.sendProgress(progress)

	cloneOptions := &git.CloneOptions{
		URL:      repo.CloneURL,
		Progress: &progressWriter{repo: repo.FullName, progress: cm.progress},
	}

	// Prefer HTTPS with token if available, otherwise try SSH
	if cm.token != "" {
		cloneOptions.Auth = &http.BasicAuth{
			Username: "token",
			Password: cm.token,
		}
	} else if cm.sshKey != "" && repo.SSHURL != "" {
		if auth, err := cm.getSSHAuth(); err == nil {
			cloneOptions.URL = repo.SSHURL
			cloneOptions.Auth = auth
		}
	}

	_, err := git.PlainCloneContext(ctx, targetPath, false, cloneOptions)

	progress.Progress = 1.0
	progress.Completed = true

	if err != nil {
		progress.Status = "Failed"
		progress.Error = fmt.Errorf("failed to clone %s: %w", repo.FullName, err)

		// Clean up partial clone
		os.RemoveAll(targetPath)
	} else {
		progress.Status = "Completed"
	}

	cm.sendProgress(progress)
}

func (cm *CloneManager) getSSHAuth() (*ssh.PublicKeys, error) {
	if cm.sshKey == "" {
		// Try default SSH key locations
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		keyPaths := []string{
			filepath.Join(homeDir, ".ssh", "id_rsa"),
			filepath.Join(homeDir, ".ssh", "id_ed25519"),
			filepath.Join(homeDir, ".ssh", "id_ecdsa"),
		}

		for _, path := range keyPaths {
			if _, err := os.Stat(path); err == nil {
				cm.sshKey = path
				break
			}
		}

		if cm.sshKey == "" {
			return nil, fmt.Errorf("no SSH key found")
		}
	}

	return ssh.NewPublicKeysFromFile("git", cm.sshKey, "")
}

func (cm *CloneManager) sendProgress(progress CloneProgress) {
	select {
	case cm.progress <- progress:
	default:
		// Channel is full, skip this update
	}
}

type progressWriter struct {
	repo     string
	progress chan<- CloneProgress
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	// Parse git progress output and send updates
	output := string(p)

	var progressValue float64
	var status string

	if strings.Contains(output, "Counting objects") {
		status = "Counting objects"
		progressValue = 0.2
	} else if strings.Contains(output, "Compressing objects") {
		status = "Compressing objects"
		progressValue = 0.4
	} else if strings.Contains(output, "Receiving objects") {
		status = "Receiving objects"
		progressValue = 0.6

		// Try to parse percentage from git output
		if idx := strings.Index(output, "%"); idx > 0 {
			start := idx - 3
			if start < 0 {
				start = 0
			}
			percentStr := strings.TrimSpace(output[start:idx])
			if percent := parsePercentage(percentStr); percent > 0 {
				progressValue = 0.4 + (float64(percent)/100.0)*0.4
			}
		}
	} else if strings.Contains(output, "Resolving deltas") {
		status = "Resolving deltas"
		progressValue = 0.8

		if idx := strings.Index(output, "%"); idx > 0 {
			start := idx - 3
			if start < 0 {
				start = 0
			}
			percentStr := strings.TrimSpace(output[start:idx])
			if percent := parsePercentage(percentStr); percent > 0 {
				progressValue = 0.8 + (float64(percent)/100.0)*0.2
			}
		}
	}

	if status != "" {
		progress := CloneProgress{
			Repository: pw.repo,
			Status:     status,
			Progress:   progressValue,
		}

		select {
		case pw.progress <- progress:
		default:
		}
	}

	return len(p), nil
}

func parsePercentage(s string) int {
	s = strings.TrimSpace(s)
	var percent int
	_, err := fmt.Sscanf(s, "%d", &percent)
	if err != nil {
		return 0 // Return 0 if parsing fails
	}
	return percent
}

// ValidateCloneTarget checks if the target directory is suitable for cloning
func ValidateCloneTarget(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if we can write to the directory
	testFile := filepath.Join(path, ".quikgit-test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("cannot write to directory: %w", err)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}
