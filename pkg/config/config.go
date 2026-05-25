package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile represents an API target environment.
type Profile struct {
	URL     string            `yaml:"url"`
	Token   string            `yaml:"token"`
	Headers map[string]string `yaml:"headers"`
}

// ArgConfig defines rules for a command argument/flag.
type ArgConfig struct {
	Type        string `yaml:"type"` // e.g. string, int, bool
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
	Required    bool   `yaml:"required"`
}

// CommandConfig defines a templated REST call configuration.
type CommandConfig struct {
	Path        string               `yaml:"path"`
	Method      string               `yaml:"method"`
	Description string               `yaml:"description"`
	Headers     map[string]string    `yaml:"headers"`
	Query       map[string]string    `yaml:"query"`
	Body        string               `yaml:"body"`
	Args        map[string]ArgConfig `yaml:"args"`
}

// Config is the unified configuration containing both profiles and commands in a single YAML.
type Config struct {
	FilePath       string                   `yaml:"-"`
	CurrentProfile string                   `yaml:"current_profile"`
	Profiles       map[string]Profile       `yaml:"profiles"`
	Commands       map[string]CommandConfig `yaml:"commands"`
}

// LoadConfig loads the unified configuration, checking for a local config first, then falling back to global.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]Profile),
		Commands: make(map[string]CommandConfig),
	}

	// 1. Look for local configuration (.devcli.yaml)
	localPath, err := FindLocalConfig()
	if err == nil && localPath != "" {
		cfg.FilePath = localPath
		data, errRead := os.ReadFile(localPath)
		if errRead == nil {
			_ = yaml.Unmarshal(data, cfg)
		}
		// Ensure maps are initialized if nil after unmarshal
		if cfg.Profiles == nil {
			cfg.Profiles = make(map[string]Profile)
		}
		if cfg.Commands == nil {
			cfg.Commands = make(map[string]CommandConfig)
		}
		return cfg, nil
	}

	// 2. Fall back to global configuration (~/.config/devcli/config.yaml)
	gPath, err := GetGlobalConfigPath()
	if err != nil {
		return nil, err
	}
	cfg.FilePath = gPath

	if _, errStat := os.Stat(gPath); errStat == nil {
		data, errRead := os.ReadFile(gPath)
		if errRead == nil {
			_ = yaml.Unmarshal(data, cfg)
		}
	}

	// Ensure maps are initialized if nil after unmarshal
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}
	if cfg.Commands == nil {
		cfg.Commands = make(map[string]CommandConfig)
	}

	return cfg, nil
}

// SaveConfig saves the configuration back to its original file source.
func SaveConfig(cfg *Config) error {
	if cfg.FilePath == "" {
		gPath, err := GetGlobalConfigPath()
		if err != nil {
			return err
		}
		cfg.FilePath = gPath
	}

	dir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.FilePath, data, 0600)
}
