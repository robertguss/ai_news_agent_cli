package tui

import (
        "github.com/charmbracelet/lipgloss"
)

var (
        PrimaryColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#874BFD"}
        AccentColor    = lipgloss.AdaptiveColor{Light: "#00D7FF", Dark: "#00D7FF"}
        SuccessColor   = lipgloss.AdaptiveColor{Light: "#00FF87", Dark: "#00FF87"}
        ErrorColor     = lipgloss.AdaptiveColor{Light: "#FF5F87", Dark: "#FF5F87"}
        WarningColor   = lipgloss.AdaptiveColor{Light: "#FFA500", Dark: "#FFA500"}
        TextColor      = lipgloss.AdaptiveColor{Light: "#1C1C1C", Dark: "#FFFFFF"}
        SubtleColor    = lipgloss.AdaptiveColor{Light: "#7D7D7D", Dark: "#7D7D7D"}
        BackgroundColor = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1C1C1C"}
)

var (
        TitleStyle = lipgloss.NewStyle().
                        Bold(true).
                        Foreground(TextColor)

        SourceStyle = lipgloss.NewStyle().
                        Foreground(SubtleColor)

        SummaryStyle = lipgloss.NewStyle().
                        Foreground(TextColor).
                        MarginLeft(2)

        TopicsStyle = lipgloss.NewStyle().
                        Foreground(AccentColor).
                        Italic(true)

        CardStyle = lipgloss.NewStyle().
                        Border(lipgloss.RoundedBorder()).
                        BorderForeground(PrimaryColor).
                        Padding(1, 2).
                        MarginBottom(1)

        TierStyle = lipgloss.NewStyle().
                        Bold(true).
                        Foreground(SuccessColor)

        DuplicatesStyle = lipgloss.NewStyle().
                        Foreground(WarningColor).
                        MarginLeft(2)

        ProgressBarStyle = lipgloss.NewStyle().
                                Border(lipgloss.RoundedBorder()).
                                BorderForeground(PrimaryColor).
                                Padding(0, 1)

        ErrorStyle = lipgloss.NewStyle().
                        Foreground(ErrorColor).
                        Bold(true)

        SuccessStyle = lipgloss.NewStyle().
                        Foreground(SuccessColor).
                        Bold(true)

        SpinnerStyle = lipgloss.NewStyle().
                        Foreground(AccentColor)

        KeyStyle = lipgloss.NewStyle().
                        Foreground(AccentColor).
                        Bold(true)

        HelpStyle = lipgloss.NewStyle().
                        Foreground(SubtleColor)

        WarningStyle = lipgloss.NewStyle().
                        Foreground(WarningColor).
                        Bold(true)
)
