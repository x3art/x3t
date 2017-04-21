package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
)

func main() {
	flag.Parse()
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(f)
	r.Comment = '/'
	r.Comma = ';'
	hdr, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}
	x, err := strconv.Atoi(hdr[0])
	if err != nil {
		log.Fatal(err)
	}

	_ = x

	for i := 0; i < 30; i++ {
		r.FieldsPerRecord = 0
		rec, err := r.Read()
		if err != nil {
			log.Fatal(err)
		}
		parse(rec)
	}
	fmt.Println(r.FieldsPerRecord)
}

func parse(rec []string) {
	var h Hdr
	rec = pi(rec, &h)
	fmt.Println(h)
	fmt.Println(rec)
}

func pi(rec []string, data interface{}) []string {
	ret, err := pstruct(rec, reflect.Indirect(reflect.ValueOf(data)))
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func pint(rec []string, v reflect.Value) ([]string, error) {
	n, err := strconv.Atoi(rec[0])
	if err != nil {
		return rec, err
	}
	v.SetInt(int64(n))
	return rec[1:], nil
}

func pfloat(rec []string, v reflect.Value) ([]string, error) {
	n, err := strconv.ParseFloat(rec[0], 64)
	if err != nil {
		return rec, err
	}
	v.SetFloat(n)
	return rec[1:], nil
}

func pstring(rec []string, v reflect.Value) ([]string, error) {
	v.SetString(rec[0])
	return rec[1:], nil
}

func parray(rec []string, v reflect.Value) ([]string, error) {
	for i := 0; i < v.Len(); i++ {
		var err error
		rec, err = pvalue(rec, v.Index(i))
		if err != nil {
			return rec, fmt.Errorf("Array field (%d): %v", i, err)
		}
	}
	return rec, nil
}

func pvalue(rec []string, v reflect.Value) ([]string, error) {
	switch v.Kind() {
	case reflect.Int:
		return pint(rec, v)
	case reflect.Float64:
		return pfloat(rec, v)
	case reflect.String:
		return pstring(rec, v)
	case reflect.Array:
		return parray(rec, v)
	case reflect.Struct:
		return pstruct(rec, v)
	default:
		return rec, fmt.Errorf("bad kind: %v", v.Kind())
	}
}

func pstruct(rec []string, v reflect.Value) ([]string, error) {
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		var err error
		rec, err = pvalue(rec, fv)
		if err != nil {
			return rec, fmt.Errorf("Parse Field (%s): %v", v.Type().Field(i).Name, err)
		}
	}
	return rec, nil
}

type Hdr struct {
	// Body file - not used
	BodyFile string
	// Picture ID - not used
	PictureID string
	Yaw       float64
	Pitch     float64
	Roll      float64
	// Class - ship class. Names of classes can be changed but classes itself are hardcoded into OBJ files
	Class string
	// Description - String ID from Page 17 of text resource files
	Description  string
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
	ShieldType string
	// Max shield count - Maximum number of shileds",
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
	VariationIndex string
	// Max Rotation Acceleration - How fast the ship can go from 0 to maximum rotation speed
	MaxRotationAcceleration int
	// Class Description - String ID from Page 17 of text resource files
	ClassDescription string
}

/*
	// Cockpit count - number of cockpit records
	 // Cockpit - repeatable
	  // Index - an index of the cockpit - starting from 1
	  // Turret index - an index starting from 0. It's not clear what it's purpose (if any)
	  // Body ID - Body ID from ship's scene for this cockpit
	  // Path index - Path index of the above Body ID
	// Gun groups count - number of gun group records
	 // Gun group - repeatable
	  // Initial laser index - calculated as 1 + count of laser parts in previous gun groups
	  // No of guns - number of laser parts
	  // Index - an index of the gun group - starting from 1
	  // No of gun records - number of gun records
	   // Gun - repeatable
	    // Index - an index of the gun. The index continues between gun groups (i.e. it's unique and global)
	    // Count of laser parts - number of laser parts in BOD/BOB file
	    // Body ID (primary) - Body ID from ship scene
	    // Path index (primary) - Path index from ship scene
	    // Body ID (secondary) - Body ID from weapon scene
	    // Path index (secondary) - Path index from weapon scene
	// Volume - ignored
	// Production RelVal (NPC) - Price for NPCs (it's not really a price)
	// Price modifier (1)
	// Price modifier (2)
	// Ware class - ignored
	// Production RelVal (player) - Price for the player (it's not really a price)
	// Min. Notoriety - minimum notoriety the player must have with corresponding race to be able to buy the ship
	// Video ID - ignored
	// Unknown value
	// Ship ID - identifier of the ship
*/
