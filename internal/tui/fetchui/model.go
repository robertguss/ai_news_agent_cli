package fetchui

import (
        "fmt"
        "runtime"
        "strings"

        "github.com/charmbracelet/bubbles/progress"
        "github.com/charmbracelet/bubbles/spinner"
        tea "github.com/charmbracelet/bubbletea"
        "github.com/robertguss/ai-news-agent-cli/internal/tui"
)

type SourceProgress struct {
        Name         string
        Current      int
        Total        int
        Status       string
        Phase        tui.Phase
        ArticleTitle string
        Error        error
        Progress     progress.Model
        Complete     bool
}

type Model struct {
        sources      map[string]*SourceProgress
        sourceOrder  []string
        spinner      spinner.Model
        totalAdded   int
        totalSources int
        successCount int
        errorCount   int
        errors       []error
        showErrors   bool
        complete     bool
        width        int
        height       int
        workerCount  int
}

func New(sourceNames []string) Model {
        s := spinner.New()
        s.Spinner = spinner.Dot
        s.Style = tui.SpinnerStyle

        sources := make(map[string]*SourceProgress)
        for _, name := range sourceNames {
                p := progress.New(progress.WithDefaultGradient())
                p.Width = 40
                sources[name] = &SourceProgress{
                        Name:     name,
                        Progress: p,
                }
        }

        return Model{
                sources:      sources,
                sourceOrder:  sourceNames,
                spinner:      s,
                workerCount:  runtime.NumCPU(),
                totalSources: len(sourceNames),
        }
}

func (m *Model) SetWorkerCount(count int) {
        m.workerCount = count
}

func (m Model) Init() tea.Cmd {
        return tea.Batch(m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        var cmds []tea.Cmd

        switch msg := msg.(type) {
        case tea.WindowSizeMsg:
                m.width = msg.Width
                m.height = msg.Height
                for _, source := range m.sources {
                        source.Progress.Width = min(40, m.width-20)
                }

        case tea.KeyMsg:
                switch msg.String() {
                case "q", "ctrl+c":
                        return m, tea.Quit
                case "e":
                        m.showErrors = !m.showErrors
                }

        case spinner.TickMsg:
                var cmd tea.Cmd
                m.spinner, cmd = m.spinner.Update(msg)
                cmds = append(cmds, cmd)

        case tui.ProgressMsg:
                if source, exists := m.sources[msg.Source]; exists {
                        source.Current = msg.Current
                        source.Total = msg.Total
                        source.Status = msg.Status
                        source.Error = msg.Error

                        if msg.Error != nil {
                                source.Complete = true
                        }
                }

        case tui.ArticleProgressMsg:
                if source, exists := m.sources[msg.Source]; exists {
                        source.Current = msg.Current
                        source.Total = msg.Total
                        source.Phase = msg.Phase
                        source.ArticleTitle = msg.ArticleTitle
                        source.Error = msg.Error

                        switch msg.Phase {
                        case tui.PhaseRSSFetch:
                                source.Status = "Fetching RSS..."
                        case tui.PhaseScrape:
                                if msg.ArticleTitle != "" {
                                        source.Status = fmt.Sprintf("Scraping: %s", msg.ArticleTitle)
                                } else {
                                        source.Status = "Scraping content..."
                                }
                        case tui.PhaseAI:
                                if msg.ArticleTitle != "" {
                                        source.Status = fmt.Sprintf("AI analyzing: %s", msg.ArticleTitle)
                                } else {
                                        source.Status = "AI analyzing..."
                                }
                        case tui.PhaseDone:
                                source.Complete = true
                                source.Status = "Complete"
                        }

                        if msg.Error != nil && msg.Phase != tui.PhaseDone {
                                source.Status = fmt.Sprintf("Error in %s: %v", msg.Phase, msg.Error)
                        }
                }

        case tui.ArticleAddedMsg:
                m.totalAdded += msg.Count

        case tui.CompletedMsg:
                if source, exists := m.sources[msg.Source]; exists {
                        source.Complete = true
                        source.Error = msg.Error

                        if msg.Error != nil {
                                m.errorCount++
                                m.errors = append(m.errors, fmt.Errorf("%s: %w", msg.Source, msg.Error))
                        } else {
                                m.successCount++
                        }
                }

        case tui.FinalSummaryMsg:
                m.complete = true
                m.totalAdded = msg.TotalAdded
                m.totalSources = msg.TotalSources
                m.successCount = msg.SuccessCount
                m.errorCount = msg.ErrorCount
                m.errors = msg.Errors
        }

        return m, tea.Batch(cmds...)
}

func (m Model) View() string {
        if m.complete {
                return m.renderComplete()
        }

        var b strings.Builder

        title := tui.TitleStyle.Render("Fetching Articles")
        b.WriteString(fmt.Sprintf("┌─ %s %s\n", title, strings.Repeat("─", max(0, m.width-len(title)-4))))
        b.WriteString("│\n")

        if !m.complete {
                status := fmt.Sprintf("%s Processing %d sources with %d workers...",
                        m.spinner.View(), m.totalSources, m.workerCount)
                b.WriteString(fmt.Sprintf("│ %s\n", status))
                b.WriteString("│\n")
        }

        for _, sourceName := range m.sourceOrder {
                source := m.sources[sourceName]
                b.WriteString(m.renderSource(source))
        }

        b.WriteString("│\n")

        if m.showErrors && len(m.errors) > 0 {
                b.WriteString(m.renderErrors())
        }

        progress := fmt.Sprintf("Progress: %d/%d sources • %d articles added",
                m.successCount+m.errorCount, m.totalSources, m.totalAdded)
        b.WriteString(fmt.Sprintf("│ %s\n", progress))

        help := "Press 'e' to toggle errors, 'q' to quit"
        b.WriteString(fmt.Sprintf("│ %s\n", tui.HelpStyle.Render(help)))

        b.WriteString(fmt.Sprintf("└%s┘", strings.Repeat("─", max(0, m.width-2))))

        return b.String()
}

func (m Model) renderSource(source *SourceProgress) string {
        var b strings.Builder

        var status string
        if source.Complete {
                if source.Error != nil {
                        status = tui.ErrorStyle.Render("⚠️")
                } else {
                        status = tui.SuccessStyle.Render("✓")
                }
        } else {
                status = "●"
        }

        sourceLine := fmt.Sprintf("│ %s %s", status, source.Name)

        if source.Total > 0 {
                progressBar := source.Progress.ViewAs(float64(source.Current) / float64(source.Total))
                progressText := fmt.Sprintf("%d/%d", source.Current, source.Total)
                sourceLine += fmt.Sprintf(" %s %s", progressBar, progressText)

                if source.Complete && source.Error == nil {
                        sourceLine += " " + tui.SuccessStyle.Render("✓")
                } else if source.Error != nil {
                        sourceLine += " " + tui.ErrorStyle.Render("⚠️")
                }
        }

        b.WriteString(sourceLine + "\n")

        if source.Status != "" && !source.Complete {
                statusLine := fmt.Sprintf("│   └─ %s", source.Status)
                b.WriteString(statusLine + "\n")
        }

        if source.Error != nil && source.Complete {
                errorLine := fmt.Sprintf("│   └─ %s", tui.ErrorStyle.Render(fmt.Sprintf("Error: %v", source.Error)))
                b.WriteString(errorLine + "\n")
        }

        return b.String()
}

func (m Model) renderErrors() string {
        var b strings.Builder

        b.WriteString("│ " + tui.ErrorStyle.Render("Errors:") + "\n")
        for _, err := range m.errors {
                b.WriteString(fmt.Sprintf("│   • %s\n", tui.ErrorStyle.Render(err.Error())))
        }
        b.WriteString("│\n")

        return b.String()
}

func (m Model) renderComplete() string {
        var b strings.Builder

        title := "Fetch Complete"
        if m.errorCount > 0 {
                title = tui.WarningStyle.Render(title)
        } else {
                title = tui.SuccessStyle.Render(title)
        }

        b.WriteString(fmt.Sprintf("┌─ %s %s\n", title, strings.Repeat("─", max(0, m.width-len(title)-4))))
        b.WriteString("│\n")

        summary := fmt.Sprintf("Added %d articles from %d sources", m.totalAdded, m.totalSources)
        b.WriteString(fmt.Sprintf("│ %s\n", tui.TitleStyle.Render(summary)))

        if m.successCount > 0 {
                success := fmt.Sprintf("✓ %d sources succeeded", m.successCount)
                b.WriteString(fmt.Sprintf("│ %s\n", tui.SuccessStyle.Render(success)))
        }

        if m.errorCount > 0 {
                errors := fmt.Sprintf("⚠️  %d sources failed", m.errorCount)
                b.WriteString(fmt.Sprintf("│ %s\n", tui.ErrorStyle.Render(errors)))

                if m.showErrors {
                        b.WriteString("│\n")
                        b.WriteString(m.renderErrors())
                } else {
                        b.WriteString(fmt.Sprintf("│ %s\n", tui.HelpStyle.Render("Press 'e' to view error details")))
                }
        }

        b.WriteString("│\n")
        b.WriteString(fmt.Sprintf("│ %s\n", tui.HelpStyle.Render("Press 'q' to quit")))
        b.WriteString(fmt.Sprintf("└%s┘", strings.Repeat("─", max(0, m.width-2))))

        return b.String()
}

func min(a, b int) int {
        if a < b {
                return a
        }
        return b
}

func max(a, b int) int {
        if a > b {
                return a
        }
        return b
}
