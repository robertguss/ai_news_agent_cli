package cmd

import (
        "context"
        "fmt"

        "github.com/robertguss/ai-news-agent-cli/internal/config"
        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/robertguss/ai-news-agent-cli/internal/fetcher"
        "github.com/spf13/cobra"
)

var (
        openDB   = database.Open
        loadCfg  = config.LoadFromPath
        fetchAnd = fetcher.FetchAndStore
        initDB   = database.InitSchema
)

var fetchCmd = &cobra.Command{
        Use:   "fetch",
        Short: "Fetch articles from configured sources and store them",
        RunE: func(cmd *cobra.Command, args []string) error {
                ctx := context.Background()
                
                configPath, _ := cmd.Flags().GetString("config")
                
                cfg, err := loadCfg(configPath)
                if err != nil {
                        return fmt.Errorf("load config: %w", err)
                }
                
                db, queries, err := openDB(cfg.DSN)
                if err != nil {
                        return fmt.Errorf("open db: %w", err)
                }
                defer db.Close()
                
                if err := initDB(db); err != nil {
                        return fmt.Errorf("init schema: %w", err)
                }
                
                var added int
                var errors []error
                
                for _, source := range cfg.Sources {
                        n, err := fetchAnd(ctx, queries, source)
                        if err != nil {
                                errors = append(errors, fmt.Errorf("source %s: %w", source.Name, err))
                                continue
                        }
                        added += n
                }
                
                if len(errors) > 0 {
                        fmt.Fprintf(cmd.OutOrStdout(), "Added %d new articles from %d sources\n", added, len(cfg.Sources))
                        fmt.Fprintf(cmd.OutOrStdout(), "%d errors occurred:\n", len(errors))
                        for _, err := range errors {
                                fmt.Fprintf(cmd.OutOrStdout(), "  - %v\n", err)
                        }
                } else {
                        fmt.Fprintf(cmd.OutOrStdout(), "Added %d new articles from %d sources\n", added, len(cfg.Sources))
                }
                
                return nil
        },
}

func init() {
        fetchCmd.Flags().StringP("config", "c", "", "Path to config file")
        rootCmd.AddCommand(fetchCmd)
}
