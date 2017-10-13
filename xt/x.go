package xt

import (
	"io"
	"sync"
)

// Each thing we access is loaded and parsed on demand. To synchronize
// this, each member is protected by a sync.Once.
type X struct {
	xf Xfiles

	textOnce sync.Once
	text     Text

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

func (x *X) Open(f string) io.ReadCloser {
	return x.xf.Open(f)
}

func (x *X) Map(f func(string, string)) {
	x.xf.Map(f)
}

// Asynchronously pre-cache everything we might want to load later.
func (x *X) PreCache() {
	go func() {
		_ = x.GetUniverse()
	}()
	for tk := range x.typeCache {
		go func() {
			_ = x.getType(tk)
		}()
	}
}
