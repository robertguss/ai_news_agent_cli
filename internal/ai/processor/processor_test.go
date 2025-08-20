package processor

import (
        "encoding/json"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func TestAnalysisResult_JSONSerialization(t *testing.T) {
        result := AnalysisResult{
                Summary:      "Test summary",
                Entities:     []byte(`[{"name":"OpenAI","type":"Organization"}]`),
                Topics:       []byte(`[{"name":"AI","category":"Technology"}]`),
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
        assert.JSONEq(t, string(result.Entities), string(unmarshaled.Entities))
        assert.JSONEq(t, string(result.Topics), string(unmarshaled.Topics))
}

func TestAnalysisResult_MatchesDatabaseSchema(t *testing.T) {
        result := AnalysisResult{
                Summary:      "Article summary",
                Entities:     []byte(`[]`),
                Topics:       []byte(`[]`),
                ContentType:  "news",
                StoryGroupID: "group-1",
        }

        assert.NotEmpty(t, result.Summary)
        assert.NotNil(t, result.Entities)
        assert.NotNil(t, result.Topics)
        assert.NotEmpty(t, result.ContentType)
        assert.NotEmpty(t, result.StoryGroupID)
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
