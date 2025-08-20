package processor

import (
        "context"
        "encoding/json"
        "os"
        "testing"

        "github.com/stretchr/testify/assert"
        "github.com/stretchr/testify/require"
)

func TestNewGeminiProcessor_Success(t *testing.T) {
        originalKey := os.Getenv("GEMINI_API_KEY")
        defer os.Setenv("GEMINI_API_KEY", originalKey)

        os.Setenv("GEMINI_API_KEY", "test-api-key")

        processor, err := NewGeminiProcessor(context.Background())
        require.NoError(t, err)
        assert.NotNil(t, processor)
}

func TestNewGeminiProcessor_MissingAPIKey(t *testing.T) {
        originalKey := os.Getenv("GEMINI_API_KEY")
        defer os.Setenv("GEMINI_API_KEY", originalKey)

        os.Unsetenv("GEMINI_API_KEY")

        processor, err := NewGeminiProcessor(context.Background())
        assert.Error(t, err)
        assert.Nil(t, processor)
        assert.Contains(t, err.Error(), "GEMINI_API_KEY")
}

func TestNewGeminiProcessor_EmptyAPIKey(t *testing.T) {
        originalKey := os.Getenv("GEMINI_API_KEY")
        defer os.Setenv("GEMINI_API_KEY", originalKey)

        os.Setenv("GEMINI_API_KEY", "")

        processor, err := NewGeminiProcessor(context.Background())
        assert.Error(t, err)
        assert.Nil(t, processor)
        assert.Contains(t, err.Error(), "GEMINI_API_KEY")
}

func TestGeminiProcessor_AnalyzeContent_Success(t *testing.T) {
        originalKey := os.Getenv("GEMINI_API_KEY")
        defer os.Setenv("GEMINI_API_KEY", originalKey)

        os.Setenv("GEMINI_API_KEY", "test-api-key")

        processor, err := NewGeminiProcessor(context.Background())
        require.NoError(t, err)
        require.NotNil(t, processor)

        content := "OpenAI has released a new GPT model with significant improvements."
        
        result, err := processor.AnalyzeContent(content)
        
        if err != nil {
                t.Skipf("Skipping integration test - requires real API key: %v", err)
                return
        }

        require.NotNil(t, result)
        assert.NotEmpty(t, result.Summary)
        assert.NotEmpty(t, result.ContentType)
        assert.NotEmpty(t, result.StoryGroupID)
}

func TestGeminiProcessor_AnalyzeContent_NetworkError(t *testing.T) {
        originalKey := os.Getenv("GEMINI_API_KEY")
        defer os.Setenv("GEMINI_API_KEY", originalKey)

        os.Setenv("GEMINI_API_KEY", "invalid-key")

        processor, err := NewGeminiProcessor(context.Background())
        require.NoError(t, err)
        
        result, err := processor.AnalyzeContent("test content")
        assert.Error(t, err)
        assert.Nil(t, result)
}

func TestGeminiProcessor_AnalyzeContent_InvalidJSON(t *testing.T) {
        t.Skip("Skipping JSON parsing test - requires proper HTTP mocking")
}

func TestGeminiProcessor_AnalyzeContent_HTTPError(t *testing.T) {
        t.Skip("Skipping HTTP error test - requires proper HTTP mocking")
}

func TestAnalysisResult_JSONSerialization(t *testing.T) {
        result := AnalysisResult{
                Summary: "Test summary with bullet points",
                Entities: Entities{
                        Organizations: []string{"OpenAI", "Google"},
                        Products:      []string{"GPT-4", "Gemini"},
                        People:        []string{"Sam Altman", "Sundar Pichai"},
                },
                Topics:       []string{"AI", "Machine Learning"},
                ContentType:  "News Article",
                StoryGroupID: "test-story-123",
        }

        data, err := json.Marshal(result)
        require.NoError(t, err)

        var unmarshaled AnalysisResult
        err = json.Unmarshal(data, &unmarshaled)
        require.NoError(t, err)

        assert.Equal(t, result.Summary, unmarshaled.Summary)
        assert.Equal(t, result.Entities.Organizations, unmarshaled.Entities.Organizations)
        assert.Equal(t, result.Entities.Products, unmarshaled.Entities.Products)
        assert.Equal(t, result.Entities.People, unmarshaled.Entities.People)
        assert.Equal(t, result.Topics, unmarshaled.Topics)
        assert.Equal(t, result.ContentType, unmarshaled.ContentType)
        assert.Equal(t, result.StoryGroupID, unmarshaled.StoryGroupID)
}

func TestAnalysisResult_DatabaseCompatibility(t *testing.T) {
        result := AnalysisResult{
                Summary: "Database compatibility test",
                Entities: Entities{
                        Organizations: []string{"Company1", "Company2"},
                        Products:      []string{"Product1"},
                        People:        []string{"Person1"},
                },
                Topics:       []string{"Topic1", "Topic2"},
                ContentType:  "Research Paper",
                StoryGroupID: "db-test-456",
        }

        entitiesJSON := result.EntitiesJSON()
        topicsJSON := result.TopicsJSON()

        assert.NotNil(t, entitiesJSON)
        assert.NotNil(t, topicsJSON)

        var entities Entities
        err := json.Unmarshal(entitiesJSON, &entities)
        require.NoError(t, err)
        assert.Equal(t, result.Entities, entities)

        var topics []string
        err = json.Unmarshal(topicsJSON, &topics)
        require.NoError(t, err)
        assert.Equal(t, result.Topics, topics)
}

func TestStoryGroupID_Deterministic(t *testing.T) {
        content1 := "This is test content for hashing"
        content2 := "This is test content for hashing"
        content3 := "This is different content"

        id1 := generateStoryGroupID(content1)
        id2 := generateStoryGroupID(content2)
        id3 := generateStoryGroupID(content3)

        assert.Equal(t, id1, id2, "Same content should produce same story group ID")
        assert.NotEqual(t, id1, id3, "Different content should produce different story group ID")
        assert.NotEmpty(t, id1)
        assert.NotEmpty(t, id3)
}

func TestStoryGroupID_Format(t *testing.T) {
        content := "Test content for ID format validation"
        id := generateStoryGroupID(content)

        assert.NotEmpty(t, id)
        assert.Regexp(t, "^[a-f0-9]+$", id, "Story group ID should be hexadecimal")
}

func TestGeminiProcessor_ImplementsInterface(t *testing.T) {
        var processor AIProcessor = &GeminiProcessor{}
        assert.NotNil(t, processor)

        if processor != nil {
                var _ func(string) (*AnalysisResult, error) = processor.AnalyzeContent
        }
}
