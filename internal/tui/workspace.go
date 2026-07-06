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
	mode    mode
	dialog  dialog
	view    ViewType
	compare compareState
}

// compareState holds the investigation comparison lifecycle. Both fields are
// result slice indices. Values of -1 mean unset. Esc destroys the state.
type compareState struct {
	marked int
	active int
}

// NewWorkspace creates the default Workspace (OBSERVE/Timeline).
func NewWorkspace() Workspace {
	return Workspace{
		mode:   modeObserve,
		dialog: dialogNone,
		view:   TimelineView,
		compare: compareState{
			marked: -1,
			active: -1,
		},
	}
}

// Orientation returns the current orientation string based on workspace state.
func (w Workspace) Orientation() string {
	switch {
	case w.dialog == dialogConfirmQuit:
		return "QUIT"
	case w.dialog == dialogRequest:
		return "REQUEST"
	case w.mode == modeCompare:
		return "COMPARE"
	case w.mode == modeInspect:
		return "INSPECT"
	default:
		return "OBSERVE"
	}
}
