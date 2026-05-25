package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAlignAliases(t *testing.T) {
	yamlData := `
current_profile: default
profiles:
  default:
    url: "http://localhost:8080"
    tokenScript: "echo 'my-token'"
`

	cfg := &Config{
		Profiles: make(map[string]Profile),
		Commands: make(map[string]CommandConfig),
	}

	err := yaml.Unmarshal([]byte(yamlData), cfg)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	// Verify that before alignment, TokenScriptCamel is set
	if cfg.Profiles["default"].TokenScriptCamel != "echo 'my-token'" {
		t.Errorf("expected TokenScriptCamel to be 'echo \\'my-token\\'', got %q", cfg.Profiles["default"].TokenScriptCamel)
	}

	alignAliases(cfg)

	// Verify that after alignment, TokenScript is aligned
	if cfg.Profiles["default"].TokenScript != "echo 'my-token'" {
		t.Errorf("expected TokenScript to be aligned to 'echo \\'my-token\\'', got %q", cfg.Profiles["default"].TokenScript)
	}
}
