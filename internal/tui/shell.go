package tui

// Region is a rectangular area within a layout. Each Region has a type that
// identifies its role in the composition.
type Region struct {
	Type   RegionType
	Width  int
	Height int
}

// ShellLayout divides the terminal into the three permanent Shell regions.
type ShellLayout struct {
	Context   Region
	Workspace Region
	Command   Region
}

// ActionCategory defines the four semantic groups an action can belong to.
type ActionCategory int

const (
	NavigationCategory ActionCategory = iota
	ConfigurationCategory
	OperationCategory
	ApplicationCategory
)

// ActionID identifies an operator intent.
type ActionID int

const (
	ActionSelect ActionID = iota
	ActionInspect
	ActionSwitchView
	ActionConfigureRequest
	ActionRun
	ActionCancel
	ActionNextField
	ActionSwitchMethod
	ActionAdjustConcurrency
	ActionNextHeader
	ActionAddHeader
	ActionDeleteHeader
	ActionBack
	ActionQuit
	ActionConfirmQuit
	ActionCtrlCQuit
	ActionQQuit
	ActionDismissCancel
)

// Action is a behavioral intent — not a presentation object.
type Action struct {
	ID       ActionID
	Category ActionCategory
	Enabled  bool
}

// configItem represents a single configuration value in the Context region.
type configItem struct {
	Identity string
	Value    string
	Valid    bool
}

// ShellState is an immutable snapshot of the shell-level state produced once
// per frame.
type ShellState struct {
	Orientation   string
	Configuration []configItem
	Actions       []Action
}

// ActionPriority defines the semantic importance of an operator action.
type ActionPriority int

const (
	PriorityCritical ActionPriority = iota
	PriorityHigh
	PriorityMedium
	PriorityLow
)

// actionBinding maps an operator intent (ActionID) to its presentation in the
// operator ribbon.
type actionBinding struct {
	Key      string
	Label    string
	Category ActionCategory
	Priority ActionPriority
}

var actionBindings = map[ActionID]actionBinding{
	ActionSelect:            {"↑↓", "Select", NavigationCategory, PriorityHigh},
	ActionInspect:           {"Enter", "Inspect", NavigationCategory, PriorityHigh},
	ActionSwitchView:        {"Tab", "Views", NavigationCategory, PriorityMedium},
	ActionConfigureRequest:  {"e", "Request", ConfigurationCategory, PriorityCritical},
	ActionRun:               {"Ctrl+R", "Run", OperationCategory, PriorityCritical},
	ActionCancel:            {"Ctrl+X", "Cancel", OperationCategory, PriorityCritical},
	ActionNextField:         {"Tab", "Next Field", ConfigurationCategory, PriorityHigh},
	ActionSwitchMethod:      {"←→", "Method", ConfigurationCategory, PriorityHigh},
	ActionAdjustConcurrency: {"↑↓", "Adjust", ConfigurationCategory, PriorityHigh},
	ActionNextHeader:        {"Tab", "Next", ConfigurationCategory, PriorityHigh},
	ActionAddHeader:         {"Ctrl+N", "Header", ConfigurationCategory, PriorityHigh},
	ActionDeleteHeader:      {"Ctrl+D", "Delete", ConfigurationCategory, PriorityLow},
	ActionBack:              {"Esc", "Back", ApplicationCategory, PriorityCritical},
	ActionQuit:              {"q", "Quit", ApplicationCategory, PriorityCritical},
	ActionConfirmQuit:       {"Enter", "Quit", ApplicationCategory, PriorityCritical},
	ActionCtrlCQuit:         {"Ctrl+C", "Quit", ApplicationCategory, PriorityCritical},
	ActionQQuit:             {"q", "Quit", ApplicationCategory, PriorityCritical},
	ActionDismissCancel:     {"Any", "Cancel", ApplicationCategory, PriorityCritical},
}

// Shell is the permanent outer boundary of Pulse. It owns orientation,
// the context bar, the ribbon, and terminal geometry.
type Shell struct {
	width  int
	height int
}

func NewShell() Shell {
	return Shell{width: 80, height: 24}
}

func (s *Shell) Resize(w, h int) {
	s.width = w
	s.height = h
}

func (s Shell) Dimensions() (int, int) {
	return s.width, s.height
}

func (s Shell) Layout() ShellLayout {
	return computeShellLayout(s.width, s.height)
}

func computeShellLayout(totalWidth, totalHeight int) ShellLayout {
	width := max(72, totalWidth)
	bodyHeight := max(1, totalHeight-5)
	return ShellLayout{
		Context:   Region{Type: ContextRegion, Width: width, Height: 1},
		Workspace: Region{Type: WorkspaceRegion, Width: width, Height: bodyHeight},
		Command:   Region{Type: CommandRegion, Width: width, Height: 1},
	}
}

func orientationLabel(m Model) string {
	return m.workspace.Orientation()
}
