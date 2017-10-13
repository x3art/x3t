package xt

import (
	"fmt"
	"math/bits"
	"strings"
)

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

// Shield strength in MJ
func ShipShieldStr(s *Ship) int {
	if s.ShieldType == nil {
		return 0
	}
	return s.ShieldType.Strength * s.MaxShieldCount / 1000
}

func (x *X) ShipDock(s *Ship) map[string]int {
	ret := make(map[string]int)
	left := s.DockingSlots
	dp := x.DockPorts()
	sc := x.Scene(fmt.Sprintf("objects/%s.pbd", strings.Replace(s.ShipScene, "\\", "/", -1)))
	for i := range sc.P {
		dps := dp[strings.ToLower(sc.P[i].B)]
		if dps == 0 {
			continue
		}
		key := []string{}
		for dps != 0 {
			bit := uint(bits.TrailingZeros(uint(dps)))
			key = append(key, dpname[1<<bit])
			dps ^= 1 << bit
		}
		k := strings.Join(key, "/")
		ret[k] = ret[k] + 1
		left--
	}
	ret["fighters"] = left
	return ret
}

const (
	DP_TS = 1 << iota
	DP_M6
	DP_M8
	DP_M1
	DP_M2
	DP_TL
)

var dpname = map[int]string{
	DP_TS: "TS",
	DP_M6: "M6",
	DP_M8: "M8",
	DP_M1: "M1",
	DP_TL: "TL",
	DP_M2: "M2",
}

var dpmap = map[string]int{
	"ANIMATEDF_DOCKPORT_TRANSPORT": DP_TS,
	"ANIMATEDF_DOCKPORT_M6":        DP_M6,
	"ANIMATEDF_DOCKPORT_M8":        DP_M8,
	"ANIMATEDF_DOCKPORT_STANDARD":  DP_TS | DP_M6,
	"ANIMATEDF_DOCKPORT_HUGESHIP":  DP_M1 | DP_M2 | DP_TL,
	"ANIMATEDF_DOCKPORT_BIGSHIP":   DP_M1 | DP_M2 | DP_M6 | DP_TL,
	"ANIMATEDF_DOCKPORT_STARTONLY": 0,
	"ANIMATEDF_DOCKPORT_FIGHTER":   0,
	"ANIMATEDF_DOCKPORT_HANGAR":    0,
	"ANIMATEDF_DOCKPORT_LANDONLY":  0,
	"ANIMATEDF_DOCKPORT_M5":        0,
	"NULL":                         0,
	"ANIMATEDF_DOCKPORT_UDDOWN":      0,
	"ANIMATEDF_DOCKPORT_UDBACK":      0,
	"ANIMATEDF_DOCKPORT_BELOW30":     0,
	"ANIMATEDF_DOCKPORT_BELOW60":     0,
	"ANIMATEDF_DOCKPORT_UDFWD":       0,
	"ANIMATEDF_DOCKPORT_QUICKLAUNCH": 0,
	"ANIMATEDF_DOCKPORT_FWDPUSH":     0,
	"ANIMATEDF_DOCKPORT_M4":          0,
}

func (x *X) DockPorts() map[string]int {
	ret := make(map[string]int)
	dums := x.GetDum()
	miss := make(map[string]int)
	for i := range dums {
		d := &dums[i]
		flags := 0
		for _, f := range strings.Split(d.Flags, "|") {
			fl, ok := dpmap[f]
			if !ok {
				miss[f] = 0
			}
			flags |= fl
		}
		if flags != 0 {
			ret[strings.ToLower(d.Id)] = flags
		}
	}
	for k := range miss {
		fmt.Printf("\"%s\": 0,\n", k)
	}
	return ret
}
