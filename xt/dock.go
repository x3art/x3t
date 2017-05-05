package xt

func (x *X) GetDocks() map[string]TDock {
	x.docksOnce.Do(func() {
		f := x.xf.Open("addon/types/TDocks.txt")
		defer f.Close()
		ds := []TDock{}
		tparse(f, x.GetText(), &ds)
		x.docks = make(map[string]TDock)
		for _, d := range ds {
			x.docks[d.ObjectID] = d
		}
	})

	return x.docks
}

type TDock struct {
	BodyFile               string
	PictureID              string
	RotX                   float64
	RotY                   float64
	RotZ                   float64
	GalaxySubtype          string
	Description            string `x3t:"page:17"`
	SoundID                string
	DockDistance           string
	RendezvousDistrance    string
	ThreeDSoundVolume      string
	SceneFile              string
	InnerScene             string
	Race                   int
	Explosion              string
	BodyExplosion          string
	ShieldPowerGen         string
	HUDIcon                string
	Volume                 string
	ProductionRelValNPC    int
	PriceModifier1         int
	PriseModifier2         int
	WareClass              string
	ProductionRelValPlayer int
	MinNotoriety           int
	VideoID                string
	Skin                   string
	ObjectID               string
}
