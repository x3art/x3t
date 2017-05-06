package xt

import "sync"

// Each thing we access is loaded and parsed on demand. To synchronize
// this, each member is protected by a sync.Once.
type X struct {
	xf Xfiles

	textOnce sync.Once
	text     Text

	shipsOnce sync.Once
	ships     []Ship

	docksOnce sync.Once
	docks     map[string]TDock

	sunsOnce sync.Once
	suns     []TSun

	universeOnce sync.Once
	universe     Universe

	shieldsOnce sync.Once
	shields     []TShield
}

// Get all the information we can get from an X3 installation.
func NewX(dir string) *X {
	return &X{xf: XFiles(dir)}
}
