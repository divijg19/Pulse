package tui

// Surface presents a Domain within a Region. Each Surface knows how to
// render a specific Domain's state into a bounded rectangular area.
// Substitutability Law: any Surface may be replaced without changing the
// Domain it presents.
type Surface interface {
	Render(region Region) string
}

// RequestSurface renders the REQUEST workspace.
type RequestSurface struct{ m Model }

func (s RequestSurface) Render(region Region) string { return s.m.renderRequest(region) }

// InspectSurface renders the INSPECT workspace.
type InspectSurface struct{ m Model }

func (s InspectSurface) Render(region Region) string { return s.m.renderInspect(region) }

// ReadySurface renders the idle OBSERVE state before any results exist.
type ReadySurface struct{ m Model }

func (s ReadySurface) Render(region Region) string { return s.m.renderReady(region) }

// TimelineSurface renders the Timeline view within OBSERVE.
type TimelineSurface struct{ m Model }

func (s TimelineSurface) Render(region Region) string { return s.m.renderTimeline(region) }

// LogsSurface renders the Logs view within OBSERVE.
type LogsSurface struct{ m Model }

func (s LogsSurface) Render(region Region) string { return s.m.renderLogs(region) }

// CompareSurface renders the COMPARE workspace.
type CompareSurface struct{ m Model }

func (s CompareSurface) Render(region Region) string { return s.m.renderCompare(region) }

// resolveSurface returns the Surface for the current Model state.
func (m Model) resolveSurface() Surface {
	switch {
	case m.workspace.dialog == dialogRequest:
		return RequestSurface{m: m}
	case m.workspace.compare.marked >= 0 && m.workspace.compare.active >= 0:
		return CompareSurface{m: m}
	case m.workspace.mode == modeInspect:
		return InspectSurface{m: m}
	case !m.running && len(m.results) == 0:
		return ReadySurface{m: m}
	case m.workspace.view == TimelineView:
		return TimelineSurface{m: m}
	default:
		return LogsSurface{m: m}
	}
}
