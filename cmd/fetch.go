package cmd

import (
        "context"
        "fmt"
        "runtime"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/robertguss/ai-news-agent-cli/internal/ai/processor"
        "github.com/robertguss/ai-news-agent-cli/internal/ai/processor/mocks"
        "github.com/robertguss/ai-news-agent-cli/internal/config"
        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/robertguss/ai-news-agent-cli/internal/fetcher"
        "github.com/robertguss/ai-news-agent-cli/internal/scraper"
        "github.com/robertguss/ai-news-agent-cli/internal/tui"
        "github.com/robertguss/ai-news-agent-cli/internal/tui/fetchui"
        "github.com/robertguss/ai-news-agent-cli/pkg/errs"
        "github.com/robertguss/ai-news-agent-cli/pkg/logging"
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
                workers, _ := cmd.Flags().GetInt("workers")
                limit, _ := cmd.Flags().GetInt("limit")

                cfg, err := loadCfg(configPath)
                if err != nil {
                        return fmt.Errorf("%s", errs.GetUserFriendlyMessage(errs.Wrap("load config", err)))
                }

                if err := logging.Init(cfg.LogFile); err != nil {
                        return fmt.Errorf("failed to initialize logging: %w", err)
                }

                db, queries, err := openDB(cfg.DSN)
                if err != nil {
                        return fmt.Errorf("%s", errs.GetUserFriendlyMessage(err))
                }
                defer db.Close()

                if err := initDB(db); err != nil {
                        return fmt.Errorf("%s", errs.GetUserFriendlyMessage(err))
                }

                var aiProcessor processor.AIProcessor

                if useMockAI {
                        mockProcessor := new(mocks.AIProcessor)
                        mockProcessor.On("AnalyzeContent", mock.Anything).Return(&processor.AnalysisResult{
                                Summary: "mock summary",
                        }, nil)
                        mockProcessor.On("AnalyzeContentWithRetry", mock.Anything, mock.Anything, mock.Anything).Return(&processor.AnalysisResult{
                                Summary: "mock summary with retry",
                        }, nil)
                        aiProcessor = mockProcessor
                } else {
                        var err error
                        aiProcessor, err = processor.NewGeminiProcessor(ctx)
                        if err != nil {
                                return fmt.Errorf("%s", errs.GetUserFriendlyMessage(errs.Wrap("initialize AI processor", err)))
                        }
                }

                opts := fetcher.FetchOptions{Limit: limit}

                if !plain && tui.ShouldUseTUI() {
                        return runInteractiveFetch(ctx, cfg, queries, aiProcessor, workers, opts)
                }

                return runPlainFetch(ctx, cmd, cfg, queries, aiProcessor, opts)
        },
}

func runInteractiveFetch(ctx context.Context, cfg *config.Config, queries *database.Queries, aiProcessor processor.AIProcessor, workers int, opts fetcher.FetchOptions) error {
        if workers <= 0 {
                workers = runtime.NumCPU()
        }

        sourceNames := make([]string, len(cfg.Sources))
        for i, source := range cfg.Sources {
                sourceNames[i] = source.Name
        }

        model := fetchui.New(sourceNames)
        model.SetWorkerCount(workers)

        program := tea.NewProgram(model, tea.WithAltScreen())

        progress := make(chan tui.ArticleProgressMsg, 100)

        go func() {
                for msg := range progress {
                        program.Send(msg)
                }
        }()

        go func() {
                defer close(progress)

                var totalAdded int
                var errors []error
                successCount := 0
                errorCount := 0

                processSource := func(ctx context.Context, source fetcher.Source, opts fetcher.FetchOptions, progressCh chan<- tui.DetailedProgressMsg) (int, error) {
                        deps := fetcher.PipelineDeps{
                                Scraper: scraper.NewJinaScraper(),
                                AI:      aiProcessor,
                                Queries: queries,
                                Config:  cfg,
                        }

                        return fetcher.FetchAndStoreWithAIProgress(ctx, deps, source, opts, progressCh)
                }

                detailedProgress := make(chan tui.DetailedProgressMsg, 100)

                go func() {
                        for msg := range detailedProgress {
                                progress <- tui.ArticleProgressMsg{
                                        Source:       msg.Source,
                                        Phase:        msg.Phase,
                                        Current:      msg.Current,
                                        Total:        msg.Total,
                                        ArticleTitle: msg.ArticleTitle,
                                        Error:        msg.Error,
                                }
                        }
                }()

                results := fetcher.ProcessSourcesConcurrently(ctx, cfg.Sources, workers, processSource, opts, detailedProgress)
                close(detailedProgress)

                for _, result := range results {
                        if result.Error != nil {
                                errorCount++
                                errors = append(errors, fmt.Errorf("source %s: %w", result.Source.Name, result.Error))
                                program.Send(tui.CompletedMsg{
                                        Source: result.Source.Name,
                                        Error:  result.Error,
                                })
                        } else {
                                successCount++
                                totalAdded += result.Added
                                program.Send(tui.CompletedMsg{
                                        Source: result.Source.Name,
                                        Added:  result.Added,
                                })
                        }
                }

                program.Send(tui.FinalSummaryMsg{
                        TotalAdded:   totalAdded,
                        TotalSources: len(cfg.Sources),
                        SuccessCount: successCount,
                        ErrorCount:   errorCount,
                        Errors:       errors,
                })
        }()

        _, err := program.Run()
        return err
}

func runPlainFetch(ctx context.Context, cmd *cobra.Command, cfg *config.Config, queries *database.Queries, aiProcessor processor.AIProcessor, opts fetcher.FetchOptions) error {
        var added int
        var errors []error

        for _, source := range cfg.Sources {
                deps := fetcher.PipelineDeps{
                        Scraper: scraper.NewJinaScraper(),
                        AI:      aiProcessor,
                        Queries: queries,
                        Config:  cfg,
                }

                n, err := fetcher.FetchAndStoreWithAI(ctx, deps, source, opts)
                if err != nil {
                        userFriendlyErr := errs.GetUserFriendlyMessage(err)
                        errors = append(errors, fmt.Errorf("source %s: %s", source.Name, userFriendlyErr))
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
        fetchCmd.Flags().IntP("workers", "w", 0, "Number of worker goroutines (0 = auto-detect based on CPU cores)")
        fetchCmd.Flags().IntP("limit", "n", 5, "Maximum number of articles to fetch per source (0 = unlimited)")
        rootCmd.AddCommand(fetchCmd)
}
