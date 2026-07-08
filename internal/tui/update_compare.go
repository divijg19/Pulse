package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleCompareKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.workspace.mode = modeObserve
		m.workspace.view = TimelineView
		return m, nil
	case "x":
		m.workspace.compare.Session = ComparisonSession{BaselineIndex: -1, CandidateIndex: -1, State: SessionIdle, Analysis: nil}
		m.workspace.compare.PinnedBaseline = nil
		m.workspace.mode = modeObserve
		m.status = "Comparison cleared"
		return m, nil
	case "s":
		m.workspace.compare.Session.BaselineIndex, m.workspace.compare.Session.CandidateIndex =
			m.workspace.compare.Session.CandidateIndex, m.workspace.compare.Session.BaselineIndex
		m.workspace.compare.Session.Analysis = m.computeComparisonAnalysis()
		m.inspectZone = zoneWhatHappened
		m.inspectBodyOffset = 0
		return m, nil
	case "q", "ctrl+c":
		m.workspace.compare.Session = ComparisonSession{BaselineIndex: -1, CandidateIndex: -1, State: SessionIdle, Analysis: nil}
		m.workspace.compare.PinnedBaseline = nil
		m.workspace.dialog = dialogConfirmQuit
		return m, nil
	case "tab":
		m.inspectZone = (m.inspectZone + 1) % 3
		m.inspectBodyOffset = 0
		return m, nil
	case "shift+tab":
		m.inspectZone = (m.inspectZone + 2) % 3
		m.inspectBodyOffset = 0
		return m, nil
	case "home":
		m.inspectZone = zoneWhatHappened
		m.inspectBodyOffset = 0
		return m, nil
	case "end":
		m.inspectZone = zoneBody
		m.inspectBodyOffset = 0
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
	}
	return m, nil
}
