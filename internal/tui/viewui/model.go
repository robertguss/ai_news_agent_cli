package viewui

import (
        "fmt"
        "os/exec"
        "runtime"
        "strings"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/charmbracelet/lipgloss"
)

type ArticleItem struct {
        ID       int64
        Title    string
        Source   string
        Summary  string
        URL      string
        IsRead   bool
}

type Model struct {
        articles      []ArticleItem
        selectedIndex int
        width         int
        height        int
        markReadFunc  func(int64) error
        toggleReadFunc func(int64) error
}

var (
        selectedStyle = lipgloss.NewStyle().
                        Background(lipgloss.Color("#874BFD")).
                        Foreground(lipgloss.Color("#FFFFFF")).
                        Bold(true)
        
        unreadStyle = lipgloss.NewStyle().
                      Foreground(lipgloss.Color("#00FF87")).
                      Bold(true)
        
        readStyle = lipgloss.NewStyle().
                    Foreground(lipgloss.Color("#7D7D7D"))
        
        previewStyle = lipgloss.NewStyle().
                       Border(lipgloss.RoundedBorder()).
                       BorderForeground(lipgloss.Color("#874BFD")).
                       Padding(1).
                       Height(10)
        
        helpStyle = lipgloss.NewStyle().
                    Foreground(lipgloss.Color("#7D7D7D")).
                    Italic(true)
)

func New(articles []ArticleItem) Model {
        return Model{
                articles:      articles,
                selectedIndex: 0,
        }
}

func (m *Model) SetCallbacks(markRead, toggleRead func(int64) error) {
        m.markReadFunc = markRead
        m.toggleReadFunc = toggleRead
}

func (m Model) Init() tea.Cmd {
        return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
        case tea.WindowSizeMsg:
                m.width = msg.Width
                m.height = msg.Height
                
        case tea.KeyMsg:
                switch msg.String() {
                case "q", "ctrl+c":
                        return m, tea.Quit
                        
                case "up", "k":
                        if m.selectedIndex > 0 {
                                m.selectedIndex--
                        }
                        
                case "down", "j":
                        if m.selectedIndex < len(m.articles)-1 {
                                m.selectedIndex++
                        }
                        
                case "enter":
                        if len(m.articles) > 0 && m.markReadFunc != nil {
                                article := &m.articles[m.selectedIndex]
                                if !article.IsRead {
                                        err := m.markReadFunc(article.ID)
                                        if err == nil {
                                                article.IsRead = true
                                        }
                                }
                        }
                        
                case " ": // Space key
                        if len(m.articles) > 0 {
                                article := &m.articles[m.selectedIndex]
                                if m.markReadFunc != nil && !article.IsRead {
                                        err := m.markReadFunc(article.ID)
                                        if err == nil {
                                                article.IsRead = true
                                        }
                                }
                                // Open in browser
                                go openBrowser(article.URL)
                        }
                        
                case "r":
                        if len(m.articles) > 0 && m.toggleReadFunc != nil {
                                article := &m.articles[m.selectedIndex]
                                err := m.toggleReadFunc(article.ID)
                                if err == nil {
                                        article.IsRead = !article.IsRead
                                }
                        }
                }
        }
        return m, nil
}

func (m Model) View() string {
        if len(m.articles) == 0 {
                return "No articles found.\n\nPress 'q' to quit"
        }
        
        // Calculate layout dimensions (50/50 split)
        listWidth := m.width / 2
        previewWidth := m.width - listWidth - 2 // Account for border
        
        if listWidth < 20 {
                listWidth = 20
        }
        if previewWidth < 20 {
                previewWidth = 20
        }
        
        // Build article list (left pane)
        listPane := m.renderArticleList(listWidth)
        
        // Build preview pane (right pane)
        previewPane := m.renderPreview(previewWidth)
        
        // Combine panes side by side
        return lipgloss.JoinHorizontal(lipgloss.Top, listPane, previewPane)
}

func (m Model) renderArticleList(width int) string {
        var b strings.Builder
        
        title := fmt.Sprintf("Articles (%d total, %d unread)", len(m.articles), m.countUnread())
        b.WriteString(lipgloss.NewStyle().Bold(true).Render(title) + "\n\n")
        
        for i, article := range m.articles {
                var style lipgloss.Style
                prefix := "  "
                
                if i == m.selectedIndex {
                        style = selectedStyle
                        prefix = "> "
                } else if article.IsRead {
                        style = readStyle
                } else {
                        style = unreadStyle
                        prefix = "● "
                }
                
                status := ""
                if article.IsRead {
                        status = "[READ] "
                } else {
                        status = "[NEW] "
                }
                
                line := fmt.Sprintf("%s%s%s", prefix, status, article.Title)
                if len(line) > width-4 {
                        line = line[:width-7] + "..."
                }
                
                b.WriteString(style.Render(line) + "\n")
                
                // Add source info for selected item
                if i == m.selectedIndex {
                        sourceLine := fmt.Sprintf("    Source: %s", article.Source)
                        if len(sourceLine) > width-4 {
                                sourceLine = sourceLine[:width-7] + "..."
                        }
                        b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#7D7D7D")).Render(sourceLine) + "\n")
                }
        }
        
        b.WriteString("\n")
        b.WriteString(helpStyle.Render("↑↓ navigate • Enter preview • Space open • R toggle • Q quit"))
        
        return lipgloss.NewStyle().Width(width).Render(b.String())
}

func (m Model) renderPreview(width int) string {
        if m.selectedIndex >= len(m.articles) {
                return previewStyle.Width(width).Render("No article selected")
        }
        
        article := m.articles[m.selectedIndex]
        
        var b strings.Builder
        
        // Title
        b.WriteString(lipgloss.NewStyle().Bold(true).Render(article.Title) + "\n\n")
        
        // Source and status
        status := "Unread"
        if article.IsRead {
                status = "Read"
        }
        b.WriteString(fmt.Sprintf("Source: %s • Status: %s\n\n", article.Source, status))
        
        // Summary
        if article.Summary != "" {
                b.WriteString("Summary:\n")
                // Word wrap the summary
                wrapped := wordWrap(article.Summary, width-4)
                b.WriteString(wrapped + "\n\n")
        }
        
        // URL
        if article.URL != "" {
                b.WriteString("URL: " + article.URL + "\n\n")
        }
        
        // Help text
        b.WriteString(helpStyle.Render("Enter to mark as read • Space to open in browser"))
        
        return previewStyle.Width(width).Render(b.String())
}

func (m Model) countUnread() int {
        count := 0
        for _, article := range m.articles {
                if !article.IsRead {
                        count++
                }
        }
        return count
}

func wordWrap(text string, width int) string {
        if len(text) <= width {
                return text
        }
        
        words := strings.Fields(text)
        if len(words) == 0 {
                return text
        }
        
        var lines []string
        var currentLine strings.Builder
        
        for _, word := range words {
                if currentLine.Len() == 0 {
                        currentLine.WriteString(word)
                } else if currentLine.Len()+1+len(word) <= width {
                        currentLine.WriteString(" " + word)
                } else {
                        lines = append(lines, currentLine.String())
                        currentLine.Reset()
                        currentLine.WriteString(word)
                }
        }
        
        if currentLine.Len() > 0 {
                lines = append(lines, currentLine.String())
        }
        
        return strings.Join(lines, "\n")
}

func openBrowser(url string) error {
        if url == "" {
                return fmt.Errorf("no URL provided")
        }
        
        var cmd *exec.Cmd
        
        switch runtime.GOOS {
        case "linux":
                cmd = exec.Command("xdg-open", url)
        case "windows":
                cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
        case "darwin":
                cmd = exec.Command("open", url)
        default:
                return fmt.Errorf("unsupported platform")
        }
        
        return cmd.Start()
}
