package tui

type Phase string

const (
	PhaseRSSFetch Phase = "rss_fetch"
	PhaseScrape   Phase = "scrape"
	PhaseAI       Phase = "ai"
	PhaseDone     Phase = "done"
)

type DetailedProgressMsg struct {
	Source       string
	Phase        Phase
	Current      int
	Total        int
	ArticleTitle string
	Error        error
}
