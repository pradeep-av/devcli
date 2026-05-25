package config

import (
	"os"
	"path/filepath"
)

// GetGlobalConfigPath returns the absolute path to the global config file (~/.config/devcli/config.yaml).
func GetGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "devcli", "config.yaml"), nil
}

// FindLocalConfig looks for a .devcli.yaml file in the current working directory,
// and traverses upwards until it finds one or reaches the root directory.
func FindLocalConfig() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		localPath := filepath.Join(dir, ".devcli.yaml")
		if _, err := os.Stat(localPath); err == nil {
			return localPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", nil
}
