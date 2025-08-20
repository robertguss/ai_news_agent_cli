package viewui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ArticleItem struct {
	ID      int64
	Title   string
	Source  string
	Summary string
	URL     string
	IsRead  bool
}

type FilterMode int

const (
	FilterNone FilterMode = iota
	FilterSearch
	FilterSource
	FilterReadStatus
)

type ReadStatusFilter int

const (
	ShowAll ReadStatusFilter = iota
	ShowUnread
	ShowRead
)

type Model struct {
	articles         []ArticleItem
	filteredArticles []ArticleItem
	selectedIndex    int
	width            int
	height           int
	markReadFunc     func(int64) error
	toggleReadFunc   func(int64) error

	// Filtering
	filterMode       FilterMode
	searchInput      textinput.Model
	sourceFilter     string
	availableSources []string
	sourceIndex      int
	readStatusFilter ReadStatusFilter
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
	searchInput := textinput.New()
	searchInput.Placeholder = "Search articles..."
	searchInput.CharLimit = 50

	model := Model{
		articles:         articles,
		filteredArticles: articles,
		selectedIndex:    0,
		filterMode:       FilterNone,
		searchInput:      searchInput,
		readStatusFilter: ShowAll,
	}

	// Extract unique sources
	sourceMap := make(map[string]bool)
	for _, article := range articles {
		sourceMap[article.Source] = true
	}

	model.availableSources = []string{"All Sources"}
	for source := range sourceMap {
		model.availableSources = append(model.availableSources, source)
	}

	return model
}

func (m *Model) SetCallbacks(markRead, toggleRead func(int64) error) {
	m.markReadFunc = markRead
	m.toggleReadFunc = toggleRead
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		// Handle filter mode specific keys
		if m.filterMode == FilterSearch {
			switch msg.String() {
			case "esc":
				m.filterMode = FilterNone
				m.searchInput.Blur()
				m.applyFilters()
			case "enter":
				m.filterMode = FilterNone
				m.searchInput.Blur()
				m.applyFilters()
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.applyFilters()
				return m, cmd
			}
		} else {
			// Normal navigation and commands
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit

			case "up", "k":
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}

			case "down", "j":
				if m.selectedIndex < len(m.filteredArticles)-1 {
					m.selectedIndex++
				}

			case "enter":
				if len(m.filteredArticles) > 0 && m.markReadFunc != nil {
					article := m.getSelectedArticle()
					if article != nil && !article.IsRead {
						err := m.markReadFunc(article.ID)
						if err == nil {
							article.IsRead = true
							m.applyFilters()
						}
					}
				}

			case " ": // Space key
				if len(m.filteredArticles) > 0 {
					article := m.getSelectedArticle()
					if article != nil {
						if m.markReadFunc != nil && !article.IsRead {
							err := m.markReadFunc(article.ID)
							if err == nil {
								article.IsRead = true
								m.applyFilters()
							}
						}
						// Open in browser
						go openBrowser(article.URL)
					}
				}

			case "r":
				if len(m.filteredArticles) > 0 && m.toggleReadFunc != nil {
					article := m.getSelectedArticle()
					if article != nil {
						err := m.toggleReadFunc(article.ID)
						if err == nil {
							article.IsRead = !article.IsRead
							m.applyFilters()
						}
					}
				}

			case "/":
				m.filterMode = FilterSearch
				m.searchInput.Focus()
				return m, textinput.Blink

			case "s":
				m.cycleSourceFilter()
				m.applyFilters()

			case "f":
				m.cycleReadStatusFilter()
				m.applyFilters()

			case "esc":
				m.clearFilters()
				m.applyFilters()
			}
		}
	}
	return m, cmd
}

func (m Model) View() string {
	if len(m.articles) == 0 {
		return "No articles found.\n\nPress 'q' to quit"
	}

	// Show search input if in search mode
	if m.filterMode == FilterSearch {
		searchView := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1).
			Render("Search: " + m.searchInput.View())

		return searchView + "\n\nPress Enter to apply, Esc to cancel"
	}

	if len(m.filteredArticles) == 0 {
		return "No articles match current filters.\n\nPress 'esc' to clear filters, 'q' to quit"
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

	// Title with filter info
	title := fmt.Sprintf("Articles (%d/%d shown, %d unread)",
		len(m.filteredArticles), len(m.articles), m.countUnread())
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(title) + "\n")

	// Filter status
	filterInfo := m.getFilterInfo()
	if filterInfo != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF")).Render(filterInfo) + "\n")
	}
	b.WriteString("\n")

	for i, article := range m.filteredArticles {
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
	b.WriteString(helpStyle.Render("↑↓ navigate • Enter preview • Space open • R toggle"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("/ search • S source • F filter • ESC clear • Q quit"))

	return lipgloss.NewStyle().Width(width).Render(b.String())
}

func (m Model) renderPreview(width int) string {
	if m.selectedIndex >= len(m.filteredArticles) {
		return previewStyle.Width(width).Render("No article selected")
	}

	article := m.filteredArticles[m.selectedIndex]

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

func (m *Model) getSelectedArticle() *ArticleItem {
	if m.selectedIndex >= len(m.filteredArticles) {
		return nil
	}

	// Find the article in the original slice to modify it
	selectedArticle := &m.filteredArticles[m.selectedIndex]
	for i := range m.articles {
		if m.articles[i].ID == selectedArticle.ID {
			return &m.articles[i]
		}
	}
	return nil
}

func (m *Model) applyFilters() {
	m.filteredArticles = nil

	for _, article := range m.articles {
		if m.matchesFilters(article) {
			m.filteredArticles = append(m.filteredArticles, article)
		}
	}

	// Adjust selected index if needed
	if m.selectedIndex >= len(m.filteredArticles) {
		m.selectedIndex = len(m.filteredArticles) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

func (m Model) matchesFilters(article ArticleItem) bool {
	// Search filter
	searchTerm := strings.ToLower(m.searchInput.Value())
	if searchTerm != "" {
		titleMatch := strings.Contains(strings.ToLower(article.Title), searchTerm)
		summaryMatch := strings.Contains(strings.ToLower(article.Summary), searchTerm)
		if !titleMatch && !summaryMatch {
			return false
		}
	}

	// Source filter
	if m.sourceFilter != "" && m.sourceFilter != "All Sources" {
		if article.Source != m.sourceFilter {
			return false
		}
	}

	// Read status filter
	switch m.readStatusFilter {
	case ShowUnread:
		if article.IsRead {
			return false
		}
	case ShowRead:
		if !article.IsRead {
			return false
		}
	}

	return true
}

func (m *Model) cycleSourceFilter() {
	m.sourceIndex = (m.sourceIndex + 1) % len(m.availableSources)
	m.sourceFilter = m.availableSources[m.sourceIndex]
}

func (m *Model) cycleReadStatusFilter() {
	switch m.readStatusFilter {
	case ShowAll:
		m.readStatusFilter = ShowUnread
	case ShowUnread:
		m.readStatusFilter = ShowRead
	case ShowRead:
		m.readStatusFilter = ShowAll
	}
}

func (m *Model) clearFilters() {
	m.searchInput.SetValue("")
	m.sourceFilter = ""
	m.sourceIndex = 0
	m.readStatusFilter = ShowAll
}

func (m Model) getFilterInfo() string {
	var filters []string

	if m.searchInput.Value() != "" {
		filters = append(filters, fmt.Sprintf("Search: %s", m.searchInput.Value()))
	}

	if m.sourceFilter != "" && m.sourceFilter != "All Sources" {
		filters = append(filters, fmt.Sprintf("Source: %s", m.sourceFilter))
	}

	switch m.readStatusFilter {
	case ShowUnread:
		filters = append(filters, "Status: Unread only")
	case ShowRead:
		filters = append(filters, "Status: Read only")
	}

	if len(filters) == 0 {
		return ""
	}

	return "Filters: " + strings.Join(filters, " • ")
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
