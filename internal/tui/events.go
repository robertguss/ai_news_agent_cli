package tui

type ProgressMsg struct {
	Source  string
	Current int
	Total   int
	Status  string
	Error   error
}

type CompletedMsg struct {
	Source string
	Added  int
	Error  error
}

type FinalSummaryMsg struct {
	TotalAdded    int
	TotalSources  int
	SuccessCount  int
	ErrorCount    int
	Errors        []error
}
