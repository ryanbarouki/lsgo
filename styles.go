package main

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	renamingStyle    lipgloss.Style
	highlightedStyle lipgloss.Style
	cursorStyle      lipgloss.Style
	checkedStyle     lipgloss.Style
	confirmDelStyle  lipgloss.Style
}

func InitStyles() Styles {
	return Styles{
		renamingStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")),
		highlightedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#d8b172")).Bold(true),
		checkedStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")),
		cursorStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#d8b172")),
		confirmDelStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff0000")),
	}
}
