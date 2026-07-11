package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleInspectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.workspace.mode = modeObserve
		return m, nil
	case "tab":
		m.inspectZone = (m.inspectZone + 1) % 3
		return m, nil
	case "shift+tab":
		m.inspectZone = (m.inspectZone + 2) % 3
		return m, nil
	case "home":
		m.inspectZone = zoneWhatHappened
		return m, nil
	case "end":
		m.inspectZone = zoneBody
		return m, nil
	case "up", "k":
		if m.inspectZone == zoneBody && m.inspectBodyOffset > 0 {
			m.inspectBodyOffset--
		}
		return m, nil
	case "down", "j":
		if m.inspectZone == zoneBody {
			m.inspectBodyOffset++
		}
		return m, nil
	case "c":
		return m.handleMark()
	case "x":
		return m.handleRenounceOrClearKey()
	case "q", "ctrl+c":
		m.workspace.dialog = dialogConfirmQuit
		return m, nil
	}
	return m, nil
}
