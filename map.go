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
	U     xt.Universe
	Docks map[string]*xt.TDock
	S     *xt.Sector
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
	u := st.x.GetUniverse()
	sect := u.SectorXY(x, y)
	if sect == nil {
		http.NotFound(w, req)
	}
	err = st.tmpl.ExecuteTemplate(w, "sector", sectorReq{u, st.x.GetDocks(), sect})
	if err != nil {
		log.Print(err)
	}
}

func (st *state) mapFuncs(fm template.FuncMap) {
	fm["SectorName"] = st.x.SectorName
	fm["SectorFlavor"] = st.x.SectorFlavor
	fm["raceName"] = st.x.RaceName
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
	fm["sectorIcons"] = func(s *xt.Sector) []string {
		ret := []string{}
		if st.x.SunPercent(s) > 150 {
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
			if st.x.DockByID(s.Docks[i].S).GalaxySubtype == "SG_DOCK_EQUIP" {
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
	fm["asteroidType"] = st.x.AsteroidType
	fm["sunPercent"] = st.x.SunPercent
}
