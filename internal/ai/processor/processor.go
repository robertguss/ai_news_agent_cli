package processor

import (
	"context"
	"encoding/json"

	"github.com/robertguss/rss-agent-cli/internal/config"
)

//go:generate mockery --name=AIProcessor

// AIProcessor defines the interface for AI-powered content analysis and processing.
type AIProcessor interface {
	AnalyzeContent(content string) (*AnalysisResult, error)
	AnalyzeContentWithRetry(ctx context.Context, content string, cfg *config.Config) (*AnalysisResult, error)
}

// Entities represents extracted entities from content analysis.
type Entities struct {
	Organizations []string `json:"organizations,omitempty"`
	Products      []string `json:"products,omitempty"`
	People        []string `json:"people,omitempty"`
}

// AnalysisResult contains the complete AI analysis output for an article.
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
