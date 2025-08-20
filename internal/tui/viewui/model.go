package viewui

import (
        "fmt"
        "strings"

        tea "github.com/charmbracelet/bubbletea"
)

type ArticleItem struct {
        ID     int64
        Title  string
        Source string
}

type Model struct {
        articles      []ArticleItem
        selectedIndex int
}

func New(articles []ArticleItem) Model {
        return Model{
                articles:      articles,
                selectedIndex: 0,
        }
}

func (m Model) Init() tea.Cmd {
        return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
        case tea.KeyMsg:
                switch msg.String() {
                case "q", "ctrl+c":
                        return m, tea.Quit
                }
        }
        return m, nil
}

func (m Model) View() string {
        var b strings.Builder
        
        b.WriteString("Articles:\n\n")
        
        for i, article := range m.articles {
                prefix := "  "
                if i == m.selectedIndex {
                        prefix = "> "
                }
                
                b.WriteString(fmt.Sprintf("%s%s - %s\n", prefix, article.Title, article.Source))
        }
        
        b.WriteString("\nPress 'q' to quit")
        
        return b.String()
}
