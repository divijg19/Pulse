package tui

import "github.com/divijg19/Pulse/internal/model"

// ViewType identifies which presentation surface the Workspace shows for
// the OBSERVE orientation.
type ViewType int

const (
	TimelineView ViewType = iota
	LogsView
)

// SessionState describes the lifecycle of a comparison session.
type SessionState int

const (
	SessionIdle           SessionState = iota // no active comparison workflow
	SessionBaselineMarked                     // baseline chosen, no candidate yet
	SessionComparing                          // both baseline and candidate set
)

// ComparisonSession tracks a single comparison workflow. A session begins
// when the operator marks a baseline result and ends when they explicitly
// clear it. The session is disposable — it is reset on startRun.
type ComparisonSession struct {
	BaselineIndex  int                 // baseline result index (-1 = unset)
	CandidateIndex int                 // candidate result index (-1 = unset)
	State          SessionState        // current lifecycle phase
	Analysis       *ComparisonAnalysis // computed analysis, refreshed on transitions
}

// CompareData holds all comparison state for the workspace. PinnedBaseline
// survives startRun; Session is ephemeral and reset on each new run.
type CompareData struct {
	PinnedBaseline *model.Result
	Session        ComparisonSession
}

// Workspace is the composition unit. A Workspace composes Views, a View
// composes Regions, a Region hosts Surfaces. Workspace is owned by Shell.
type Workspace struct {
	mode    mode
	dialog  dialog
	view    ViewType
	compare CompareData
}

// NewWorkspace creates the default Workspace (OBSERVE/Timeline).
func NewWorkspace() Workspace {
	return Workspace{
		mode:   modeObserve,
		dialog: dialogNone,
		view:   TimelineView,
		compare: CompareData{
			PinnedBaseline: nil,
			Session: ComparisonSession{
				BaselineIndex:  -1,
				CandidateIndex: -1,
				State:          SessionIdle,
				Analysis:       nil,
			},
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
