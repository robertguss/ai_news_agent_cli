package cmd

import (
        "context"
        "fmt"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/robertguss/ai-news-agent-cli/internal/ai/processor"
        "github.com/robertguss/ai-news-agent-cli/internal/ai/processor/mocks"
        "github.com/robertguss/ai-news-agent-cli/internal/config"
        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/robertguss/ai-news-agent-cli/internal/fetcher"
        "github.com/robertguss/ai-news-agent-cli/internal/scraper"
        "github.com/robertguss/ai-news-agent-cli/internal/tui"
        "github.com/robertguss/ai-news-agent-cli/internal/tui/fetchui"
        "github.com/spf13/cobra"
        "github.com/stretchr/testify/mock"
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
                useMockAI, _ := cmd.Flags().GetBool("use-mock-ai")
                plain, _ := cmd.Flags().GetBool("plain")

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

                var aiProcessor processor.AIProcessor
                
                if useMockAI {
                        mockProcessor := new(mocks.AIProcessor)
                        mockProcessor.On("AnalyzeContent", mock.Anything).Return(&processor.AnalysisResult{
                                Summary: "mock summary",
                        }, nil)
                        aiProcessor = mockProcessor
                } else {
                        var err error
                        aiProcessor, err = processor.NewGeminiProcessor(ctx)
                        if err != nil {
                                return fmt.Errorf("failed to initialize Gemini processor: %w", err)
                        }
                }

                if !plain && tui.ShouldUseTUI() {
                        return runInteractiveFetch(ctx, cfg, queries, aiProcessor)
                }

                return runPlainFetch(ctx, cmd, cfg, queries, aiProcessor)
        },
}

func runInteractiveFetch(ctx context.Context, cfg *config.Config, queries *database.Queries, aiProcessor processor.AIProcessor) error {
        sourceNames := make([]string, len(cfg.Sources))
        for i, source := range cfg.Sources {
                sourceNames[i] = source.Name
        }

        model := fetchui.New(sourceNames)
        
        program := tea.NewProgram(model, tea.WithAltScreen())
        
        go func() {
                var added int
                var errors []error
                successCount := 0
                errorCount := 0

                for _, source := range cfg.Sources {
                        program.Send(tui.ProgressMsg{
                                Source: source.Name,
                                Status: "Starting...",
                        })

                        deps := fetcher.PipelineDeps{
                                Scraper: scraper.NewJinaScraper(),
                                AI:      aiProcessor,
                                Queries: queries,
                        }
                        
                        n, err := fetcher.FetchAndStoreWithAI(ctx, deps, source)
                        if err != nil {
                                errorCount++
                                errors = append(errors, fmt.Errorf("source %s: %w", source.Name, err))
                                program.Send(tui.CompletedMsg{
                                        Source: source.Name,
                                        Error:  err,
                                })
                                continue
                        }
                        
                        successCount++
                        added += n
                        program.Send(tui.CompletedMsg{
                                Source: source.Name,
                                Added:  n,
                        })
                }

                program.Send(tui.FinalSummaryMsg{
                        TotalAdded:   added,
                        TotalSources: len(cfg.Sources),
                        SuccessCount: successCount,
                        ErrorCount:   errorCount,
                        Errors:       errors,
                })
        }()

        _, err := program.Run()
        return err
}

func runPlainFetch(ctx context.Context, cmd *cobra.Command, cfg *config.Config, queries *database.Queries, aiProcessor processor.AIProcessor) error {
        var added int
        var errors []error

        for _, source := range cfg.Sources {
                deps := fetcher.PipelineDeps{
                        Scraper: scraper.NewJinaScraper(),
                        AI:      aiProcessor,
                        Queries: queries,
                }
                
                n, err := fetcher.FetchAndStoreWithAI(ctx, deps, source)
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
}

func init() {
        fetchCmd.Flags().StringP("config", "c", "", "Path to config file")
        fetchCmd.Flags().Bool("use-mock-ai", false, "Use mock AI processor for testing")
        fetchCmd.Flags().Bool("plain", false, "Use plain text output instead of interactive TUI")
        rootCmd.AddCommand(fetchCmd)
}
