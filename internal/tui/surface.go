package tui

// Surface presents a Domain within a Region. Each Surface knows how to
// render a specific Domain's state into a bounded rectangular area.
// Substitutability Law: any Surface may be replaced without changing the
// Domain it presents.
type Surface interface {
	Render(region Region) string
}

// surfaceFunc adapts a function to the Surface interface.
type surfaceFunc func(Region) string

func (f surfaceFunc) Render(region Region) string {
	return f(region)
}

// resolveSurface returns the Surface for the current Model state.
func (m Model) resolveSurface() Surface {
	switch {
	case m.dialog == dialogRequest:
		return surfaceFunc(m.renderRequest)
	case m.mode == modeInspect:
		return surfaceFunc(m.renderInspect)
	case !m.running && len(m.results) == 0:
		return surfaceFunc(m.renderReady)
	case m.view == viewTimeline:
		return surfaceFunc(m.renderTimeline)
	default:
		return surfaceFunc(m.renderLogs)
	}
}
