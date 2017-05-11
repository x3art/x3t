package xt

// Apparently this is arcane, hardcoded knowledge that can not be
// extracted from any configuration file.

var laserTypes = []string{
	"SG_LASER_IRE",
	"SG_LASER_PAC",
	"SG_LASER_MASS",
	"SG_LASER_ARGON_LIGHT",
	"SG_LASER_TELADI_LIGHT",
	"SG_LASER_PARANID_LIGHT",
	"SG_LASER_HEPT",
	"SG_LASER_BORON_LIGHT",
	"SG_LASER_PBE",
	"SG_LASER_PIRATE_LIGHT",
	"SG_LASER_TERRAN_LIGHT",
	"SG_LASER_CIG",
	"SG_LASER_BORON_MEDIUM",
	"SG_LASER_SPLIT_MEDIUM",
	"SG_LASER_TERRAN_MEDIUM",
	"SG_LASER_TELADI_AF",
	"SG_LASER_ARGON_AF",
	"SG_LASER_SPLIT_AF",
	"SG_LASER_PARANID_AF",
	"SG_LASER_TERRAN_AF",
	"SG_LASER_PPC",
	"SG_LASER_BORON_HEAVY",
	"SG_LASER_TELADI_HEAVY",
	"SG_LASER_PIRATE_HEAVY",
	"SG_LASER_TERRAN_HEAVY",
	"SG_LASER_ARGON_BEAM",
	"SG_LASER_PARANID_BEAM",
	"SG_LASER_TERRAN_BEAM",
	"SG_LASER_SPECIAL",
	"SG_LASER_UNKNOWN1",
	"SG_LASER_UNKNOWN2",
	"SG_LASER_KYON",
}

var ltOff map[string]uint

func init() {
	ltOff = make(map[string]uint)
	for i := range laserTypes {
		ltOff[laserTypes[i]] = uint(i)
	}
}

func (x *X) LtMask(lt string) uint {
	return 1 << ltOff[lt]
}

func (x *X) SectorName(s *Sector) string {
	r, _ := x.GetText().Get(7, 1020000+100*(s.Y+1)+(s.X+1))
	return r
}

func (x *X) SectorFlavor(s *Sector) string {
	r, _ := x.GetText().Get(19, 1030000+100*(s.Y+1)+(s.X+1))
	return r
}

func (x *X) RaceName(r int) string {
	switch r {
	case 1:
		return "Argon"
	case 2:
		return "Boron"
	case 3:
		return "Split"
	case 4:
		return "Paranid"
	case 5:
		return "Teladi"
	case 6:
		return "Xenon"
	case 7:
		return "Kha'ak"
	case 8:
		return "Pirates"
	case 9:
		return "Goner"
	case 17:
		return "ATF"
	case 18:
		return "Terran"
	case 19:
		return "Yaki"
	default:
		return "Unknown"
	}
}

func (x *X) AsteroidType(i int) string {
	switch i {
	case 0:
		return "Ore"
	case 1:
		return "Silicon Wafers"
	case 2:
		return "Nividium"
	case 3:
		return "Ice"
	default:
		return "Unknown"
	}
}

func (x *X) SunPercent(s *Sector) int {
	tsuns := x.GetSuns()
	b := 0
	for i := range s.Suns {
		b += tsuns[s.Suns[i].S].Brightness
	}
	return b * 100 / 65536
}

func ShipSpeedMax(s *Ship) int {
	// I haven't been able to find any documentation for the calcuation below, but it seems to work.
	return (s.Speed + s.Speed*s.MaxEngineTuning/10) / 500
}
