package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pradeep-av/devcli/pkg/client"
	"github.com/pradeep-av/devcli/pkg/config"
)

var placeholderRegex = regexp.MustCompile(`\{([a-zA-Z0-9_-]+)\}`)

func registerDynamicCommands(root *cobra.Command, cfg *config.Config) error {
	for name, cmdConf := range cfg.Commands {
		cmd := createDynamicCommand(name, cmdConf, cfg)
		root.AddCommand(cmd)
	}
	return nil
}

func createDynamicCommand(name string, cmdConf config.CommandConfig, cfg *config.Config) *cobra.Command {
	// Gather all template strings to inspect for placeholders
	var templateStrings []string
	templateStrings = append(templateStrings, cmdConf.Path)
	for _, v := range cmdConf.Query {
		templateStrings = append(templateStrings, v)
	}
	for _, v := range cmdConf.Headers {
		templateStrings = append(templateStrings, v)
	}
	if cmdConf.Body != "" {
		templateStrings = append(templateStrings, cmdConf.Body)
	}

	placeholders := extractPlaceholders(templateStrings...)

	// Track placeholders present in the path to enforce "required" by default
	inPath := make(map[string]bool)
	for _, m := range placeholderRegex.FindAllStringSubmatch(cmdConf.Path, -1) {
		if len(m) > 1 {
			inPath[m[1]] = true
		}
	}

	cmd := &cobra.Command{
		Use:   name,
		Short: cmdConf.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Resolve target profile
			profileName := cfg.CurrentProfile
			if profileName == "" {
				return fmt.Errorf("no active profile set; use 'devcli profile use <name>' or 'devcli profile set <name>'")
			}
			profile, exists := cfg.Profiles[profileName]
			if !exists {
				return fmt.Errorf("profile %q not found", profileName)
			}
			if profile.URL == "" {
				return fmt.Errorf("profile %q has no URL configured", profileName)
			}

			// 2. Retrieve variables from flags
			vars := make(map[string]string)
			for _, pName := range placeholders {
				flagName := toKebabCase(pName)
				argConf, hasArgConf := cmdConf.Args[pName]
				if hasArgConf && argConf.Type == "bool" {
					val, err := cmd.Flags().GetBool(flagName)
					if err == nil {
						vars[pName] = fmt.Sprintf("%t", val)
					}
				} else if hasArgConf && argConf.Type == "int" {
					val, err := cmd.Flags().GetInt(flagName)
					if err == nil {
						vars[pName] = fmt.Sprintf("%d", val)
					}
				} else {
					val, err := cmd.Flags().GetString(flagName)
					if err == nil {
						vars[pName] = val
					}
				}
			}

			// 3. Resolve path variables
			resolvedPath := resolveTemplate(cmdConf.Path, vars)

			// 4. Resolve query params (remove param if placeholder value not set/provided)
			resolvedQuery := make(map[string]string)
			for k, v := range cmdConf.Query {
				resolvedVal := resolveTemplate(v, vars)
				if placeholderRegex.MatchString(resolvedVal) {
					continue
				}
				resolvedQuery[k] = resolvedVal
			}

			// 5. Build and resolve headers
			resolvedHeaders := make(map[string]string)
			// Apply profile-wide custom headers
			for k, v := range profile.Headers {
				resolvedHeaders[k] = v
			}
			// Apply profile auth token
			if profile.Token != "" {
				resolvedHeaders["Authorization"] = "Bearer " + profile.Token
			}
			// Apply command-specific headers
			for k, v := range cmdConf.Headers {
				resolvedVal := resolveTemplate(v, vars)
				if placeholderRegex.MatchString(resolvedVal) {
					continue
				}
				resolvedHeaders[k] = resolvedVal
			}

			// 6. Resolve and clean request body
			var resolvedBody string
			if cmdConf.Body != "" {
				resolvedBody = resolveTemplate(cmdConf.Body, vars)
				
				// Attempt to clean JSON body from empty optional keys containing unresolved placeholders
				var jsonVal any
				if err := json.Unmarshal([]byte(resolvedBody), &jsonVal); err == nil {
					cleaned, _ := cleanJSON(jsonVal)
					if cleanedBytes, err := json.Marshal(cleaned); err == nil {
						resolvedBody = string(cleanedBytes)
					}
				}
			}

			// 7. Formulate request URL
			baseURL := strings.TrimSuffix(profile.URL, "/")
			requestURL := baseURL + "/" + strings.TrimPrefix(resolvedPath, "/")

			if len(resolvedQuery) > 0 {
				qVals := url.Values{}
				for k, v := range resolvedQuery {
					qVals.Set(k, v)
				}
				requestURL += "?" + qVals.Encode()
			}

			// 8. Execute request
			return client.Send(client.RequestOptions{
				Method:  cmdConf.Method,
				URL:     requestURL,
				Headers: resolvedHeaders,
				Body:    resolvedBody,
				Verbose: Verbose,
			})
		},
	}

	// Register command-line flags dynamically
	for _, pName := range placeholders {
		flagName := toKebabCase(pName)
		argConf, hasArgConf := cmdConf.Args[pName]

		desc := fmt.Sprintf("Value for placeholder {%s}", pName)
		if hasArgConf && argConf.Description != "" {
			desc = argConf.Description
		}

		isRequired := false
		if inPath[pName] {
			isRequired = true
		}
		if hasArgConf && argConf.Required {
			isRequired = true
		}

		if hasArgConf && argConf.Type == "bool" {
			var defVal bool
			if argConf.Default == "true" {
				defVal = true
			}
			cmd.Flags().Bool(flagName, defVal, desc)
		} else if hasArgConf && argConf.Type == "int" {
			var defVal int
			fmt.Sscanf(argConf.Default, "%d", &defVal)
			cmd.Flags().Int(flagName, defVal, desc)
		} else {
			defVal := ""
			if hasArgConf {
				defVal = argConf.Default
			}
			cmd.Flags().String(flagName, defVal, desc)
		}

		if isRequired {
			_ = cmd.MarkFlagRequired(flagName)
		}
	}

	return cmd
}

// extractPlaceholders parses a set of strings and extracts all distinct occurrences of {placeholder}
func extractPlaceholders(texts ...string) []string {
	seen := make(map[string]bool)
	var list []string
	for _, text := range texts {
		matches := placeholderRegex.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			if len(m) > 1 {
				name := m[1]
				if !seen[name] {
					seen[name] = true
					list = append(list, name)
				}
			}
		}
	}
	return list
}

// toKebabCase converts camelCase or snake_case string into kebab-case
func toKebabCase(s string) string {
	s = strings.ReplaceAll(s, "_", "-")
	var res strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := s[i-1]
			if (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9') {
				res.WriteRune('-')
			}
		}
		res.WriteRune(r)
	}
	return strings.ToLower(res.String())
}

// resolveTemplate substitutes placeholder strings with their corresponding values
func resolveTemplate(tpl string, vars map[string]string) string {
	res := tpl
	for k, v := range vars {
		res = strings.ReplaceAll(res, "{"+k+"}", v)
	}
	return res
}

// cleanJSON recursively deletes keys that contain unresolved `{placeholder}` strings from map/slice objects
func cleanJSON(val any) (any, bool) {
	switch m := val.(type) {
	case map[string]any:
		for k, v := range m {
			if s, ok := v.(string); ok && placeholderRegex.MatchString(s) {
				delete(m, k)
				continue
			}
			cleaned, keep := cleanJSON(v)
			if !keep {
				delete(m, k)
			} else {
				m[k] = cleaned
			}
		}
		return m, true
	case []any:
		var newList []any
		for _, item := range m {
			if s, ok := item.(string); ok && placeholderRegex.MatchString(s) {
				continue
			}
			cleaned, keep := cleanJSON(item)
			if keep {
				newList = append(newList, cleaned)
			}
		}
		return newList, true
	default:
		return val, true
	}
}
