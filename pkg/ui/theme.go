package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Title lipgloss.Style
	Label lipgloss.Style
	Value lipgloss.Style
	Box   lipgloss.Style
	Error lipgloss.Style
	OK    lipgloss.Style
}

func DefaultTheme() Theme {
	return Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")),

		Label: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")),

		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		Box: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			MarginBottom(1).
			BorderForeground(lipgloss.Color("62")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),

		OK: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true),
	}
}
