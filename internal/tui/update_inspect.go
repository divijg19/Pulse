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
		switch m.workspace.compare.Session.State {
		case SessionIdle:
			if m.workspace.compare.PinnedBaseline != nil {
				m.workspace.compare.Session.CandidateIndex = m.selected
				m.workspace.compare.Session.State = SessionComparing
				m.workspace.compare.Session.BaselineIndex = -1
				m.workspace.mode = modeCompare
				m.inspectZone = zoneWhatHappened
				m.inspectBodyOffset = 0
				m.workspace.compare.Session.Analysis = m.computeComparisonAnalysis()
			} else {
				m.workspace.compare.Session.BaselineIndex = m.selected
				m.workspace.compare.Session.State = SessionBaselineMarked
				m.status = "Baseline marked"
				m.workspace.mode = modeObserve
			}
		case SessionBaselineMarked:
			if m.workspace.compare.Session.BaselineIndex == m.selected {
				m.workspace.compare.Session = ComparisonSession{BaselineIndex: -1, CandidateIndex: -1, State: SessionIdle, Analysis: nil}
				m.status = "Comparison cleared"
			} else {
				m.workspace.compare.Session.CandidateIndex = m.selected
				m.workspace.compare.Session.State = SessionComparing
				m.workspace.mode = modeCompare
				m.inspectZone = zoneWhatHappened
				m.inspectBodyOffset = 0
				m.workspace.compare.Session.Analysis = m.computeComparisonAnalysis()
			}
		case SessionComparing:
			if m.workspace.compare.Session.BaselineIndex == m.selected {
				m.workspace.compare.Session = ComparisonSession{BaselineIndex: -1, CandidateIndex: -1, State: SessionIdle, Analysis: nil}
				m.workspace.mode = modeObserve
				m.status = "Comparison cleared"
			} else {
				m.workspace.compare.Session.CandidateIndex = m.selected
				m.inspectZone = zoneWhatHappened
				m.inspectBodyOffset = 0
				m.workspace.compare.Session.Analysis = m.computeComparisonAnalysis()
			}
		}
		return m, nil
	case "q", "ctrl+c":
		m.workspace.dialog = dialogConfirmQuit
		return m, nil
	}
	return m, nil
}
