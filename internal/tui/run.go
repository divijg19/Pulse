package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func Run() error {
	program := tea.NewProgram(NewModel())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run tui: %w", err)
	}
	return nil
}
