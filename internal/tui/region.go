package tui

// RegionType distinguishes the two kinds of region in the Shell layout.
type RegionType int

const (
	ContextRegion RegionType = iota
	WorkspaceRegion
	CommandRegion
)
