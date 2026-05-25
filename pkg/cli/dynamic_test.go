package cli

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"userId", "user-id"},
		{"user_id", "user-id"},
		{"APIKey", "apikey"},
		{"apiKey", "api-key"},
		{"simple", "simple"},
		{"camelCaseString", "camel-case-string"},
	}

	for _, tt := range tests {
		actual := toKebabCase(tt.input)
		if actual != tt.expected {
			t.Errorf("toKebabCase(%q) = %q; want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestExtractPlaceholders(t *testing.T) {
	input := []string{
		"/users/{id}",
		"fields={fields}&v={version}",
		`{"name": "{name}", "email": "{email}"}`,
	}
	expected := []string{"id", "fields", "version", "name", "email"}

	actual := extractPlaceholders(input...)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("extractPlaceholders() = %v; want %v", actual, expected)
	}
}

func TestResolveTemplate(t *testing.T) {
	tpl := "/users/{id}/details?fields={fields}"
	vars := map[string]string{
		"id":     "123",
		"fields": "name,email",
	}
	expected := "/users/123/details?fields=name,email"

	actual := resolveTemplate(tpl, vars)
	if actual != expected {
		t.Errorf("resolveTemplate() = %q; want %q", actual, expected)
	}
}

func TestCleanJSON(t *testing.T) {
	jsonStr := `{
		"name": "Alice",
		"email": "{email}",
		"role": "{role}",
		"tags": ["admin", "{tag1}"]
	}`

	var val any
	if err := json.Unmarshal([]byte(jsonStr), &val); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	cleaned, _ := cleanJSON(val)
	cleanedBytes, err := json.Marshal(cleaned)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(cleanedBytes, &result); err != nil {
		t.Fatalf("failed to unmarshal cleaned JSON: %v", err)
	}

	if result["name"] != "Alice" {
		t.Errorf("expected 'name' to be 'Alice'; got %v", result["name"])
	}
	if _, exists := result["email"]; exists {
		t.Errorf("expected 'email' to be removed")
	}
	if _, exists := result["role"]; exists {
		t.Errorf("expected 'role' to be removed")
	}

	tags, ok := result["tags"].([]any)
	if !ok {
		t.Fatalf("expected 'tags' to be a slice; got %T", result["tags"])
	}
	if len(tags) != 1 || tags[0] != "admin" {
		t.Errorf("expected 'tags' to be ['admin']; got %v", tags)
	}
}
