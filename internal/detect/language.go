package detect

import (
	"encoding/json"
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

// Framework detection based on package.json dependencies
var JSFrameworkDetectors = []ProjectType{
	{
		Name:     "Next.js (Yarn Package)",
		Language: "JavaScript",
		Files:    []string{"package.json", "yarn.lock"},
		Commands: []Command{
			{
				Name:        "yarn-install-next",
				Command:     "yarn",
				Args:        []string{"install"},
				Description: "Install Next.js dependencies via Yarn",
				Required:    true,
			},
		},
		Description: "Next.js React framework with Yarn (detected via package.json)",
	},
	{
		Name:     "Next.js (Package)",
		Language: "JavaScript",
		Files:    []string{"package.json"},
		Commands: []Command{
			{
				Name:        "npm-install-next",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Next.js dependencies",
				Required:    true,
			},
		},
		Description: "Next.js React framework (detected via package.json)",
	},
	{
		Name:     "Vue.js (Yarn Package)",
		Language: "JavaScript",
		Files:    []string{"package.json", "yarn.lock"},
		Commands: []Command{
			{
				Name:        "yarn-install-vue",
				Command:     "yarn",
				Args:        []string{"install"},
				Description: "Install Vue.js dependencies via Yarn",
				Required:    true,
			},
		},
		Description: "Vue.js framework with Yarn (detected via package.json)",
	},
	{
		Name:     "Vue.js (Package)",
		Language: "JavaScript",
		Files:    []string{"package.json"},
		Commands: []Command{
			{
				Name:        "npm-install-vue",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Vue.js dependencies",
				Required:    true,
			},
		},
		Description: "Vue.js framework (detected via package.json)",
	},
	{
		Name:     "Angular (Yarn Package)",
		Language: "TypeScript",
		Files:    []string{"package.json", "yarn.lock"},
		Commands: []Command{
			{
				Name:        "yarn-install-angular",
				Command:     "yarn",
				Args:        []string{"install"},
				Description: "Install Angular dependencies via Yarn",
				Required:    true,
			},
		},
		Description: "Angular framework with Yarn (detected via package.json)",
	},
	{
		Name:     "Angular (Package)",
		Language: "TypeScript",
		Files:    []string{"package.json"},
		Commands: []Command{
			{
				Name:        "npm-install-angular",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Angular dependencies",
				Required:    true,
			},
		},
		Description: "Angular framework (detected via package.json)",
	},
	{
		Name:     "React (Yarn Package)",
		Language: "JavaScript",
		Files:    []string{"package.json", "yarn.lock"},
		Commands: []Command{
			{
				Name:        "yarn-install-react",
				Command:     "yarn",
				Args:        []string{"install"},
				Description: "Install React dependencies via Yarn",
				Required:    true,
			},
		},
		Description: "React framework with Yarn (detected via package.json)",
	},
	{
		Name:     "React (Package)",
		Language: "JavaScript",
		Files:    []string{"package.json"},
		Commands: []Command{
			{
				Name:        "npm-install-react",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install React dependencies",
				Required:    true,
			},
		},
		Description: "React framework (detected via package.json)",
	},
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
	{
		Name:     "Next.js",
		Language: "JavaScript",
		Files:    []string{"next.config.js", "next.config.mjs", "next.config.ts", "pages/", "app/"},
		Commands: []Command{
			{
				Name:        "npm-install-next",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Next.js dependencies",
				Required:    true,
			},
		},
		Description: "Next.js React framework",
	},
	{
		Name:     "Next.js (Yarn)",
		Language: "JavaScript",
		Files:    []string{"yarn.lock", "next.config.js", "next.config.mjs", "next.config.ts", "pages/", "app/"},
		Commands: []Command{
			{
				Name:        "yarn-install-next",
				Command:     "yarn",
				Args:        []string{"install"},
				Description: "Install Next.js dependencies via Yarn",
				Required:    true,
			},
		},
		Description: "Next.js React framework with Yarn",
	},
	{
		Name:     "Vue.js",
		Language: "JavaScript",
		Files:    []string{"vue.config.js", "vite.config.js", "src/main.js", "src/App.vue"},
		Commands: []Command{
			{
				Name:        "npm-install-vue",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Vue.js dependencies",
				Required:    true,
			},
		},
		Description: "Vue.js framework",
	},
	{
		Name:     "Vue.js (Yarn)",
		Language: "JavaScript",
		Files:    []string{"yarn.lock", "vue.config.js", "vite.config.js", "src/main.js", "src/App.vue"},
		Commands: []Command{
			{
				Name:        "yarn-install-vue",
				Command:     "yarn",
				Args:        []string{"install"},
				Description: "Install Vue.js dependencies via Yarn",
				Required:    true,
			},
		},
		Description: "Vue.js framework with Yarn",
	},
	{
		Name:     "Angular",
		Language: "TypeScript",
		Files:    []string{"angular.json", "src/app/app.module.ts"},
		Commands: []Command{
			{
				Name:        "npm-install-angular",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Angular dependencies",
				Required:    true,
			},
		},
		Description: "Angular framework",
	},
	{
		Name:     "Angular (Yarn)",
		Language: "TypeScript",
		Files:    []string{"yarn.lock", "angular.json", "src/app/app.module.ts"},
		Commands: []Command{
			{
				Name:        "yarn-install-angular",
				Command:     "yarn",
				Args:        []string{"install"},
				Description: "Install Angular dependencies via Yarn",
				Required:    true,
			},
		},
		Description: "Angular framework with Yarn",
	},
	{
		Name:     "Svelte",
		Language: "JavaScript",
		Files:    []string{"svelte.config.js", "src/App.svelte"},
		Commands: []Command{
			{
				Name:        "npm-install-svelte",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Svelte dependencies",
				Required:    true,
			},
		},
		Description: "Svelte framework",
	},
	{
		Name:     "SvelteKit",
		Language: "JavaScript",
		Files:    []string{"svelte.config.js", "src/app.html"},
		Commands: []Command{
			{
				Name:        "npm-install-sveltekit",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install SvelteKit dependencies",
				Required:    true,
			},
		},
		Description: "SvelteKit framework",
	},
	{
		Name:     "Nuxt.js",
		Language: "JavaScript",
		Files:    []string{"nuxt.config.js", "nuxt.config.ts"},
		Commands: []Command{
			{
				Name:        "npm-install-nuxt",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Nuxt.js dependencies",
				Required:    true,
			},
		},
		Description: "Nuxt.js Vue framework",
	},
	{
		Name:     "Gatsby",
		Language: "JavaScript",
		Files:    []string{"gatsby-config.js", "gatsby-config.ts"},
		Commands: []Command{
			{
				Name:        "npm-install-gatsby",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Gatsby dependencies",
				Required:    true,
			},
		},
		Description: "Gatsby React framework",
	},
	{
		Name:     "Vite",
		Language: "JavaScript",
		Files:    []string{"vite.config.js", "vite.config.ts"},
		Commands: []Command{
			{
				Name:        "npm-install-vite",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Vite dependencies",
				Required:    true,
			},
		},
		Description: "Vite build tool",
	},
	{
		Name:     "Astro",
		Language: "JavaScript",
		Files:    []string{"astro.config.mjs", "astro.config.js"},
		Commands: []Command{
			{
				Name:        "npm-install-astro",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Astro dependencies",
				Required:    true,
			},
		},
		Description: "Astro static site generator",
	},
	{
		Name:     "Remix",
		Language: "JavaScript",
		Files:    []string{"remix.config.js", "app/entry.client.tsx"},
		Commands: []Command{
			{
				Name:        "npm-install-remix",
				Command:     "npm",
				Args:        []string{"install"},
				Description: "Install Remix dependencies",
				Required:    true,
			},
		},
		Description: "Remix React framework",
	},
}

type Detector struct {
	projectPath string
}

type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

func NewDetector(projectPath string) *Detector {
	return &Detector{projectPath: projectPath}
}

func (d *Detector) parsePackageJSON() (*PackageJSON, error) {
	packagePath := filepath.Join(d.projectPath, "package.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	err = json.Unmarshal(data, &pkg)
	if err != nil {
		return nil, err
	}

	return &pkg, nil
}

func (d *Detector) hasPackageDependency(depNames ...string) bool {
	pkg, err := d.parsePackageJSON()
	if err != nil {
		return false
	}

	for _, depName := range depNames {
		if _, exists := pkg.Dependencies[depName]; exists {
			return true
		}
		if _, exists := pkg.DevDependencies[depName]; exists {
			return true
		}
	}
	return false
}

func (d *Detector) DetectProjects() ([]*ProjectType, error) {
	var detected []*ProjectType

	// First, check for JavaScript frameworks via package.json
	for _, project := range JSFrameworkDetectors {
		if d.matchesJSFramework(project) {
			projectCopy := project
			detected = append(detected, &projectCopy)
		}
	}

	// Then check regular file-based detection
	for _, project := range SupportedProjects {
		if d.matchesProject(project) {
			projectCopy := project
			detected = append(detected, &projectCopy)
		}
	}

	return detected, nil
}

func (d *Detector) matchesJSFramework(project ProjectType) bool {
	// Check if package.json exists first
	if !d.hasMatchingFiles("package.json") {
		return false
	}

	// Determine if this is a yarn or npm project
	hasYarnLock := d.hasMatchingFiles("yarn.lock")

	// Check for specific framework dependencies
	switch project.Name {
	case "Next.js (Yarn Package)":
		return hasYarnLock && d.hasPackageDependency("next")
	case "Next.js (Package)":
		return !hasYarnLock && d.hasPackageDependency("next")
	case "Vue.js (Yarn Package)":
		return hasYarnLock && d.hasPackageDependency("vue", "@vue/cli", "nuxt")
	case "Vue.js (Package)":
		return !hasYarnLock && d.hasPackageDependency("vue", "@vue/cli", "nuxt")
	case "Angular (Yarn Package)":
		return hasYarnLock && d.hasPackageDependency("@angular/core", "@angular/cli")
	case "Angular (Package)":
		return !hasYarnLock && d.hasPackageDependency("@angular/core", "@angular/cli")
	case "React (Yarn Package)":
		return hasYarnLock && d.hasPackageDependency("react") && !d.hasPackageDependency("next", "gatsby", "@remix-run/react")
	case "React (Package)":
		return !hasYarnLock && d.hasPackageDependency("react") && !d.hasPackageDependency("next", "gatsby", "@remix-run/react")
	}

	return false
}

func (d *Detector) DetectPrimaryProject() (*ProjectType, error) {
	projects, err := d.DetectProjects()
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, nil
	}

	// Priority order for conflicting project types (package.json detection gets highest priority)
	priorities := map[string]int{
		// Package.json based detection (highest priority)
		"Next.js (Yarn Package)":    35,
		"Next.js (Package)":         34,
		"Angular (Yarn Package)":    33,
		"Angular (Package)":         32,
		"Vue.js (Yarn Package)":     31,
		"Vue.js (Package)":          30,
		"React (Yarn Package)":      29,
		"React (Package)":           28,

		// File-based framework detection
		"Next.js (Yarn)":     25,
		"Next.js":            24,
		"Nuxt.js":            23,
		"Gatsby":             22,
		"Remix":              21,
		"Angular (Yarn)":     20,
		"Angular":            19,
		"Vue.js (Yarn)":      18,
		"Vue.js":             17,
		"SvelteKit":          16,
		"Svelte":             15,
		"Astro":              14,
		"Vite":               13,

		// Generic Node.js (lower priority than frameworks)
		"Node.js (yarn)":     12,
		"Node.js (npm)":      11,

		// Other languages
		"Python (Poetry)":    10,
		"Python (Pipenv)":    9,
		"Python (pip)":       8,
		"Go":                 7,
		"Rust":               6,
		"Java (Gradle)":      5,
		"Java (Maven)":       4,
		"Ruby (Bundler)":     3,
		"Dart (Flutter)":     2,
		"Swift":              1,
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
	stat, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	
	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		return stat.IsDir()
	}
	
	return true
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
