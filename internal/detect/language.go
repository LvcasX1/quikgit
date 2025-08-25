package detect

import (
	"os"
	"path/filepath"
	"strings"
)

type ProjectType struct {
	Name        string
	Language    string
	Files       []string
	Commands    []Command
	Description string
}

type Command struct {
	Name        string
	Command     string
	Args        []string
	Description string
	Required    bool
}

var SupportedProjects = []ProjectType{
	{
		Name:     "Go",
		Language: "Go",
		Files:    []string{"go.mod", "go.sum", "*.go"},
		Commands: []Command{
			{
				Name:        "go-mod-tidy",
				Command:     "go",
				Args:        []string{"mod", "tidy"},
				Description: "Download and organize dependencies",
				Required:    true,
			},
			{
				Name:        "go-mod-download",
				Command:     "go",
				Args:        []string{"mod", "download"},
				Description: "Download dependencies to cache",
				Required:    false,
			},
		},
		Description: "Go module project",
	},
	{
		Name:     "Node.js (npm)",
		Language: "JavaScript",
		Files:    []string{"package.json", "package-lock.json"},
		Commands: []Command{
			{
				Name:        "npm-install",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Node.js dependencies via npm",
				Required:    true,
			},
		},
		Description: "Node.js project with npm",
	},
	{
		Name:     "Node.js (yarn)",
		Language: "JavaScript",
		Files:    []string{"package.json", "yarn.lock"},
		Commands: []Command{
			{
				Name:        "yarn-install",
				Command:     "yarn",
				Args:        []string{"install"},
				Description: "Install Node.js dependencies via Yarn",
				Required:    true,
			},
		},
		Description: "Node.js project with Yarn",
	},
	{
		Name:     "Python (pip)",
		Language: "Python",
		Files:    []string{"requirements.txt"},
		Commands: []Command{
			{
				Name:        "pip-install",
				Command:     "pip",
				Args:        []string{"install", "-r", "requirements.txt"},
				Description: "Install Python dependencies via pip",
				Required:    true,
			},
		},
		Description: "Python project with requirements.txt",
	},
	{
		Name:     "Python (Pipenv)",
		Language: "Python",
		Files:    []string{"Pipfile"},
		Commands: []Command{
			{
				Name:        "pipenv-install",
				Command:     "pipenv",
				Args:        []string{"install"},
				Description: "Install Python dependencies via Pipenv",
				Required:    true,
			},
		},
		Description: "Python project with Pipenv",
	},
	{
		Name:     "Python (Poetry)",
		Language: "Python",
		Files:    []string{"pyproject.toml", "poetry.lock"},
		Commands: []Command{
			{
				Name:        "poetry-install",
				Command:     "poetry",
				Args:        []string{"install"},
				Description: "Install Python dependencies via Poetry",
				Required:    true,
			},
		},
		Description: "Python project with Poetry",
	},
	{
		Name:     "Ruby (Bundler)",
		Language: "Ruby",
		Files:    []string{"Gemfile"},
		Commands: []Command{
			{
				Name:        "bundle-install",
				Command:     "bundle",
				Args:        []string{"install"},
				Description: "Install Ruby gems via Bundler",
				Required:    true,
			},
		},
		Description: "Ruby project with Bundler",
	},
	{
		Name:     "Rust",
		Language: "Rust",
		Files:    []string{"Cargo.toml"},
		Commands: []Command{
			{
				Name:        "cargo-build",
				Command:     "cargo",
				Args:        []string{"build"},
				Description: "Build Rust project and download dependencies",
				Required:    true,
			},
		},
		Description: "Rust project with Cargo",
	},
	{
		Name:     "PHP (Composer)",
		Language: "PHP",
		Files:    []string{"composer.json"},
		Commands: []Command{
			{
				Name:        "composer-install",
				Command:     "composer",
				Args:        []string{"install"},
				Description: "Install PHP dependencies via Composer",
				Required:    true,
			},
		},
		Description: "PHP project with Composer",
	},
	{
		Name:     "Java (Maven)",
		Language: "Java",
		Files:    []string{"pom.xml"},
		Commands: []Command{
			{
				Name:        "maven-install",
				Command:     "mvn",
				Args:        []string{"install"},
				Description: "Build Java project and install dependencies via Maven",
				Required:    true,
			},
		},
		Description: "Java project with Maven",
	},
	{
		Name:     "Java (Gradle)",
		Language: "Java",
		Files:    []string{"build.gradle", "build.gradle.kts"},
		Commands: []Command{
			{
				Name:        "gradle-build",
				Command:     "gradle",
				Args:        []string{"build"},
				Description: "Build Java project via Gradle",
				Required:    true,
			},
		},
		Description: "Java project with Gradle",
	},
	{
		Name:     "C++ (CMake)",
		Language: "C++",
		Files:    []string{"CMakeLists.txt"},
		Commands: []Command{
			{
				Name:        "cmake-build",
				Command:     "cmake",
				Args:        []string{".", "-B", "build"},
				Description: "Configure CMake build",
				Required:    true,
			},
			{
				Name:        "make-build",
				Command:     "make",
				Args:        []string{"-C", "build"},
				Description: "Build C++ project",
				Required:    false,
			},
		},
		Description: "C++ project with CMake",
	},
	{
		Name:     "C# (.NET)",
		Language: "C#",
		Files:    []string{"*.csproj", "*.sln"},
		Commands: []Command{
			{
				Name:        "dotnet-restore",
				Command:     "dotnet",
				Args:        []string{"restore"},
				Description: "Restore .NET dependencies",
				Required:    true,
			},
			{
				Name:        "dotnet-build",
				Command:     "dotnet",
				Args:        []string{"build"},
				Description: "Build .NET project",
				Required:    false,
			},
		},
		Description: ".NET project",
	},
	{
		Name:     "Swift",
		Language: "Swift",
		Files:    []string{"Package.swift"},
		Commands: []Command{
			{
				Name:        "swift-build",
				Command:     "swift",
				Args:        []string{"build"},
				Description: "Build Swift package",
				Required:    true,
			},
		},
		Description: "Swift package",
	},
	{
		Name:     "Dart (Flutter)",
		Language: "Dart",
		Files:    []string{"pubspec.yaml"},
		Commands: []Command{
			{
				Name:        "flutter-pub-get",
				Command:     "flutter",
				Args:        []string{"pub", "get"},
				Description: "Get Flutter dependencies",
				Required:    true,
			},
		},
		Description: "Flutter project",
	},
}

type Detector struct {
	projectPath string
}

func NewDetector(projectPath string) *Detector {
	return &Detector{projectPath: projectPath}
}

func (d *Detector) DetectProjects() ([]*ProjectType, error) {
	var detected []*ProjectType

	for _, project := range SupportedProjects {
		if d.matchesProject(project) {
			projectCopy := project
			detected = append(detected, &projectCopy)
		}
	}

	return detected, nil
}

func (d *Detector) DetectPrimaryProject() (*ProjectType, error) {
	projects, err := d.DetectProjects()
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, nil
	}

	// Priority order for conflicting project types
	priorities := map[string]int{
		"Node.js (yarn)":  10,
		"Node.js (npm)":   9,
		"Python (Poetry)": 8,
		"Python (Pipenv)": 7,
		"Python (pip)":    6,
		"Go":              5,
		"Rust":            4,
		"Java (Gradle)":   3,
		"Java (Maven)":    2,
		"Ruby (Bundler)":  1,
	}

	var best *ProjectType
	bestPriority := -1

	for _, project := range projects {
		if priority, exists := priorities[project.Name]; exists && priority > bestPriority {
			best = project
			bestPriority = priority
		}
	}

	if best == nil {
		return projects[0], nil
	}

	return best, nil
}

func (d *Detector) matchesProject(project ProjectType) bool {
	for _, pattern := range project.Files {
		if d.hasMatchingFiles(pattern) {
			return true
		}
	}
	return false
}

func (d *Detector) hasMatchingFiles(pattern string) bool {
	if strings.Contains(pattern, "*") {
		matches, err := filepath.Glob(filepath.Join(d.projectPath, pattern))
		if err != nil {
			return false
		}
		return len(matches) > 0
	}

	filePath := filepath.Join(d.projectPath, pattern)
	_, err := os.Stat(filePath)
	return err == nil
}

func (d *Detector) GetProjectInfo(projectType *ProjectType) map[string]interface{} {
	info := map[string]interface{}{
		"name":        projectType.Name,
		"language":    projectType.Language,
		"description": projectType.Description,
		"files":       []string{},
		"commands":    projectType.Commands,
	}

	// List existing files
	var existingFiles []string
	for _, pattern := range projectType.Files {
		if strings.Contains(pattern, "*") {
			matches, err := filepath.Glob(filepath.Join(d.projectPath, pattern))
			if err == nil {
				for _, match := range matches {
					relPath, err := filepath.Rel(d.projectPath, match)
					if err == nil {
						existingFiles = append(existingFiles, relPath)
					}
				}
			}
		} else {
			filePath := filepath.Join(d.projectPath, pattern)
			if _, err := os.Stat(filePath); err == nil {
				existingFiles = append(existingFiles, pattern)
			}
		}
	}

	info["files"] = existingFiles
	return info
}

func GetSupportedLanguages() []string {
	languageSet := make(map[string]bool)
	for _, project := range SupportedProjects {
		languageSet[project.Language] = true
	}

	var languages []string
	for lang := range languageSet {
		languages = append(languages, lang)
	}

	return languages
}

func GetProjectByName(name string) *ProjectType {
	for _, project := range SupportedProjects {
		if project.Name == name {
			projectCopy := project
			return &projectCopy
		}
	}
	return nil
}
