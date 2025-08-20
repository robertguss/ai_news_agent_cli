package processor

//go:generate mockery --name=AIProcessor

type AIProcessor interface {
        AnalyzeContent(content string) (*AnalysisResult, error)
}

type AnalysisResult struct {
        Summary      string `json:"summary"`
        Entities     []byte `json:"entities"`
        Topics       []byte `json:"topics"`
        ContentType  string `json:"content_type"`
        StoryGroupID string `json:"story_group_id"`
}
