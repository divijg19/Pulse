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

	surface Surface
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

// IsRequesting returns true when the REQUEST dialog is active.
func (w Workspace) IsRequesting() bool { return w.dialog == dialogRequest }

// IsInspecting returns true when INSPECT mode is active.
func (w Workspace) IsInspecting() bool { return w.mode == modeInspect }

// IsObserving returns true when in the default OBSERVE mode with no dialog.
func (w Workspace) IsObserving() bool {
	return w.mode == modeObserve && w.dialog == dialogNone
}

// IsQuitting returns true when the confirm-quit dialog is shown.
func (w Workspace) IsQuitting() bool { return w.dialog == dialogConfirmQuit }
