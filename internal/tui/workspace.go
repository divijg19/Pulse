package tui

import "github.com/divijg19/Pulse/internal/model"

// ViewType identifies which presentation surface the Workspace shows for
// the OBSERVE orientation.
type ViewType int

const (
	TimelineView ViewType = iota
	LogsView
)

// CompareState describes the lifecycle phase of a comparison workflow.
type CompareState int

const (
	CompareIdle           CompareState = iota // no active comparison workflow
	CompareBaselineMarked                     // baseline chosen, no candidate yet
	CompareComparing                          // both baseline and candidate set
)

// CompareView is a presentation of the comparison analysis. Every view
// consumes the same immutable CompareContext; switching views never triggers
// recomputation, only a different presentation.
type CompareView int

const (
	CompareViewOverview CompareView = iota
	CompareViewEvidence
	CompareViewDiff
	CompareViewHeaders
	CompareViewBody
	CompareViewRaw
)

// compareViewCount is the number of views; used to wrap bracket navigation.
const compareViewCount = 6

// compareViewNames maps each view to its heading label.
var compareViewNames = [compareViewCount]string{
	CompareViewOverview: "Overview",
	CompareViewEvidence: "Evidence",
	CompareViewDiff:     "Diff",
	CompareViewHeaders:  "Headers",
	CompareViewBody:     "Body",
	CompareViewRaw:      "Raw",
}

// CompareContext is the renderer-facing projection of a comparison. It contains
// only presentation data. The renderer never knows whether Baseline originated
// from the current session or a reference request — it simply renders what it is
// given. Workflow details remain inside CompareWorkspace.
type CompareContext struct {
	Baseline  model.Result
	Candidate model.Result
	Analysis  *ComparisonAnalysis
}

// CompareWorkspace owns all comparison state. It is the single source of truth
// for workflow transitions; renderers never mutate it. Every transition is
// expressed through exactly one operation method below.
type CompareWorkspace struct {
	Baseline  *model.Result // current baseline (resolved value)
	Candidate *model.Result // current candidate (resolved value)
	Reference *model.Result // canonical reference request (survives runs)
	State     CompareState
	View      CompareView
	Analysis  *ComparisonAnalysis
}

// NewCompareWorkspace returns an empty CompareWorkspace.
func NewCompareWorkspace() CompareWorkspace {
	return CompareWorkspace{
		Baseline:  nil,
		Candidate: nil,
		Reference: nil,
		State:     CompareIdle,
		View:      CompareViewOverview,
		Analysis:  nil,
	}
}

// --- Predicates ------------------------------------------------------------

// HasBaseline reports whether a baseline result is available.
func (w CompareWorkspace) HasBaseline() bool { return w.Baseline != nil }

// IsComparing reports whether both baseline and candidate are set.
func (w CompareWorkspace) IsComparing() bool { return w.State == CompareComparing }

// HasReference reports whether a reference request survives across runs.
func (w CompareWorkspace) HasReference() bool { return w.Reference != nil }

// IsBaselineResult reports whether r is the current baseline.
func (w CompareWorkspace) IsBaselineResult(r model.Result) bool {
	return w.Baseline != nil && resultsEqual(*w.Baseline, r)
}

// IsCandidateResult reports whether r is the current candidate.
func (w CompareWorkspace) IsCandidateResult(r model.Result) bool {
	return w.Candidate != nil && resultsEqual(*w.Candidate, r)
}

// --- Workflow operations ---------------------------------------------------

// MarkBaseline establishes r as the baseline. Any prior candidate is dropped;
// the session enters BaselineMarked. The analysis is cleared until a candidate
// is selected.
func (w *CompareWorkspace) MarkBaseline(r model.Result) {
	b := r
	w.Baseline = &b
	w.Candidate = nil
	w.State = CompareBaselineMarked
	w.Analysis = nil
	w.View = CompareViewOverview
}

// Unmark clears the baseline and any candidate, returning to Idle. The
// reference request is preserved.
func (w *CompareWorkspace) Unmark() {
	w.Baseline = nil
	w.Candidate = nil
	w.State = CompareIdle
	w.Analysis = nil
}

// SelectCandidate establishes r as the candidate and enters the Comparing
// state. The analysis is recomputed exactly once.
func (w *CompareWorkspace) SelectCandidate(r model.Result) {
	c := r
	w.Candidate = &c
	w.State = CompareComparing
	w.refreshAnalysis()
}

// ReplaceCandidate replaces the current candidate with r. The analysis is
// recomputed exactly once.
func (w *CompareWorkspace) ReplaceCandidate(r model.Result) {
	c := r
	w.Candidate = &c
	w.refreshAnalysis()
}

// ReplaceBaseline replaces the current baseline with r. The analysis is
// recomputed exactly once.
func (w *CompareWorkspace) ReplaceBaseline(r model.Result) {
	b := r
	w.Baseline = &b
	w.refreshAnalysis()
}

// Swap exchanges the baseline and candidate. The analysis is recomputed so the
// directional verdict reflects the new perspective. Comparison content (the
// underlying deltas) is unchanged.
func (w *CompareWorkspace) Swap() {
	w.Baseline, w.Candidate = w.Candidate, w.Baseline
	w.refreshAnalysis()
}

// Clear ends the active comparison session. Baseline, candidate and analysis
// are reset to zero; the reference survives.
func (w *CompareWorkspace) Clear() {
	w.Baseline = nil
	w.Candidate = nil
	w.State = CompareIdle
	w.Analysis = nil
}

// RenounceReference clears the reference request, ending its persistence across runs.
// Only this operation removes the reference.
func (w *CompareWorkspace) RenounceReference() {
	w.Reference = nil
}

// NextView advances to the next comparison view, wrapping around.
func (w *CompareWorkspace) NextView() {
	w.View = (w.View + 1) % compareViewCount
}

// PrevView moves to the previous comparison view, wrapping around.
func (w *CompareWorkspace) PrevView() {
	w.View = (w.View + compareViewCount - 1) % compareViewCount
}

// refreshAnalysis recomputes the ComparisonAnalysis from the current baseline
// and candidate. It is the single place analysis is produced. It is a no-op
// unless both baseline and candidate are present.
func (w *CompareWorkspace) refreshAnalysis() {
	if w.State == CompareComparing && w.Baseline != nil && w.Candidate != nil {
		a := AnalyzeComparison(*w.Baseline, *w.Candidate)
		w.Analysis = &a
	}
}

// Context projects the workspace into a renderer-facing CompareContext. It
// resolves the baseline and candidate into plain values so the renderer never
// needs to know their origin.
func (w CompareWorkspace) Context() CompareContext {
	ctx := CompareContext{Analysis: w.Analysis}
	if w.Baseline != nil {
		ctx.Baseline = *w.Baseline
	}
	if w.Candidate != nil {
		ctx.Candidate = *w.Candidate
	}
	return ctx
}

// Workspace is the composition unit. A Workspace composes Views, a View
// composes Regions, a Region hosts Surfaces. Workspace is owned by Shell.
type Workspace struct {
	mode    mode
	dialog  dialog
	view    ViewType
	compare CompareWorkspace
}

// NewWorkspace creates the default Workspace (OBSERVE/Timeline).
func NewWorkspace() Workspace {
	return Workspace{
		mode:    modeObserve,
		dialog:  dialogNone,
		view:    TimelineView,
		compare: NewCompareWorkspace(),
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
