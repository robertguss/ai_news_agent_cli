package fetcher

import (
        "context"
        "database/sql"
        "fmt"
        "sort"
        "sync"
        "time"

        "github.com/mmcdole/gofeed"
        "github.com/robertguss/rss-agent-cli/internal/ai/processor"
        "github.com/robertguss/rss-agent-cli/internal/config"
        "github.com/robertguss/rss-agent-cli/internal/database"
        "github.com/robertguss/rss-agent-cli/internal/scraper"
        "github.com/robertguss/rss-agent-cli/internal/tui"
        "github.com/robertguss/rss-agent-cli/pkg/errs"
        "github.com/robertguss/rss-agent-cli/pkg/logging"
        "github.com/robertguss/rss-agent-cli/pkg/retry"
)

type Source = config.Source

// FetchOptions holds configuration options for fetching articles.
type FetchOptions struct {
        Limit int // Maximum articles to fetch per source (0 = unlimited)
}

// Article represents a news article with basic metadata from RSS feeds.
type Article struct {
        Title         string
        Link          string
        PublishedDate time.Time
}

// PipelineDeps holds dependencies for the AI-enhanced article processing pipeline.
type PipelineDeps struct {
        Scraper scraper.Scraper
        AI      processor.AIProcessor
        Queries *database.Queries
        Config  *config.Config
}

// Fetch retrieves articles from an RSS feed source with timeout and retry logic.
// It parses the RSS feed and returns a list of articles with metadata.
func Fetch(ctx context.Context, source Source, cfg *config.Config, opts FetchOptions) ([]Article, error) {
        ctx, cancel := context.WithTimeout(ctx, cfg.NetworkTimeout)
        defer cancel()

        var feed *gofeed.Feed
        err := retry.DoWithCallback(ctx, cfg.RetryConfig(), func() error {
                parser := gofeed.NewParser()
                var e error
                feed, e = parser.ParseURLWithContext(source.URL, ctx)
                return e
        }, func(attempt int, err error) {
                logging.Retry("fetch_rss", attempt, err)
        })

        if err != nil {
                wrappedErr := errs.Wrap("fetch rss "+source.URL, err)
                logging.Error("fetch_rss", wrappedErr)
                return nil, wrappedErr
        }

        sort.SliceStable(feed.Items, func(i, j int) bool {
                ti := feed.Items[i].PublishedParsed
                tj := feed.Items[j].PublishedParsed
                if ti == nil || tj == nil {
                        return false
                }
                return ti.After(*tj)
        })

        limit := opts.Limit
        if limit == 0 || limit > len(feed.Items) {
                limit = len(feed.Items)
        }

        articles := make([]Article, 0, limit)
        for i, item := range feed.Items {
                if opts.Limit != 0 && i >= opts.Limit {
                        break
                }

                publishedDate := time.Now()
                if item.PublishedParsed != nil {
                        publishedDate = *item.PublishedParsed
                }

                article := Article{
                        Title:         item.Title,
                        Link:          item.Link,
                        PublishedDate: publishedDate,
                }
                articles = append(articles, article)
        }

        if opts.Limit == 0 {
                logging.Info("fetch_rss", fmt.Sprintf("Fetched %d articles from %s", len(articles), source.Name))
        } else {
                logging.Info("fetch_rss", fmt.Sprintf("Fetched %d (limit=%d) articles from %s", len(articles), opts.Limit, source.Name))
        }
        return articles, nil
}

func StoreArticles(ctx context.Context, queries *database.Queries, articles []Article, source Source, cfg *config.Config) (int, error) {
        stored := 0

        for _, article := range articles {
                err := retry.Do(ctx, cfg.RetryConfig(), func() error {
                        _, err := queries.GetArticleByUrl(ctx, sql.NullString{
                                String: article.Link,
                                Valid:  true,
                        })

                        if err == sql.ErrNoRows {
                                params := database.CreateArticleParams{
                                        Title: sql.NullString{
                                                String: article.Title,
                                                Valid:  true,
                                        },
                                        Url: sql.NullString{
                                                String: article.Link,
                                                Valid:  true,
                                        },
                                        SourceName: sql.NullString{
                                                String: source.Name,
                                                Valid:  true,
                                        },
                                        PublishedDate: sql.NullTime{
                                                Time:  article.PublishedDate,
                                                Valid: true,
                                        },
                                        Summary: sql.NullString{
                                                String: "",
                                                Valid:  false,
                                        },
                                        Entities: nil,
                                        ContentType: sql.NullString{
                                                String: "",
                                                Valid:  false,
                                        },
                                        Topics: nil,
                                        Status: sql.NullString{
                                                String: "unread",
                                                Valid:  true,
                                        },
                                        AnalysisStatus: sql.NullString{
                                                String: "unprocessed",
                                                Valid:  true,
                                        },
                                        StoryGroupID: sql.NullString{
                                                String: "",
                                                Valid:  false,
                                        },
                                }

                                _, err = queries.CreateArticle(ctx, params)
                                if err != nil {
                                        return errs.Wrap("create article", err)
                                }
                                stored++
                                return nil
                        } else if err != nil {
                                return errs.Wrap("check article exists", err)
                        }
                        return nil
                })

                if err != nil {
                        logging.Error("store_article", err)
                        return stored, err
                }
        }

        return stored, nil
}

// FetchAndStore fetches articles from a source and stores them in the database.
// It combines the fetch and store operations, returning the number of articles stored.
func FetchAndStore(ctx context.Context, queries *database.Queries, source Source, cfg *config.Config, opts FetchOptions) (int, error) {
        articles, err := Fetch(ctx, source, cfg, opts)
        if err != nil {
                return 0, err
        }

        return StoreArticles(ctx, queries, articles, source, cfg)
}

func FetchAndStoreWithAI(ctx context.Context, deps PipelineDeps, source Source, opts FetchOptions) (int, error) {
        articles, err := Fetch(ctx, source, deps.Config, opts)
        if err != nil {
                return 0, err
        }

        return StoreArticlesWithAI(ctx, deps, articles, source)
}

func StoreArticlesWithAI(ctx context.Context, deps PipelineDeps, articles []Article, source Source) (int, error) {
        stored := 0

        for _, article := range articles {
                err := retry.Do(ctx, deps.Config.RetryConfig(), func() error {
                        _, err := deps.Queries.GetArticleByUrl(ctx, sql.NullString{
                                String: article.Link,
                                Valid:  true,
                        })

                        if err == sql.ErrNoRows {
                                var summary sql.NullString
                                var entities []byte
                                var topics []byte
                                var contentType sql.NullString
                                var storyGroupID sql.NullString
                                var analysisStatus = "unprocessed"
                                var articleContent sql.NullString

                                if deps.Scraper != nil && deps.AI != nil {
                                        content, scrapeErr := deps.Scraper.ScrapeWithRetry(ctx, article.Link, deps.Config)
                                        if scrapeErr != nil {
                                                logging.Warn("scrape_article", fmt.Sprintf("Failed to scrape %s: %v", article.Link, scrapeErr))
                                        } else {
                                                // Store the scraped content
                                                articleContent = sql.NullString{
                                                        String: content,
                                                        Valid:  true,
                                                }
                                                
                                                result, aiErr := deps.AI.AnalyzeContentWithRetry(ctx, content, deps.Config)
                                                if aiErr != nil {
                                                        logging.Warn("ai_analysis", fmt.Sprintf("Failed to analyze %s: %v", article.Link, aiErr))
                                                        analysisStatus = "pending"
                                                } else if result != nil {
                                                        summary = sql.NullString{
                                                                String: result.Summary,
                                                                Valid:  true,
                                                        }
                                                        entities = result.EntitiesJSON()
                                                        topics = result.TopicsJSON()
                                                        contentType = sql.NullString{
                                                                String: result.ContentType,
                                                                Valid:  true,
                                                        }
                                                        storyGroupID = sql.NullString{
                                                                String: result.StoryGroupID,
                                                                Valid:  true,
                                                        }
                                                        analysisStatus = "completed"
                                                }
                                        }
                                }

                                params := database.CreateArticleParams{
                                        Title: sql.NullString{
                                                String: article.Title,
                                                Valid:  true,
                                        },
                                        Url: sql.NullString{
                                                String: article.Link,
                                                Valid:  true,
                                        },
                                        SourceName: sql.NullString{
                                                String: source.Name,
                                                Valid:  true,
                                        },
                                        PublishedDate: sql.NullTime{
                                                Time:  article.PublishedDate,
                                                Valid: true,
                                        },
                                        Summary:     summary,
                                        Entities:    entities,
                                        ContentType: contentType,
                                        Topics:      topics,
                                        Status: sql.NullString{
                                                String: "unread",
                                                Valid:  true,
                                        },
                                        AnalysisStatus: sql.NullString{
                                                String: analysisStatus,
                                                Valid:  true,
                                        },
                                        StoryGroupID: storyGroupID,
                                        Content:      articleContent,
                                }

                                _, err = deps.Queries.CreateArticle(ctx, params)
                                if err != nil {
                                        return errs.Wrap("create article with AI", err)
                                }
                                stored++
                                return nil
                        } else if err != nil {
                                return errs.Wrap("check article exists", err)
                        }
                        return nil
                })

                if err != nil {
                        logging.Error("store_article_with_ai", err)
                        return stored, err
                }
        }

        return stored, nil
}

type SourceResult struct {
        Source Source
        Added  int
        Error  error
}

func FetchAndStoreWithAIProgress(ctx context.Context, deps PipelineDeps, source Source, opts FetchOptions, progress chan<- tui.DetailedProgressMsg) (int, error) {
        progress <- tui.DetailedProgressMsg{
                Source: source.Name,
                Phase:  tui.PhaseRSSFetch,
        }

        articles, err := Fetch(ctx, source, deps.Config, opts)
        if err != nil {
                progress <- tui.DetailedProgressMsg{
                        Source: source.Name,
                        Phase:  tui.PhaseRSSFetch,
                        Error:  err,
                }
                return 0, err
        }

        total := len(articles)
        stored := 0

        for i, article := range articles {
                progress <- tui.DetailedProgressMsg{
                        Source:       source.Name,
                        Phase:        tui.PhaseScrape,
                        Current:      i + 1,
                        Total:        total,
                        ArticleTitle: article.Title,
                }

                _, err := deps.Queries.GetArticleByUrl(ctx, sql.NullString{
                        String: article.Link,
                        Valid:  true,
                })

                if err == sql.ErrNoRows {
                        var summary sql.NullString
                        var entities []byte
                        var topics []byte
                        var contentType sql.NullString
                        var storyGroupID sql.NullString
                        var analysisStatus = "unprocessed"
                        var articleContent sql.NullString

                        if deps.Scraper != nil && deps.AI != nil {
                                content, scrapeErr := deps.Scraper.Scrape(article.Link)
                                if scrapeErr != nil {
                                        analysisStatus = "pending"
                                } else {
                                        // Store the scraped content
                                        articleContent = sql.NullString{
                                                String: content,
                                                Valid:  true,
                                        }

                                        progress <- tui.DetailedProgressMsg{
                                                Source:       source.Name,
                                                Phase:        tui.PhaseAI,
                                                Current:      i + 1,
                                                Total:        total,
                                                ArticleTitle: article.Title,
                                        }

                                        result, aiErr := deps.AI.AnalyzeContent(content)
                                        if aiErr != nil {
                                                analysisStatus = "pending"
                                        } else if result != nil {
                                                summary = sql.NullString{
                                                        String: result.Summary,
                                                        Valid:  true,
                                                }
                                                entities = result.EntitiesJSON()
                                                topics = result.TopicsJSON()
                                                contentType = sql.NullString{
                                                        String: result.ContentType,
                                                        Valid:  true,
                                                }
                                                storyGroupID = sql.NullString{
                                                        String: result.StoryGroupID,
                                                        Valid:  true,
                                                }
                                                analysisStatus = "completed"
                                        }
                                }
                        }

                        params := database.CreateArticleParams{
                                Title: sql.NullString{
                                        String: article.Title,
                                        Valid:  true,
                                },
                                Url: sql.NullString{
                                        String: article.Link,
                                        Valid:  true,
                                },
                                SourceName: sql.NullString{
                                        String: source.Name,
                                        Valid:  true,
                                },
                                PublishedDate: sql.NullTime{
                                        Time:  article.PublishedDate,
                                        Valid: true,
                                },
                                Summary:     summary,
                                Entities:    entities,
                                ContentType: contentType,
                                Topics:      topics,
                                Status: sql.NullString{
                                        String: "unread",
                                        Valid:  true,
                                },
                                AnalysisStatus: sql.NullString{
                                        String: analysisStatus,
                                        Valid:  true,
                                },
                                StoryGroupID: storyGroupID,
                                Content:      articleContent,
                        }

                        _, err = deps.Queries.CreateArticle(ctx, params)
                        if err == nil {
                                stored++
                                progress <- tui.DetailedProgressMsg{
                                        Source: source.Name,
                                        Phase:  tui.PhaseAI,
                                        Current: stored,
                                        Total:   total,
                                        ArticleTitle: "Article stored",
                                }
                        }
                } else if err != nil {
                        return stored, err
                }
        }

        progress <- tui.DetailedProgressMsg{
                Source: source.Name,
                Phase:  tui.PhaseDone,
        }

        return stored, nil
}

func ProcessSourcesConcurrently(ctx context.Context, sources []Source, workerCount int, processFunc func(context.Context, Source, FetchOptions, chan<- tui.DetailedProgressMsg) (int, error), opts FetchOptions, progress chan<- tui.DetailedProgressMsg) []SourceResult {
        results := make([]SourceResult, len(sources))
        sourceCh := make(chan int, len(sources))
        var wg sync.WaitGroup

        for i := 0; i < workerCount; i++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        for idx := range sourceCh {
                                source := sources[idx]
                                added, err := processFunc(ctx, source, opts, progress)
                                results[idx] = SourceResult{
                                        Source: source,
                                        Added:  added,
                                        Error:  err,
                                }
                        }
                }()
        }

        for i := range sources {
                sourceCh <- i
        }
        close(sourceCh)

        wg.Wait()
        return results
}
