package tui

// RegionType distinguishes the three kinds of region in the Shell layout.
type RegionType int

const (
	RegionUnset RegionType = iota
	ContextRegion
	WorkspaceRegion
	CommandRegion
)
