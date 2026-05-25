package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/pradeep-av/devcli/pkg/config"
)

var (
	// Verbose controls request/response detail logging
	Verbose bool
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "devcli",
	Short: "devcli simplifies REST API calls with dynamic templated commands",
	Long:  `devcli is a Go CLI designed to supercharge developer API testing with templated paths, query parameters, bodies, and profile switching.`,
}

// Execute is the main entry point for running the CLI.
func Execute() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Persistent flags
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Print detailed request and response logging")

	// Register static subcommands
	rootCmd.AddCommand(newProfileCmd())

	// Dynamically register templated subcommands from .devcli.yaml
	if err := registerDynamicCommands(rootCmd, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering dynamic commands: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
