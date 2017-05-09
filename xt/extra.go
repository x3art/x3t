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
