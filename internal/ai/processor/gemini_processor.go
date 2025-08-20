package processor

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProcessor struct {
	client *genai.Client
}

func NewGeminiProcessor(ctx context.Context) (*GeminiProcessor, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProcessor{
		client: client,
	}, nil
}

func (gp *GeminiProcessor) AnalyzeContent(content string) (*AnalysisResult, error) {
	ctx := context.Background()
	
	model := gp.client.GenerativeModel("gemini-pro")
	
	prompt := fmt.Sprintf(`Analyze this AI news article and return a JSON response with the following structure:
{
  "summary": "• Bullet point summary\n• Key points\n• Important details",
  "entities": {
    "organizations": ["Company1", "Company2"],
    "products": ["Product1", "Model1"],
    "people": ["Person1", "Person2"]
  },
  "topics": ["Topic1", "Topic2"],
  "content_type": "Research Paper|Product Launch|News Article|Opinion Piece|Tutorial",
  "story_group_id": "unique-identifier-for-deduplication"
}

Article content:
%s`, content)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no response from Gemini API")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	var geminiResponse struct {
		Summary      string   `json:"summary"`
		Entities     Entities `json:"entities"`
		Topics       []string `json:"topics"`
		ContentType  string   `json:"content_type"`
		StoryGroupID string   `json:"story_group_id"`
	}

	if err := json.Unmarshal([]byte(responseText), &geminiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	return &AnalysisResult{
		Summary:      geminiResponse.Summary,
		Entities:     geminiResponse.Entities,
		Topics:       geminiResponse.Topics,
		ContentType:  geminiResponse.ContentType,
		StoryGroupID: generateStoryGroupID(content),
	}, nil
}

func generateStoryGroupID(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)[:16]
}
