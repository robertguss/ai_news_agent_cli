package viewui

import (
        "testing"

        tea "github.com/charmbracelet/bubbletea"
        "github.com/stretchr/testify/assert"
)

func TestModel_InitialState(t *testing.T) {
        articles := []ArticleItem{
                {ID: 1, Title: "Test Article 1", Source: "Test Source 1"},
                {ID: 2, Title: "Test Article 2", Source: "Test Source 2"},
        }
        
        model := New(articles)
        
        assert.Equal(t, 2, len(model.articles))
        assert.Equal(t, 0, model.selectedIndex)
        assert.Equal(t, "Test Article 1", model.articles[0].Title)
}

func TestModel_QuitOnQKey(t *testing.T) {
        articles := []ArticleItem{
                {ID: 1, Title: "Test Article", Source: "Test Source"},
        }
        
        model := New(articles)
        
        quitMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
        updatedModel, cmd := model.Update(quitMsg)
        
        assert.NotNil(t, cmd)
        
        // Execute the command to see if it's a quit command
        msg := cmd()
        _, isQuit := msg.(tea.QuitMsg)
        assert.True(t, isQuit, "Should return quit command when 'q' is pressed")
        
        // Model should remain unchanged
        assert.Equal(t, model.selectedIndex, updatedModel.(Model).selectedIndex)
}

func TestModel_ShowsArticleTitlesInView(t *testing.T) {
        articles := []ArticleItem{
                {ID: 1, Title: "First Article", Source: "Source 1"},
                {ID: 2, Title: "Second Article", Source: "Source 2"},
        }
        
        model := New(articles)
        view := model.View()
        
        assert.Contains(t, view, "First Article")
        assert.Contains(t, view, "Second Article")
        assert.Contains(t, view, "Source 1")
        assert.Contains(t, view, "Source 2")
}
