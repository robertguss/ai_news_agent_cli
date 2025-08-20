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

type ArticleProgressMsg struct {
        Source       string
        Phase        Phase
        Current      int
        Total        int
        ArticleTitle string
        Error        error
}
