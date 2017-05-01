package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"x3t/xt"
)

type sectorReq struct {
	St *state
	S  *xt.Sector
}

func (st *state) sector(w http.ResponseWriter, req *http.Request) {
	s := strings.Split(strings.TrimPrefix(req.URL.Path, "/sector/"), "/")
	if len(s) != 2 {
		http.NotFound(w, req)
		return
	}
	x, err := strconv.Atoi(s[0])
	if err != nil {
		http.NotFound(w, req)
		return
	}
	y, err := strconv.Atoi(s[1])
	if err != nil {
		http.NotFound(w, req)
		return
	}
	sect := st.U.SectorXY(x, y)
	if sect == nil {
		http.NotFound(w, req)
	}
	err = st.tmpl.ExecuteTemplate(w, "sector", sectorReq{st, sect})
	if err != nil {
		log.Print(err)
	}
}

func (st *state) mapFuncs(fm template.FuncMap) {
	fm["sectName"] = func(s *xt.Sector) string {
		return s.Name(st.text)
	}
	fm["raceName"] = func(r int) string {
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
	fm["lnBreak"] = func(maxl int, s string) []string {
		ret := make([]string, 0)
		for _, e := range strings.Split(s, " ") {
			// If two substrings are shorter than maxl, combine them
			if len(ret) != 0 && len(ret[len(ret)-1])+len(e) < maxl {
				ret[len(ret)-1] = ret[len(ret)-1] + " " + e
			} else {
				ret = append(ret, e)
			}
		}
		return ret
	}
	fm["sectorIcons"] = func(s xt.Sector) []string {
		ret := []string{}
		if s.SunPercent() > 150 {
			ret = append(ret, "sunny")
		}
		sil, ore := 0, 0
		for i := range s.Asteroids {
			switch s.Asteroids[i].Type {
			case 0:
				ore += s.Asteroids[i].Amount
			case 1:
				sil += s.Asteroids[i].Amount
			}
		}
		if sil > 300 {
			ret = append(ret, "silicon")
		}
		if ore > 600 {
			ret = append(ret, "ore")
		}
		for i := range s.Docks {
			// Stab in the dark, but it matches some
			// equipment docks and military outposts.
			if st.Docks[s.Docks[i].S].GalaxySubtype == "SG_DOCK_EQUIP" {
				ret = append(ret, "dock")
				break
			}
		}
		return ret
	}
	fm["validGate"] = func(g xt.Gate) bool {
		if g.S == 4 {
			return false
		}
		switch g.Gid {
		case 0, 1, 2, 3:
			return true
		default:
			log.Print("unknown gatepos ", g)
			return false
		}
	}
	fm["asteroidType"] = func(i int) string {
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
}
