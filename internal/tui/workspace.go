package tui

// ViewType identifies which presentation surface the Workspace shows for
// the OBSERVE orientation.
type ViewType int

const (
	TimelineView ViewType = iota
	LogsView
)

// Workspace is the composition unit. A Workspace composes Views, a View
// composes Regions, a Region hosts Surfaces. Workspace is owned by Shell.
type Workspace struct {
	mode   mode
	dialog dialog
	view   ViewType
}

// NewWorkspace creates the default Workspace (OBSERVE/Timeline).
func NewWorkspace() Workspace {
	return Workspace{
		mode:   modeObserve,
		dialog: dialogNone,
		view:   TimelineView,
	}
}

// Orientation returns the current orientation string based on workspace state.
func (w Workspace) Orientation() string {
	switch {
	case w.dialog == dialogConfirmQuit:
		return "QUIT"
	case w.dialog == dialogRequest:
		return "REQUEST"
	case w.mode == modeInspect:
		return "INSPECT"
	default:
		return "OBSERVE"
	}
}
