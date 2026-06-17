package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func Run() error {
	program := tea.NewProgram(NewModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}
	return nil
}
