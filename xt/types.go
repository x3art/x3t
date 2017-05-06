package xt

import (
	"fmt"
	"reflect"
	"strconv"
)

var typeMap = map[string]struct {
	fn string
	t  reflect.Type
}{
	"Suns":    {"addon/types/TSuns.txt", reflect.TypeOf(TSun{})},
	"Shields": {"addon/types/TShields.txt", reflect.TypeOf(TShield{})},
	"Ships":   {"addon/types/TShips.txt", reflect.TypeOf(Ship{})},
}

func (x *X) typeLookup(typ string, value string, index bool) (reflect.Value, error) {
	t := x.getType(typ)
	if index {
		i, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(t).Index(i).Addr(), nil
	}
	return reflect.Value{}, fmt.Errorf("not implemented")
}

func (x *X) getType(t string) interface{} {
	x.typeCache[t].once.Do(func() {
		f := x.xf.Open(typeMap[t].fn)
		defer f.Close()
		v := reflect.Indirect(reflect.New(reflect.SliceOf(typeMap[t].t)))
		x.tparsev(f, v)
		x.typeCache[t].v = v.Interface()
	})
	return x.typeCache[t].v
}

func (x *X) GetSuns() []TSun {
	return x.getType("Suns").([]TSun)
}

// the only documentation of this I found was wrong.
type TSun struct {
	Unknown01  int
	Unknown02  int
	Unknown03  int
	Unknown04  int
	Unknown05  float64
	Unknown06  int
	Unknown07  int
	Unknown08  int
	Unknown09  int
	Brightness int // Seems to be fixed point scaled by 2^16
	Unknown11  int
	Unknown12  int
	Unknown13  int
	Unknown14  int
	Unknown15  int
	Unknown16  int
	Unknown17  int
	Unknown18  int
	ObjectID   string
}

func (x *X) GetShips() []Ship {
	return x.getType("Ships").([]Ship)
}

type Ship struct {
	BodyFile  string
	PictureID string
	Yaw       float64
	Pitch     float64
	Roll      float64
	// Class - ship class. Names of classes can be changed but classes itself are hardcoded into OBJ files
	Class        string
	Description  string `x3t:"page:17"`
	Speed        int
	Acceleration int
	// Engine sound - Index to Sounds.txt
	EngineSound          string
	AverageReactionDelay string
	// Engine effect - Index to Effects.txt
	EngineEffect     string
	EngineGlowEffect string
	ReactorOutput    string
	SoundVolumeMax   string
	SoundVolumeMin   string
	// Ship scene - scene containing the ship graphics
	ShipScene string
	// Cockpit scene - scene containing the cockpit graphics (the real cockpit where you control the ship from)
	CockPitScene string
	//Possible lasers - bit mask
	PossibleLasers int
	// Gun count - sum of count of laser parts of all gun records
	GunCount int
	// Weapons energy - how is it related to TLaser.txt energy?
	WeaponsEnergy int
	// Weapons recharge rate - how is it related to TLaser.txt recharge rate?
	WeaponsRechargeRate float64
	// Shield type - biggest shield the ship can equip - index to TShields.txt
	//ShieldType string
	ShieldType *TShield `x3t:"tref:Shields,index"`
	// Max shield count - Maximum number of shields
	MaxShieldCount int
	// Possible missiles - bit mask
	PossibleMissiles int
	// Number of missiles (NPC) - Maximum number of missiles an NPC ship can carry
	NumberOfMissiles int
	// Max # of engine tunning
	MaxEngineTuning int
	// Max # of rudder tunning
	MaxRudderTuning int
	// Cargo min (buy) - minimum cargo capacity (when the ship is bought)
	CargoMin int
	// Cargo max - maximum cargo capacity
	CargoMax        int
	PredefinedWares string
	// Turret descriptor - fixed length array - the reason why there is only 6 + 1 turrets
	TurretDescriptor [6]struct {
		// Cockpit index - index to TCockpits.txt
		CIndex int
		//Cockpit *Cockpit `x3t:"tref:Cockpits,index"`
		// Cockpit position - front, rear, left, right, top, bottom - not sure what it's used for
		CPos int
	}
	// Docking slots - maximum number of ships which can dock
	DockingSlots int
	// Cargo type - maximum cargo size the ship can carry - Ware class of TWare.txts
	CargoType string
	// Race - Race of the ship. Probably only used in Jobs.txt
	Race         string
	HullStrength int
	// Explosion definition - index to Effects.txt
	ExplosionDefinition string
	// Body explosion definition - index to Effects.txt
	BodyExplosionDefinition string
	// Engine Trail - Particle Emitter - index to Effects.txt
	EngineTrailParticleEmitter string
	// Variation index - Hauler, Vanguard,... - index to Page 17 in text resource files. The String ID is calculated as 10000 + Variation index
	Variation string `x3t:"page:17,offset:10000,ignore:20"`
	// Max Rotation Acceleration - How fast the ship can go from 0 to maximum rotation speed
	MaxRotationAcceleration int
	// Class Description - String ID from Page 17 of text resource files (no, it isn't)
	ClassDescription string
	Cockpit          []struct {
		Index       int
		TurretIndex string
		BodyID      string
		PathIndex   string
	}
	GunGroup []struct {
		// Initial laser index - calculated as 1 + count of laser parts in previous gun groups
		InitialLaserIndex int
		// No of guns - number of laser parts
		NumGuns int
		// Index - an index of the gun group - starting from 1
		GunGroupIndex int
		Gun           []struct {
			// Index - an index of the gun. The index continues between gun groups (i.e. it's unique and global)
			Index int
			// Count of laser parts - number of laser parts in BOD/BOB file
			CountLaserParts int
			// Body ID (primary) - Body ID from ship scene
			BodyID string
			// Path index (primary) - Path index from ship scene
			PathIndex string
			// Body ID (secondary) - Body ID from weapon scene
			BodyID2 string
			// Path index (secondary) - Path index from weapon scene
			PathIndex2 string
		}
	}
	Volume string
	// Production RelVal (NPC) - Price for NPCs (it's not really a price)
	ProductionRelValNPC int
	// Price modifier (1)
	PriceModifier1 int
	// Price modifier (2)
	PriceModifier2 int
	// Ware class - ignored
	WareClass string
	// Production RelVal (player) - Price for the player (it's not really a price)
	ProductionRelValPlayer int
	// Min. Notoriety - minimum notoriety the player must have with corresponding race to be able to buy the ship
	MinNotoriety int
	// Video ID - ignored
	VideoID string
	// Unknown value
	UknownValue string
	// Ship ID - identifier of the ship
	ShipID string
}

/*
func GetCockpits(xf Xfiles, text Text) []Cockpit {
	f := xf.Open("addon/types/TCockpits.txt")
	defer f.Close()
	ret := []Cockpit{}
	x.tparse(f, &ret)
	return ret
}
*/

type Cockpit struct {
	BodyFile               string
	PictureID              string
	RotX                   float64
	RotY                   float64
	RotZ                   float64
	GalaxySubtype          string
	Description            string
	SceneFile              string
	LaserMask              int
	Volume                 string
	ProductionRelValNPC    string
	PriceModifier1         string
	PriceModifier2         string
	WareClass              string
	ProductionRelValPlayer string
	MinNotoriety           string
	VideoID                string
	Skin                   string
	ObjectID               string
}

func (x *X) GetDocks() map[string]TDock {
	x.docksOnce.Do(func() {
		f := x.xf.Open("addon/types/TDocks.txt")
		defer f.Close()
		ds := []TDock{}
		x.tparse(f, &ds)
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

/*
func GetLasers(xf Xfiles, text Text) []Laser {
	f := xf.Open("addon/types/TLasers.txt")
	defer f.Close()
	ret := []Laser{}
	tparse(f, text, &ret)
	return ret
}
*/

type Laser struct {
	BodyFile               string
	PictureID              string
	RotX                   float64
	RotY                   float64
	RotZ                   float64
	Index                  string
	Description            string `x3t:"page:17"`
	RoF                    int
	Sound                  int
	Projectile             int
	Energy                 int
	ChargeRate             float64
	HUDIcon                string
	Volume                 string
	ProductionRelValNPC    string
	PriceMod1              string
	PriceMod2              string
	WareClass              string
	ProductionRelValPlayer string
	MinNotoriety           string
	VideoID                string
	Skin                   string
	OjectID                string
}

func (x *X) GetShields() []TShield {
	return x.getType("Shields").([]TShield)
}

type TShield struct {
	BodyFile               string
	PictureID              string
	Yaw                    float64
	Pitch                  float64
	Roll                   float64
	Index                  string
	Description            string `x3t:"page:17"`
	ChargeRate             int    // kJ/s
	Strength               int    // kJ
	HitEffect              string
	Efficiency             string
	Volume                 string
	ProductionRelValNPC    int
	PriceModifier1         int
	PriceModifier2         int
	WareClass              string
	ProductionRelValPlayer int
	MinNotoriety           int
	VideoID                string
	Skin                   string
	ObjectID               string
}
