package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleCompareKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	w := &m.workspace.compare
	switch msg.String() {
	case "esc":
		// Exit never destroys workflow state.
		w.Exit()
		m.workspace.mode = modeObserve
		m.workspace.view = TimelineView
		return m, nil
	case "x":
		// Clear ends the comparison but preserves the pinned baseline.
		w.Clear()
		m.workspace.mode = modeObserve
		m.inspectBodyOffset = 0
		return m, nil
	case "s":
		// Swap exchanges perspective only.
		w.Swap()
		m.inspectBodyOffset = 0
		return m, nil
	case "q", "ctrl+c":
		w.Clear()
		m.workspace.dialog = dialogConfirmQuit
		return m, nil
	case "[":
		w.PrevView()
		return m, nil
	case "]":
		w.NextView()
		return m, nil
	case "tab":
		w.NextView()
		return m, nil
	case "shift+tab":
		w.PrevView()
		return m, nil
	case "home":
		w.View = CompareViewOverview
		m.inspectBodyOffset = 0
		return m, nil
	case "end":
		w.View = CompareViewRaw
		m.inspectBodyOffset = 0
		return m, nil
	case "up", "k":
		if m.inspectBodyOffset > 0 {
			m.inspectBodyOffset--
		}
		return m, nil
	case "down", "j":
		m.inspectBodyOffset++
		return m, nil
	}
	return m, nil
}
