package cmd

import (
        "context"
        "database/sql"
        "encoding/json"
        "fmt"
        "strconv"
        "strings"
        "time"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/robertguss/ai-news-agent-cli/internal/state"
        "github.com/robertguss/ai-news-agent-cli/internal/tui"
        "github.com/robertguss/ai-news-agent-cli/internal/tui/viewui"
        "github.com/spf13/cobra"
)

type ViewOptions struct {
        All    bool
        Source string
        Topic  string
}

var databaseOpen = database.Open
var shouldUseTUIFunc = tui.ShouldUseTUI
var runTUIViewFunc = runTUIView

var viewCmd = &cobra.Command{
        Use:   "view",
        Short: "List articles stored in the database with enhanced styling and filtering",
        RunE: func(cmd *cobra.Command, args []string) error {
                dbPath, _ := cmd.Flags().GetString("db")
                all, _ := cmd.Flags().GetBool("all")
                source, _ := cmd.Flags().GetString("source")
                topic, _ := cmd.Flags().GetString("topic")

                opts := ViewOptions{
                        All:    all,
                        Source: source,
                        Topic:  topic,
                }

                if shouldUseTUIFunc() {
                        return runTUIViewFunc(dbPath, opts)
                }

                return runLegacyView(cmd, dbPath, opts)
        },
}

func runTUIView(dbPath string, opts ViewOptions) error {
        db, q, err := databaseOpen(dbPath)
        if err != nil {
                return err
        }
        defer db.Close()

        err = database.InitSchema(db)
        if err != nil {
                return err
        }

        ctx := context.Background()
        articles, err := getFilteredArticles(ctx, q, opts.All, opts.Source, opts.Topic)
        if err != nil {
                return err
        }

        // Convert database articles to TUI articles
        var tuiArticles []viewui.ArticleItem
        for _, article := range articles {
                tuiArticles = append(tuiArticles, viewui.ArticleItem{
                        ID:      article.ID,
                        Title:   formatNullString(article.Title, "(no title)"),
                        Source:  formatNullString(article.SourceName, "(no source)"),
                        Summary: formatNullString(article.Summary, "No summary available"),
                        URL:     formatNullString(article.Url, ""),
                        IsRead:  article.Status.String == "read",
                })
        }

        model := viewui.New(tuiArticles)

        // Set up database callback functions
        model.SetCallbacks(
                func(id int64) error {
                        return q.MarkArticleAsRead(ctx, id)
                },
                func(id int64) error {
                        // Get current status
                        article, err := q.GetArticle(ctx, id)
                        if err != nil {
                                return err
                        }

                        newStatus := "read"
                        if article.Status.String == "read" {
                                newStatus = "unread"
                        }

                        return q.UpdateArticleStatus(ctx, database.UpdateArticleStatusParams{
                                ID:     id,
                                Status: sql.NullString{String: newStatus, Valid: true},
                        })
                },
        )

        p := tea.NewProgram(model, tea.WithAltScreen())

        _, err = p.Run()
        if err != nil {
                return err
        }

        if len(articles) > 0 {
                stateMap := make(map[string]state.ArticleRef)
                for i, article := range articles {
                        idx := strconv.Itoa(i + 1)
                        stateMap[idx] = state.ArticleRef{
                                ID:           article.ID,
                                URL:          formatNullString(article.Url, ""),
                                Title:        formatNullString(article.Title, ""),
                                StoryGroupID: formatNullString(article.StoryGroupID, ""),
                        }
                }

                viewState := &state.ViewState{
                        Timestamp: time.Now().UTC(),
                        Articles:  stateMap,
                }

                _ = state.Save(viewState)
        }

        return nil
}

func runLegacyView(cmd *cobra.Command, dbPath string, opts ViewOptions) error {
        db, q, err := databaseOpen(dbPath)
        if err != nil {
                return err
        }
        defer db.Close()

        err = database.InitSchema(db)
        if err != nil {
                return err
        }

        ctx := context.Background()
        articles, err := getFilteredArticles(ctx, q, opts.All, opts.Source, opts.Topic)
        if err != nil {
                return err
        }

        if len(articles) == 0 {
                fmt.Fprintln(cmd.OutOrStdout(), "No articles found.")
                return nil
        }

        groupedArticles := groupArticlesByStory(articles)
        var articleIDs []int64
        stateMap := make(map[string]state.ArticleRef)

        for i, group := range groupedArticles {
                primary := group[0]
                var duplicates []string

                for _, dup := range group[1:] {
                        duplicates = append(duplicates, formatNullString(dup.SourceName, "Unknown"))
                }

                title := formatNullString(primary.Title, "(no title)")
                sourceName := formatNullString(primary.SourceName, "(no source)")
                summary := formatNullString(primary.Summary, "")
                topics := formatTopics(primary.Topics)

                card := formatCard(i+1, title, sourceName, summary, topics, duplicates)
                fmt.Fprint(cmd.OutOrStdout(), card)

                idx := strconv.Itoa(i + 1)
                stateMap[idx] = state.ArticleRef{
                        ID:           primary.ID,
                        URL:          formatNullString(primary.Url, ""),
                        Title:        title,
                        StoryGroupID: formatNullString(primary.StoryGroupID, ""),
                }

                for _, article := range group {
                        articleIDs = append(articleIDs, article.ID)
                }
        }

        if !opts.All && len(articleIDs) > 0 {
                err = q.MarkArticlesAsRead(ctx, articleIDs)
                if err != nil {
                        return err
                }
        }

        if len(stateMap) > 0 {
                viewState := &state.ViewState{
                        Timestamp: time.Now().UTC(),
                        Articles:  stateMap,
                }

                _ = state.Save(viewState)
        }

        return nil
}

func formatNullString(ns sql.NullString, placeholder string) string {
        if ns.Valid && ns.String != "" {
                return ns.String
        }
        return placeholder
}

func formatTopics(topicsJSON interface{}) string {
        if topicsJSON == nil {
                return ""
        }

        topicsStr, ok := topicsJSON.(string)
        if !ok {
                return ""
        }

        if topicsStr == "" {
                return ""
        }

        var topics []string
        if err := json.Unmarshal([]byte(topicsStr), &topics); err != nil {
                return topicsStr
        }

        return strings.Join(topics, ", ")
}

func getFilteredArticles(ctx context.Context, q *database.Queries, all bool, source, topic string) ([]database.Article, error) {
        hasSource := source != ""
        hasTopic := topic != ""

        switch {
        case all && hasSource && hasTopic:
                return q.ListAllArticlesBySourceAndTopic(ctx, database.ListAllArticlesBySourceAndTopicParams{
                        SourceName: sql.NullString{String: source, Valid: true},
                        Column2:    sql.NullString{String: topic, Valid: true},
                })
        case all && hasSource:
                return q.ListAllArticlesBySource(ctx, sql.NullString{String: source, Valid: true})
        case all && hasTopic:
                return q.ListAllArticlesByTopic(ctx, sql.NullString{String: topic, Valid: true})
        case all:
                return q.ListAllArticles(ctx)
        case hasSource && hasTopic:
                return q.ListArticlesBySourceAndTopic(ctx, database.ListArticlesBySourceAndTopicParams{
                        SourceName: sql.NullString{String: source, Valid: true},
                        Column2:    sql.NullString{String: topic, Valid: true},
                })
        case hasSource:
                return q.ListArticlesBySource(ctx, sql.NullString{String: source, Valid: true})
        case hasTopic:
                return q.ListArticlesByTopic(ctx, sql.NullString{String: topic, Valid: true})
        default:
                return q.ListUnreadArticles(ctx)
        }
}

func groupArticlesByStory(articles []database.Article) [][]database.Article {
        storyGroups := make(map[string][]database.Article)
        var ungrouped []database.Article

        for _, article := range articles {
                if article.StoryGroupID.Valid && article.StoryGroupID.String != "" {
                        storyGroups[article.StoryGroupID.String] = append(storyGroups[article.StoryGroupID.String], article)
                } else {
                        ungrouped = append(ungrouped, article)
                }
        }

        var result [][]database.Article

        for _, group := range storyGroups {
                if len(group) > 0 {
                        result = append(result, group)
                }
        }

        for _, article := range ungrouped {
                result = append(result, []database.Article{article})
        }

        return result
}

func init() {
        viewCmd.Flags().StringP("db", "d", "ai-news.db", "Database file path")
        viewCmd.Flags().Bool("all", false, "Show all articles (read and unread) and don't mark as read")
        viewCmd.Flags().String("source", "", "Filter articles by source name")
        viewCmd.Flags().String("topic", "", "Filter articles by topic")
        rootCmd.AddCommand(viewCmd)
}
