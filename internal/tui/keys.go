package tui

const (
	KeyLeft  = "h"
	KeyRight = "l"
	KeyDown  = "j"
	KeyUp    = "k"
)

type screen int

const (
	screenBoard screen = iota
	screenTaskDetail
	screenQuickAdd
	screenProjectView
	screenHelp
	screenSearch
	screenArchive
	screenEdit
)
