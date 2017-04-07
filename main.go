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
	fmt.Println(h, rec)
}

func pi(rec []string, data interface{}) []string {
	ret, err := pstruct(rec, reflect.Indirect(reflect.ValueOf(data)))
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

var pKinds = map[reflect.Kind]func([]string, reflect.Value) ([]string, error){
	reflect.Int:     pint,
	reflect.Float64: pfloat,
	reflect.String:  pstring,
	reflect.Array:   parray,
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
	return rec, nil
}

func pvalue(rec []string, v reflect.Value) ([]string, error) {
	fn, ok := pKinds[v.Kind()]
	if !ok {
		return rec, fmt.Errorf("bad kind: %v", v.Kind())
	}
	return fn(rec, v)
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
	BodyFile             string // Body file - not used
	PictureID            string // Picture ID - not used
	Yaw                  float64
	Pitch                float64
	Roll                 float64
	Class                string // Class - ship class. Names of classes can be changed but classes itself are hardcoded into OBJ files
	Description          string // Description - String ID from Page 17 of text resource files
	Speed                int
	Acceleration         int
	EngineSound          string // Engine sound - Index to Sounds.txt
	AverageReactionDelay string
	EngineEffect         string // Engine effect - Index to Effects.txt
	EngineGlowEffect     string
	ReactorOutput        string
	SoundVolumeMax       string
	SoundVolumeMin       string
	ShipScene            string  // Ship scene - scene containing the ship graphics
	CockPitScene         string  // Cockpit scene - scene containing the cockpit graphics (the real cockpit where you control the ship from)
	PossibleLasers       int     //Possible lasers - bit mask
	GunCount             int     // Gun count - sum of count of laser parts of all gun records
	WeaponsEnergy        int     // Weapons energy - how is it related to TLaser.txt energy?
	WeaponsRechargeRate  float64 // Weapons recharge rate - how is it related to TLaser.txt recharge rate?
	ShieldType           string  // Shield type - biggest shield the ship can equip - index to TShields.txt
	MaxShieldCount       int     // Max shield count - Maximum number of shileds",
	PossibleMissiles     int     // Possible missiles - bit mask
	NumberOfMissiles     int     // Number of missiles (NPC) - Maximum number of missiles an NPC ship can carry
	MaxEngineTuning      int     // Max # of engine tunning
	MaxRudderTuning      int     // Max # of rudder tunning
	CargoMin             int     // Cargo min (buy) - minimum cargo capacity (when the ship is bought)",
	CargoMax             int     // Cargo max - maximum cargo capacity",
	PredefinedWares      string
	TurretDescriptor     [6]struct{ CIndex, CPos int }
}

var tm = []string{
	"Turret descriptor - fixed length array - the reason why there is only 6 + 1 turrets",
	"TD 1",
	"Cockpit index - index to TCockpits.txt",
	"Cockpit position - front, rear, left, right, top, bottom - not sure what it's used for",
	"TD 2",
	"Cockpit index",
	"Cockpit position",
	"TD 3",
	"Cockpit index",
	"Cockpit position",
	"TD 4",
	"Cockpit index",
	"Cockpit position",
	"TD 5",
	"Cockpit index",
	"Cockpit position",
	"TD 6",
	"Cockpit index",
	"Cockpit position",
	"Docking slots - maximum number of ships which can dock",
	"Cargo type - maximum cargo size the ship can carry - Ware class of TWare.txts",
	"Race - Race of the ship. Probably only used in Jobs.txt",
	"Hull strength",
	"Explosion definition - index to Effects.txt",
	"Body explosion definition - index to Effects.txt",
	"Engine Trail - Particle Emitter - index to Effects.txt",
	"Variation index - Hauler, Vanguard,... - index to Page 17 in text resource files. The String ID is calculated as 10000 + Variation index",
	"Max Rotation Acceleration - How fast the ship can go from 0 to maximum rotation speed",
	"Class Description - String ID from Page 17 of text resource files",
	"Cockpit count - number of cockpit records",
	"Cockpit - repeatable",
	"Index - an index of the cockpit - starting from 1",
	"Turret index - an index starting from 0. It's not clear what it's purpose (if any)",
	"Body ID - Body ID from ship's scene for this cockpit",
	"Path index - Path index of the above Body ID",
	"Gun groups count - number of gun group records",
	"Gun group - repeatable",
	"Initial laser index - calculated as 1 + count of laser parts in previous gun groups",
	"No of guns - number of laser parts",
	"Index - an index of the gun group - starting from 1",
	"No of gun records - number of gun records",
	"Gun - repeatable",
	"Index - an index of the gun. The index continues between gun groups (i.e. it's unique and global)",
	"Count of laser parts - number of laser parts in BOD/BOB file",
	"Body ID (primary) - Body ID from ship scene",
	"Path index (primary) - Path index from ship scene",
	"Body ID (secondary) - Body ID from weapon scene",
	"Path index (secondary) - Path index from weapon scene",
	"Volume - ignored",
	"Production RelVal (NPC) - Price for NPCs (it's not really a price)",
	"Price modifier (1)",
	"Price modifier (2)",
	"Ware class - ignored",
	"Production RelVal (player) - Price for the player (it's not really a price)",
	"Min. Notoriety - minimum notoriety the player must have with corresponding race to be able to buy the ship",
	"Video ID - ignored",
	"Unknown value",
	"Ship ID - identifier of the ship",
}
