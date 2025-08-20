package processor

import "encoding/json"

//go:generate mockery --name=AIProcessor

type AIProcessor interface {
        AnalyzeContent(content string) (*AnalysisResult, error)
}

type Entities struct {
        Organizations []string `json:"organizations,omitempty"`
        Products      []string `json:"products,omitempty"`
        People        []string `json:"people,omitempty"`
}

type AnalysisResult struct {
        Summary      string   `json:"summary"`
        Entities     Entities `json:"entities"`
        Topics       []string `json:"topics"`
        ContentType  string   `json:"content_type"`
        StoryGroupID string   `json:"story_group_id"`
}

func (a *AnalysisResult) EntitiesJSON() []byte {
        data, _ := json.Marshal(a.Entities)
        return data
}

func (a *AnalysisResult) TopicsJSON() []byte {
        data, _ := json.Marshal(a.Topics)
        return data
}
