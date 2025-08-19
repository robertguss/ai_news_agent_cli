package cmd

import (
        "context"
        "database/sql"
        "fmt"

        "github.com/robertguss/ai-news-agent-cli/internal/database"
        "github.com/spf13/cobra"
)

var databaseOpen = database.Open

var viewCmd = &cobra.Command{
        Use:   "view",
        Short: "List articles stored in the database",
        RunE: func(cmd *cobra.Command, args []string) error {
                dbPath, _ := cmd.Flags().GetString("db")
                
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
                articles, err := q.ListArticles(ctx)
                if err != nil {
                        return err
                }
                
                if len(articles) == 0 {
                        fmt.Fprintln(cmd.OutOrStdout(), "No articles found.")
                        return nil
                }
                
                for _, article := range articles {
                        title := formatNullString(article.Title, "(no title)")
                        source := formatNullString(article.SourceName, "(no source)")
                        fmt.Fprintf(cmd.OutOrStdout(), "%s - %s\n", title, source)
                }
                
                return nil
        },
}

func formatNullString(ns sql.NullString, placeholder string) string {
        if ns.Valid && ns.String != "" {
                return ns.String
        }
        return placeholder
}

func init() {
        viewCmd.Flags().StringP("db", "d", "news.db", "Database file path")
        rootCmd.AddCommand(viewCmd)
}
