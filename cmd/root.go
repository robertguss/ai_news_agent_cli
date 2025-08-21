package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

// NewRootCmd creates and returns the root cobra command for the RSS agent CLI.
// It sets up the basic command structure with version information and default behavior.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rss-agent-cli",
		Short: "AI-powered RSS aggregation CLI",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "RSS Agent")
		},
	}
	cmd.Version = version
	return cmd
}

var rootCmd = NewRootCmd()

// Execute runs the root command and handles any execution errors.
// This is the main entry point for the CLI application.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
