package cmd

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/robertguss/rss-agent-cli/internal/browserutil"
	"github.com/robertguss/rss-agent-cli/internal/state"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <article-number>",
	Short: "Open an article in your default browser",
	Long: `Open an article from the last view in your default web browser.

The article number corresponds to the numbers shown in the 'view' command output.
You must run 'view' first to populate the article list.

Examples:
  ai-news open 1    # Opens article #1 from last view
  ai-news open 5    # Opens article #5 from last view`,
	Args: cobra.ExactArgs(1),
	RunE: runOpen,
}

func runOpen(cmd *cobra.Command, args []string) error {
	if _, err := strconv.Atoi(args[0]); err != nil {
		return fmt.Errorf("invalid article number %q: must be a positive integer", args[0])
	}
	key := args[0]

	vs, err := state.Load()
	if err != nil {
		return fmt.Errorf("failed to load view state: %w", err)
	}

	if len(vs.Articles) == 0 {
		return errors.New("no viewed articles found - run 'ai-news view' first to see available articles")
	}

	ref, ok := vs.Articles[key]
	if !ok {
		return fmt.Errorf("article %s not found in last view - available articles: run 'ai-news view' to see current list", key)
	}

	if ref.URL == "" {
		return fmt.Errorf("article %s has no URL available", key)
	}

	if err := browserutil.OpenURL(ref.URL); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Opening: %s\n%s\n", ref.Title, ref.URL)
	return nil
}

func init() {
	rootCmd.AddCommand(openCmd)
}
