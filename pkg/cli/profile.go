package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/pradeep-av/devcli/pkg/config"
)

func newProfileCmd() *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage target environment profiles",
		Long:  `Configure API target environments, base URLs, auth tokens, and headers.`,
	}

	profileCmd.AddCommand(newProfileSetCmd())
	profileCmd.AddCommand(newProfileListCmd())
	profileCmd.AddCommand(newProfileUseCmd())

	return profileCmd
}

func newProfileSetCmd() *cobra.Command {
	var urlFlag string
	var tokenFlag string
	var tokenScriptFlag string
	var headersFlag []string

	cmd := &cobra.Command{
		Use:   "set [profile-name]",
		Short: "Create or update a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Load fresh configuration
			localCfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			// Get or initialize profile
			p, exists := localCfg.Profiles[name]
			if !exists {
				p = config.Profile{
					Headers: make(map[string]string),
				}
			}

			if urlFlag != "" {
				p.URL = urlFlag
			}
			if tokenFlag != "" {
				p.Token = tokenFlag
			}
			if tokenScriptFlag != "" {
				p.TokenScript = tokenScriptFlag
				p.TokenScriptCamel = tokenScriptFlag
			}

			// Parse custom headers specified via CLI flags
			for _, h := range headersFlag {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) != 2 {
					parts = strings.SplitN(h, "=", 2)
				}
				if len(parts) != 2 {
					return fmt.Errorf("invalid header format %q, use 'Key: Value' or 'Key=Value'", h)
				}
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				if p.Headers == nil {
					p.Headers = make(map[string]string)
				}
				p.Headers[k] = v
			}

			localCfg.Profiles[name] = p

			// Set active profile automatically if none is set
			if localCfg.CurrentProfile == "" {
				localCfg.CurrentProfile = name
			}

			err = config.SaveConfig(localCfg)
			if err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			color.Green("Profile %q successfully updated.", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&urlFlag, "url", "", "API server base URL (e.g. http://localhost:8080)")
	cmd.Flags().StringVar(&tokenFlag, "token", "", "API Authorization token")
	cmd.Flags().StringVar(&tokenScriptFlag, "token-script", "", "Shell command to run to generate the token dynamically")
	cmd.Flags().StringSliceVar(&headersFlag, "header", nil, "Custom headers in format 'Key: Value' (can be specified multiple times)")

	return cmd
}

func newProfileListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			localCfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			if len(localCfg.Profiles) == 0 {
				fmt.Println("No profiles found. Use 'devcli profile set <name>' to create one.")
				return nil
			}

			fmt.Println("Available Profiles:")
			for name, p := range localCfg.Profiles {
				isActive := name == localCfg.CurrentProfile
				prefix := "  "
				nameColor := color.New(color.FgCyan)
				if isActive {
					prefix = "* "
					nameColor = color.New(color.FgGreen, color.Bold)
				}

				fmt.Print(prefix)
				nameColor.Printf("%-15s", name)
				fmt.Printf(" URL: %s", p.URL)
				if p.Token != "" {
					fmt.Printf(" | Token: ********")
				}
				if p.TokenScript != "" {
					fmt.Printf(" | TokenScript: %s", p.TokenScript)
				}
				if len(p.Headers) > 0 {
					var hKeys []string
					for k := range p.Headers {
						hKeys = append(hKeys, k)
					}
					fmt.Printf(" | Headers: [%s]", strings.Join(hKeys, ", "))
				}
				fmt.Println()
			}
			return nil
		},
	}
}

func newProfileUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use [profile-name]",
		Short: "Switch active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			localCfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			if _, exists := localCfg.Profiles[name]; !exists {
				return fmt.Errorf("profile %q does not exist", name)
			}

			localCfg.CurrentProfile = name
			err = config.SaveConfig(localCfg)
			if err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			color.Green("Switched to profile %q.", name)
			return nil
		},
	}
}
