package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	sourceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D7D7D"))

	summaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			MarginLeft(2)

	topicsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Italic(true)

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			MarginBottom(1)

	tierStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF87"))

	duplicatesStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			MarginLeft(2)
)

func getSourceTier(sourceName string) int {
	tierMap := map[string]int{
		"Google AI Blog":   1,
		"OpenAI Blog":      1,
		"Ars Technica AI": 2,
		"The Verge":       2,
		"Ars Technica":    2,
	}
	
	if tier, exists := tierMap[sourceName]; exists {
		return tier
	}
	return 3
}

func formatCard(index int, title, sourceName, summary, topics string, duplicates []string) string {
	tier := getSourceTier(sourceName)
	
	var cardContent strings.Builder
	
	cardContent.WriteString(fmt.Sprintf("[%d] %s\n", index, titleStyle.Render(title)))
	
	sourceInfo := fmt.Sprintf("Source: %s (Tier %d)", sourceName, tier)
	cardContent.WriteString(sourceStyle.Render(sourceInfo))
	cardContent.WriteString("\n")
	
	if summary != "" {
		cardContent.WriteString("Summary:\n")
		bulletPoints := strings.Split(summary, ". ")
		for _, point := range bulletPoints {
			if strings.TrimSpace(point) != "" {
				cardContent.WriteString(summaryStyle.Render(fmt.Sprintf("• %s\n", strings.TrimSpace(point))))
			}
		}
	}
	
	if topics != "" {
		cardContent.WriteString(fmt.Sprintf("Topics: %s\n", topicsStyle.Render(topics)))
	}
	
	if len(duplicates) > 0 {
		cardContent.WriteString(duplicatesStyle.Render(fmt.Sprintf("Also covered by: %s\n", strings.Join(duplicates, ", "))))
	}
	
	return cardStyle.Render(cardContent.String())
}
