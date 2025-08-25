package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitHub     GitHubConfig   `yaml:"github"`
	Clone      CloneConfig    `yaml:"clone"`
	Install    InstallConfig  `yaml:"install"`
	UI         UIConfig       `yaml:"ui"`
	Defaults   DefaultsConfig `yaml:"defaults"`
	ConfigPath string         `yaml:"-"`
}

type GitHubConfig struct {
	Token       string `yaml:"token,omitempty"`
	PreferSSH   bool   `yaml:"prefer_ssh"`
	SSHKeyPath  string `yaml:"ssh_key_path,omitempty"`
	DefaultUser string `yaml:"default_user,omitempty"`
	DefaultOrg  string `yaml:"default_org,omitempty"`
}

type CloneConfig struct {
	DefaultPath   string `yaml:"default_path,omitempty"`
	Concurrent    int    `yaml:"concurrent"`
	UseCurrentDir bool   `yaml:"use_current_dir"`
	CreateSubdirs bool   `yaml:"create_subdirs"`
}

type InstallConfig struct {
	Enabled        bool `yaml:"enabled"`
	Concurrent     int  `yaml:"concurrent"`
	TimeoutMinutes int  `yaml:"timeout_minutes"`
	SkipOnError    bool `yaml:"skip_on_error"`
	AutoInstall    bool `yaml:"auto_install"`
}

type UIConfig struct {
	Theme           string `yaml:"theme"`
	ShowIcons       bool   `yaml:"show_icons"`
	AnimationsSpeed string `yaml:"animations_speed"`
	MouseSupport    bool   `yaml:"mouse_support"`
	ShowLineNumbers bool   `yaml:"show_line_numbers"`
}

type DefaultsConfig struct {
	SearchSort     string `yaml:"search_sort"`
	SearchOrder    string `yaml:"search_order"`
	ResultsPerPage int    `yaml:"results_per_page"`
	PreferredAuth  string `yaml:"preferred_auth"`
}

var DefaultConfig = Config{
	GitHub: GitHubConfig{
		PreferSSH: false,
	},
	Clone: CloneConfig{
		Concurrent:    3,
		UseCurrentDir: true,
		CreateSubdirs: false,
	},
	Install: InstallConfig{
		Enabled:        true,
		Concurrent:     3,
		TimeoutMinutes: 10,
		SkipOnError:    false,
		AutoInstall:    true,
	},
	UI: UIConfig{
		Theme:           "default",
		ShowIcons:       true,
		AnimationsSpeed: "normal",
		MouseSupport:    true,
		ShowLineNumbers: false,
	},
	Defaults: DefaultsConfig{
		SearchSort:     "stars",
		SearchOrder:    "desc",
		ResultsPerPage: 30,
		PreferredAuth:  "https",
	},
}

func Load() (*Config, error) {
	config := DefaultConfig

	configPath, err := getConfigPath()
	if err != nil {
		return &config, nil // Return default config if we can't find config path
	}

	config.ConfigPath = configPath

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &config, nil // Return default config if file doesn't exist
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save() error {
	if c.ConfigPath == "" {
		var err error
		c.ConfigPath, err = getConfigPath()
		if err != nil {
			return err
		}
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(c.ConfigPath), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(c.ConfigPath, data, 0644)
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".quikgit", "config.yaml"), nil
}

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".quikgit")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}
