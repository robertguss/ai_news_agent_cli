package fetcher

import (
        "context"
        "database/sql"
        "sync"
        "time"

        "github.com/mmcdole/gofeed"
        "github.com/robertguss/ai-news-agent-cli/internal/ai/processor"
        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/robertguss/ai-news-agent-cli/internal/scraper"
        "github.com/robertguss/ai-news-agent-cli/internal/tui"
)

type Source struct {
        Name     string `mapstructure:"name"`
        URL      string `mapstructure:"url"`
        Type     string `mapstructure:"type"`
        Priority int    `mapstructure:"priority"`
}

type Article struct {
        Title         string
        Link          string
        PublishedDate time.Time
}

type PipelineDeps struct {
        Scraper scraper.Scraper
        AI      processor.AIProcessor
        Queries *database.Queries
}

func Fetch(ctx context.Context, source Source) ([]Article, error) {
        parser := gofeed.NewParser()
        feed, err := parser.ParseURLWithContext(source.URL, ctx)
        if err != nil {
                return nil, err
        }

        var articles []Article
        for _, item := range feed.Items {
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

        return articles, nil
}

func StoreArticles(ctx context.Context, queries *database.Queries, articles []Article, source Source) (int, error) {
        stored := 0

        for _, article := range articles {
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
                                StoryGroupID: sql.NullString{
                                        String: "",
                                        Valid:  false,
                                },
                        }

                        _, err = queries.CreateArticle(ctx, params)
                        if err != nil {
                                return stored, err
                        }
                        stored++
                } else if err != nil {
                        return stored, err
                }
        }

        return stored, nil
}

func FetchAndStore(ctx context.Context, queries *database.Queries, source Source) (int, error) {
        articles, err := Fetch(ctx, source)
        if err != nil {
                return 0, err
        }

        return StoreArticles(ctx, queries, articles, source)
}

func FetchAndStoreWithAI(ctx context.Context, deps PipelineDeps, source Source) (int, error) {
        articles, err := Fetch(ctx, source)
        if err != nil {
                return 0, err
        }

        return StoreArticlesWithAI(ctx, deps, articles, source)
}

func StoreArticlesWithAI(ctx context.Context, deps PipelineDeps, articles []Article, source Source) (int, error) {
        stored := 0

        for _, article := range articles {
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
                        
                        if deps.Scraper != nil && deps.AI != nil {
                                content, scrapeErr := deps.Scraper.Scrape(article.Link)
                                if scrapeErr == nil {
                                        result, aiErr := deps.AI.AnalyzeContent(content)
                                        if aiErr == nil && result != nil {
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
                                Summary: summary,
                                Entities: entities,
                                ContentType: contentType,
                                Topics: topics,
                                Status: sql.NullString{
                                        String: "unread",
                                        Valid:  true,
                                },
                                StoryGroupID: storyGroupID,
                        }

                        _, err = deps.Queries.CreateArticle(ctx, params)
                        if err != nil {
                                return stored, err
                        }
                        stored++
                } else if err != nil {
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

func FetchAndStoreWithAIProgress(ctx context.Context, deps PipelineDeps, source Source, progress chan<- tui.DetailedProgressMsg) (int, error) {
        progress <- tui.DetailedProgressMsg{
                Source: source.Name,
                Phase:  tui.PhaseRSSFetch,
        }

        articles, err := Fetch(ctx, source)
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

                        if deps.Scraper != nil && deps.AI != nil {
                                content, scrapeErr := deps.Scraper.Scrape(article.Link)
                                if scrapeErr == nil {
                                        progress <- tui.DetailedProgressMsg{
                                                Source:       source.Name,
                                                Phase:        tui.PhaseAI,
                                                Current:      i + 1,
                                                Total:        total,
                                                ArticleTitle: article.Title,
                                        }

                                        result, aiErr := deps.AI.AnalyzeContent(content)
                                        if aiErr == nil && result != nil {
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
                                Summary:      summary,
                                Entities:     entities,
                                ContentType:  contentType,
                                Topics:       topics,
                                Status: sql.NullString{
                                        String: "unread",
                                        Valid:  true,
                                },
                                StoryGroupID: storyGroupID,
                        }

                        _, err = deps.Queries.CreateArticle(ctx, params)
                        if err == nil {
                                stored++
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

func ProcessSourcesConcurrently(ctx context.Context, sources []Source, workerCount int, processFunc func(context.Context, Source, chan<- tui.DetailedProgressMsg) (int, error), progress chan<- tui.DetailedProgressMsg) []SourceResult {
        results := make([]SourceResult, len(sources))
        sourceCh := make(chan int, len(sources))
        var wg sync.WaitGroup

        for i := 0; i < workerCount; i++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        for idx := range sourceCh {
                                source := sources[idx]
                                added, err := processFunc(ctx, source, progress)
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
