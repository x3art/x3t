package xt

import "sync"

type typeCache struct {
	once sync.Once
	v    interface{}
}

// Each thing we access is loaded and parsed on demand. To synchronize
// this, each member is protected by a sync.Once.
type X struct {
	xf Xfiles

	textOnce sync.Once
	text     Text

	docksOnce sync.Once
	docks     map[string]TDock

	typeCache map[string]*typeCache

	universeOnce sync.Once
	universe     Universe
}

// Get all the information we can get from an X3 installation.
func NewX(dir string) *X {
	x := &X{xf: XFiles(dir)}
	x.typeCache = make(map[string]*typeCache)
	for k := range typeMap {
		x.typeCache[k] = &typeCache{}
	}
	return x
}
