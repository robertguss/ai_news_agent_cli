package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-news-agent-cli",
	Short: "AI-powered news aggregation CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ai-news-agent-cli -- help with `-h`")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
