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
}

// Get all the information we can get from an X3 installation.
func NewX(dir string) *X {
	return &X{xf: XFiles(dir)}
}

func (x *X) GetText() Text {
	x.textOnce.Do(func() {
		x.text = GetText(x.xf)
	})
	return x.text
}

func (x *X) GetShips() []Ship {
	x.shipsOnce.Do(func() {
		x.ships = GetShips(x.xf, x.GetText())
	})
	return x.ships
}

func (x *X) GetDocks() map[string]TDock {
	x.docksOnce.Do(func() {
		x.docks = make(map[string]TDock)
		for _, d := range GetDocks(x.xf, x.GetText()) {
			x.docks[d.ObjectID] = d
		}
	})
	return x.docks
}

func (x *X) GetSuns() []TSun {
	x.sunsOnce.Do(func() {
		x.suns = GetSuns(x.xf, x.GetText())
	})
	return x.suns
}

func (x *X) GetUniverse() Universe {
	x.universeOnce.Do(func() {
		x.universe = GetUniverse(x.xf)
	})
	return x.universe
}
