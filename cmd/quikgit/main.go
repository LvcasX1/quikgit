package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/lvcasx1/quikgit/internal/ui/bubbletea"
	"github.com/lvcasx1/quikgit/pkg/config"
)

var (
	version = "dev"      // Injected by goreleaser
	commit  = "none"     // Injected by goreleaser
	date    = "unknown"  // Injected by goreleaser
)

const (
	appName = "QuikGit"
)

var (
	showVersion = flag.Bool("version", false, "Show version information")
	showHelp    = flag.Bool("help", false, "Show help information")
	configPath  = flag.String("config", "", "Path to configuration file")
	debug       = flag.Bool("debug", false, "Enable debug mode")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s v%s\n", appName, version)
		fmt.Println("A GitHub repository manager TUI")
		os.Exit(0)
	}

	if *showHelp {
		showUsage()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config path if provided
	if *configPath != "" {
		cfg.ConfigPath = *configPath
	}

	// Set up debug logging if enabled
	if *debug {
		f, err := os.OpenFile("quikgit.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	// Initialize and run the Bubble Tea application
	app := bubbletea.NewApplication(cfg)
	if err := app.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}

func showUsage() {
	fmt.Printf(`%s v%s - GitHub Repository Manager TUI

USAGE:
    %s [OPTIONS]

OPTIONS:
    --version          Show version information
    --help             Show this help message
    --config PATH      Path to configuration file
    --debug            Enable debug logging

DESCRIPTION:
    QuikGit is a terminal user interface for managing GitHub repositories.
    It provides an intuitive way to search, clone, and set up repositories
    with automatic dependency installation.

FEATURES:
    • GitHub OAuth authentication with QR code
    • Repository search and filtering
    • Multi-repository cloning
    • Automatic dependency detection and installation
    • Support for multiple programming languages
    • Mouse and keyboard navigation

KEYBOARD SHORTCUTS:
    q, Ctrl+C         Quit application
    Esc              Go back to previous screen
    F1, ?            Show help
    Tab              Navigate between fields
    Enter            Confirm/Select
    Space            Toggle selection (in lists)

CONFIGURATION:
    Configuration file is located at ~/.quikgit/config.yaml
    GitHub token is stored securely at ~/.quikgit/token

SUPPORTED LANGUAGES:
    Go, Node.js, Python, Ruby, Rust, Java, C++, C#, Swift, PHP, Dart

EXAMPLES:
    # Start QuikGit with default configuration
    %s

    # Use custom configuration file
    %s --config /path/to/config.yaml

    # Enable debug logging
    %s --debug

For more information, visit: https://github.com/lvcasx1/quikgit
`, appName, version, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}
