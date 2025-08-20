package cmd

import (
        "context"
        "errors"
        "fmt"
        "strconv"
        "time"

        "github.com/robertguss/ai-news-agent-cli/internal/article"
        "github.com/robertguss/ai-news-agent-cli/internal/state"
        "github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
        Use:   "read <article-number>",
        Short: "Read full article content in terminal",
        Long: `Read the full content of an article from the last view in your terminal.

The article number corresponds to the numbers shown in the 'view' command output.
You must run 'view' first to populate the article list.

The content is fetched using Jina Reader and displayed with beautiful markdown formatting.
If cached content exists, it will be used unless --no-cache is specified.

Examples:
  ai-news read 1              # Read article #1 with styling
  ai-news read 5 --no-style   # Read article #5 as plain text
  ai-news read 3 --no-cache   # Force fresh fetch of article #3`,
        Args: cobra.ExactArgs(1),
        RunE: runRead,
}

func runRead(cmd *cobra.Command, args []string) error {
        if _, err := strconv.Atoi(args[0]); err != nil {
                return fmt.Errorf("invalid article number %q: must be a positive integer", args[0])
        }
        key := args[0]

        noStyle, _ := cmd.Flags().GetBool("no-style")
        noCache, _ := cmd.Flags().GetBool("no-cache")

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

        content, err := getArticleContent(cmd.Context(), ref, noCache)
        if err != nil {
                return fmt.Errorf("failed to get article content: %w", err)
        }

        if content != ref.Content {
                ref.Content = content
                now := time.Now()
                ref.ContentFetchedAt = &now
                vs.Articles[key] = ref
                
                if saveErr := state.Save(vs); saveErr != nil {
                        fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to cache content: %v\n", saveErr)
                }
        }

        fmt.Fprintf(cmd.OutOrStdout(), "Reading: %s\n\n", ref.Title)
        
        if err := article.RenderMarkdown(content, !noStyle, cmd.OutOrStdout()); err != nil {
                return fmt.Errorf("failed to render content: %w", err)
        }

        return nil
}

func getArticleContent(ctx context.Context, ref state.ArticleRef, noCache bool) (string, error) {
        if !noCache && ref.Content != "" && ref.ContentFetchedAt != nil {
                if time.Since(*ref.ContentFetchedAt) < 24*time.Hour {
                        return ref.Content, nil
                }
        }

        ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
        defer cancel()

        return article.FetchArticle(ctx, ref.URL, noCache)
}

func init() {
        readCmd.Flags().Bool("no-style", false, "Display content as plain text without markdown styling")
        readCmd.Flags().Bool("no-cache", false, "Force fresh fetch instead of using cached content")
        rootCmd.AddCommand(readCmd)
}
