package fetcher

import (
        "context"
        "database/sql"
        "time"

        "github.com/mmcdole/gofeed"
        "github.com/robertguss/ai-news-agent-cli/internal/ai/processor"
        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/robertguss/ai-news-agent-cli/internal/scraper"
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
                        
                        if deps.Scraper != nil && deps.AI != nil {
                                content, scrapeErr := deps.Scraper.Scrape(article.Link)
                                if scrapeErr == nil {
                                        result, aiErr := deps.AI.AnalyzeContent(content)
                                        if aiErr == nil && result != nil {
                                                summary = sql.NullString{
                                                        String: result.Summary,
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
