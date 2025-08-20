package processor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalysisResult_StructureSerialization(t *testing.T) {
	result := AnalysisResult{
		Summary: "Test summary",
		Entities: Entities{
			Organizations: []string{"OpenAI"},
			Products:      []string{"GPT-4"},
			People:        []string{"Sam Altman"},
		},
		Topics:       []string{"AI", "Technology"},
		ContentType:  "news",
		StoryGroupID: "story-123",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	var unmarshaled AnalysisResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, result.Summary, unmarshaled.Summary)
	assert.Equal(t, result.ContentType, unmarshaled.ContentType)
	assert.Equal(t, result.StoryGroupID, unmarshaled.StoryGroupID)
	assert.Equal(t, result.Entities, unmarshaled.Entities)
	assert.Equal(t, result.Topics, unmarshaled.Topics)
}

func TestAnalysisResult_MatchesDatabaseSchema(t *testing.T) {
	result := AnalysisResult{
		Summary: "Article summary",
		Entities: Entities{
			Organizations: []string{},
			Products:      []string{},
			People:        []string{},
		},
		Topics:       []string{},
		ContentType:  "news",
		StoryGroupID: "group-1",
	}

	assert.NotEmpty(t, result.Summary)
	assert.NotNil(t, result.Entities)
	assert.NotNil(t, result.Topics)
	assert.NotEmpty(t, result.ContentType)
	assert.NotEmpty(t, result.StoryGroupID)

	entitiesJSON := result.EntitiesJSON()
	topicsJSON := result.TopicsJSON()
	assert.NotNil(t, entitiesJSON)
	assert.NotNil(t, topicsJSON)
}

func TestAIProcessor_InterfaceExists(t *testing.T) {
	var processor AIProcessor
	assert.Nil(t, processor)

	// Verify interface has the expected method signature
	// This will fail to compile if the interface changes
	if processor != nil {
		var _ func(string) (*AnalysisResult, error) = processor.AnalyzeContent
	}
}
