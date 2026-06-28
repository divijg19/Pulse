package tui

// Region is a rectangular area within a layout.
type Region struct {
	Width  int
	Height int
}

// ShellLayout divides the terminal into the three permanent Shell regions.
type ShellLayout struct {
	Context   Region
	Workspace Region
	Command   Region
}

// ShellColumnWidth is the fixed width reserved for the orientation label
// in the operator ribbon.
const ShellColumnWidth = 16

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

// actionBinding maps an operator intent (ActionID) to its presentation in the
// operator ribbon.
type actionBinding struct {
	Key      string
	Label    string
	Category ActionCategory
}

var actionBindings = map[ActionID]actionBinding{
	ActionSelect:            {"↑↓", "Select", NavigationCategory},
	ActionInspect:           {"Enter", "Inspect", NavigationCategory},
	ActionSwitchView:        {"Tab", "Views", NavigationCategory},
	ActionConfigureRequest:  {"e", "Request", ConfigurationCategory},
	ActionRun:               {"Ctrl+R", "Run", OperationCategory},
	ActionCancel:            {"Ctrl+X", "Cancel", OperationCategory},
	ActionNextField:         {"Tab", "Next Field", ConfigurationCategory},
	ActionSwitchMethod:      {"←→", "Method", ConfigurationCategory},
	ActionAdjustConcurrency: {"↑↓", "Adjust", ConfigurationCategory},
	ActionNextHeader:        {"Tab", "Next", ConfigurationCategory},
	ActionAddHeader:         {"Ctrl+N", "Header", ConfigurationCategory},
	ActionDeleteHeader:      {"Ctrl+D", "Delete", ConfigurationCategory},
	ActionBack:              {"Esc", "Back", ApplicationCategory},
	ActionQuit:              {"q", "Quit", ApplicationCategory},
	ActionConfirmQuit:       {"Enter", "Quit", ApplicationCategory},
	ActionCtrlCQuit:         {"Ctrl+C", "Quit", ApplicationCategory},
	ActionQQuit:             {"q", "Quit", ApplicationCategory},
	ActionDismissCancel:     {"Any", "Cancel", ApplicationCategory},
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
		Context:   Region{Width: width, Height: 1},
		Workspace: Region{Width: width, Height: bodyHeight},
		Command:   Region{Width: width, Height: 1},
	}
}

func orientationLabel(m Model) string {
	switch {
	case m.dialog == dialogConfirmQuit:
		return "QUIT"
	case m.mode == modeInspect:
		return "INSPECT"
	case m.dialog == dialogRequest:
		return "REQUEST"
	default:
		return "OBSERVE"
	}
}
