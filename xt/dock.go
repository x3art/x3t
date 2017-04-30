package xt

func GetDocks(xf Xfiles, text Text) []TDock {
	f := xf.Open("addon/types/TDocks.txt")
	defer f.Close()
	ret := []TDock{}
	tparse(f, text, &ret)
	return ret
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
	BosyExplosion          string
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
