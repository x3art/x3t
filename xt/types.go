package xt

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"
)

type typeCache struct {
	once sync.Once
	v    interface{}
	byid map[string]interface{}
}

var typeMap = map[string]struct {
	fn string
	t  reflect.Type
}{
	"Suns":          {"addon/types/TSuns.txt", reflect.TypeOf(TSun{})},
	"Shields":       {"addon/types/TShields.txt", reflect.TypeOf(TShield{})},
	"Ships":         {"addon/types/TShips.txt", reflect.TypeOf(Ship{})},
	"Cockpits":      {"addon/types/TCockpits.txt", reflect.TypeOf(Cockpit{})},
	"Lasers":        {"addon/types/TLaser.txt", reflect.TypeOf(TLaser{})},
	"Docks":         {"addon/types/TDocks.txt", reflect.TypeOf(TDock{})},
	"Bullets":       {"addon/types/TBullets.txt", reflect.TypeOf(TBullet{})},
	"DummyAnimated": {"addon/types/Dummies.txt", reflect.TypeOf(DummyAnimated{})},
}

func (x *X) typeLookup(typ string, value string, index bool) (reflect.Value, error) {
	t := x.getType(typ).v
	if index {
		i, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, err
		}
		if i < 0 {
			return reflect.Zero(reflect.ValueOf(t).Index(0).Addr().Type()), nil
		}
		return reflect.ValueOf(t).Index(i).Addr(), nil
	}
	return reflect.Value{}, fmt.Errorf("not implemented")
}

func (x *X) getType(t string) *typeCache {
	tc := x.typeCache[t]
	tc.once.Do(func() {
		idField, hasID := typeMap[t].t.FieldByName("ObjectID")
		if hasID {
			tc.byid = make(map[string]interface{})
		}
		f := x.xf.Open(typeMap[t].fn)
		defer f.Close()
		v := reflect.Indirect(reflect.New(reflect.SliceOf(typeMap[t].t)))
		x.tparsev(f, v)
		tc.v = v.Interface()
		if hasID {
			for i := 0; i < v.Len(); i++ {
				elem := v.Index(i)
				tc.byid[elem.FieldByIndex(idField.Index).String()] = elem.Addr().Interface()
			}
		}
	})
	return tc
}

func (x *X) GetSuns() []TSun {
	return x.getType("Suns").v.([]TSun)
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
	return x.getType("Ships").v.([]Ship)
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
	Speed        int    // Seems to be scaled by 500.
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
	MaxEngineTuning  int // One engine tuning seems to increase the speed by 10% of the minimum speed.
	MaxRudderTuning  int
	// Cargo min (buy) - minimum cargo capacity (when the ship is bought)
	CargoMin int
	// Cargo max - maximum cargo capacity
	CargoMax        int
	PredefinedWares string
	// Turret descriptor - fixed length array - the reason why there is only 6 + 1 turrets
	TurretDescriptor [6]struct {
		// Cockpit index - index to TCockpits.txt
		//CIndex int
		Cockpit *Cockpit `x3t:"tref:Cockpits,index"`
		// Cockpit position - front, rear, left, right, top, bottom - not sure what it's used for
		CPos int
	}
	// Docking slots - maximum number of ships which can dock
	DockingSlots int
	// Cargo type - maximum cargo size the ship can carry - Ware class of TWare.txts
	CargoType int
	// Race - Race of the ship.
	Race         int
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
	WareClass      int
	// Production RelVal (player) - Price for the player (it's not really a price)
	ProductionRelValPlayer int
	// Min. Notoriety - minimum notoriety the player must have with corresponding race to be able to buy the ship
	MinNotoriety int
	// Video ID - ignored
	VideoID string
	// Unknown value
	UknownValue string
	ObjectID    string
}

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
	WareClass              int
	ProductionRelValPlayer string
	MinNotoriety           string
	VideoID                string
	Skin                   string
	ObjectID               string
}

func (x *X) DockByID(id string) *TDock {
	return x.getType("Docks").byid[id].(*TDock)
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
	WareClass              int
	ProductionRelValPlayer int
	MinNotoriety           int
	VideoID                string
	Skin                   string
	ObjectID               string
}

func (x *X) GetLasers() []TLaser {
	return x.getType("Lasers").v.([]TLaser)
}

type TLaser struct {
	BodyFile               string
	PictureID              string
	RotX                   float64
	RotY                   float64
	RotZ                   float64
	Index                  string
	Description            string `x3t:"page:17"`
	RoF                    int    // 1 shot per RoF milliseconds
	Sound                  int
	Projectile             *TBullet `x3t:"tref:Bullets,index"`
	Energy                 int
	ChargeRate             float64
	HUDIcon                string
	Volume                 string
	ProductionRelValNPC    string
	PriceMod1              string
	PriceMod2              string
	WareClass              int
	ProductionRelValPlayer string
	MinNotoriety           string
	VideoID                string
	Skin                   string
	OjectID                string
}

func (x *X) GetShields() []TShield {
	return x.getType("Shields").v.([]TShield)
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
	WareClass              int
	ProductionRelValPlayer int
	MinNotoriety           int
	VideoID                string
	Skin                   string
	ObjectID               string
}

type TBullet struct {
	BodyFile               string
	PictureID              string
	Yaw                    float64
	Pitch                  float64
	Roll                   float64
	Index                  string
	Description            string `x3t:"page:17"`
	ShieldDamage           int
	EnergyUsed             int
	Sound                  string // index to Sounds.txt
	Lifetime               int
	Speed                  int
	Flags                  int
	ColorB                 int
	ColorG                 int
	ColorR                 int
	SizeX                  float64
	SizeY                  float64
	SizeZ                  float64
	EngineEffect           string // Effects.txt
	ImpactEffect           string // Effects.txt
	LauchEffect            string // Effects.txt
	HullDamage             int
	EngineTrail            string
	AmbientSound           string // Sounds.txt
	VolumeMin              int
	VolumeMax              int
	Unknown1               int
	Unknown2               int
	Unknown3               int
	Unknown4               int
	Unknown5               int
	Unknown6               int
	Unknown7               int
	Unknown8               int
	Unknown9               int
	Unknown10              int
	Unknown11              int
	Unknown12              int
	Ammo                   string // TWareT.txt (or 128?)
	ProductionRelValNPC    int
	PriceModifier1         int
	PriceModifier2         int
	WareClass              int
	ProductionRelValPlayer int
	MinNotoriety           int
	VideoID                string
	Skin                   string
	ObjectID               string
}

type DummyAnimated struct {
	Id       string
	Flags    string
	Unknown1 int
	Unknown2 string
	Unknown3 int
}

func (x *X) GetDum() []DummyAnimated {
	return x.getType("DummyAnimated").v.([]DummyAnimated)
}
