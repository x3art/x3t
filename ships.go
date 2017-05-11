package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"x3t/xt"
)

func (st *state) ship(w http.ResponseWriter, req *http.Request) {
	ships := st.x.GetShips()
	s := strings.SplitN(strings.TrimPrefix(req.URL.Path, "/ship/"), "/", 2)
	var name, variation string

	switch len(s) {
	case 1:
		name = s[0]
	case 2:
		name = s[0]
		variation = s[1]
	}

	for i := range ships {
		if ships[i].Description == name && ships[i].Variation == variation {
			err := st.tmpl.ExecuteTemplate(w, "ship", &ships[i])
			if err != nil {
				log.Print(err)
			}
			return
		}
	}

	http.NotFound(w, req)
}

// filters out a set of ships from all the ships
type shipFilter interface {
	Match(*xt.Ship) bool
}

// set union
type sfUnion []shipFilter

func (u sfUnion) Match(s *xt.Ship) bool {
	for _, m := range u {
		if m.Match(s) {
			return true
		}
	}
	return false
}

// set intersection
type sfIntersection []shipFilter

func (i sfIntersection) Match(s *xt.Ship) bool {
	for _, m := range i {
		if !m.Match(s) {
			return false
		}
	}
	return true
}

type sfInit func(string) shipFilter

type sfClass string

func (c sfClass) Match(s *xt.Ship) bool {
	return s.ClassDescription == string(c) || s.ClassDescription == "OBJ_SHIP_"+string(c)
}

func sfClassInit(s string) shipFilter {
	return sfClass(s)
}

type sfRace int

func (r sfRace) Match(s *xt.Ship) bool {
	return sfRace(s.Race) == r
}

func sfRaceInit(s string) shipFilter {
	x, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return sfRace(x)
}

type sfSpeed int

func (sp sfSpeed) Match(s *xt.Ship) bool {
	return xt.ShipSpeedMax(s) >= int(sp)
}

func sfSpeedInit(s string) shipFilter {
	x, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return sfSpeed(x)
}

type sfShields int

func (sh sfShields) Match(s *xt.Ship) bool {
	return xt.ShipShieldStr(s) >= int(sh)
}

func sfShieldsInit(s string) shipFilter {
	x, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return sfShields(x)
}

var shipFilters = map[string]sfInit{
	"class":       sfClassInit,
	"race":        sfRaceInit,
	"minMaxSpeed": sfSpeedInit,
	"minShields":  sfShieldsInit,
}

type shipsReq struct {
	Ships []*xt.Ship
	Q     url.Values
}

func (st *state) ships(w http.ResponseWriter, req *http.Request) {
	ships := st.x.GetShips()
	q := req.URL.Query()

	inter := sfIntersection{}

	// In a query like: /ships?foo=1&foo=2&bar=2 each of the
	// repeated queries (foo in this case) are in a union (OR) and
	// each uniqe query (foo,bar) is in an intersection.
	for qfilter, vals := range q {
		if sfinit, ok := shipFilters[qfilter]; ok {
			// If this was written for performance we'd
			// not create the union and intersection if
			// not necessary, but it isn't, so we won't
			// bother. Amusingly enough this comment is
			// longer than the code required.
			un := sfUnion{}
			for i := range vals {
				sf := sfinit(vals[i])
				if sf == nil {
					http.NotFound(w, req)
					return
				}
				un = append(un, sf)
			}
			inter = append(inter, un)
		} else {
			http.NotFound(w, req)
			return
		}
	}

	sr := shipsReq{Q: q}
	for i := range ships {
		s := &ships[i]
		if inter.Match(s) {
			sr.Ships = append(sr.Ships, s)
		}
	}
	err := st.tmpl.ExecuteTemplate(w, "ships", sr)
	if err != nil {
		log.Print(err)
	}
}

var cockpitPos = []string{
	"",
	"Front",
	"Rear",
	"Left",
	"Right",
	"Top",
	"Bottom",
}

func (st *state) shipFuncs(fm template.FuncMap) {
	fm["maskToLasers"] = func(mask int, wareclass int) (ret []*xt.TLaser) {
		ls := st.x.GetLasers()
		for i := range st.x.GetLasers() {
			if ls[i].WareClass <= wareclass && (st.x.LtMask(ls[i].Index)&uint(mask)) != 0 {
				ret = append(ret, &ls[i])
			}
		}
		return
	}
	fm["cockpitPos"] = func(p int) string {
		return cockpitPos[p]
	}
	fm["shipClassName"] = func(s string) string {
		if strings.HasPrefix(s, "OBJ_SHIP_") {
			// This is lazy
			return strings.TrimPrefix(s, "OBJ_SHIP_")
		}
		return "special"
	}
	fm["countGuns"] = func(x xt.Ship) int {
		n := 0
		for i := range x.GunGroup {
			n += x.GunGroup[i].NumGuns
		}
		return n
	}
	fm["shipClassList"] = func() []string {
		return []string{"M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "TS", "TL", "TM"}
	}
	fm["isChecked"] = func(q url.Values, k, v string) bool {
		for _, val := range q[k] {
			if val == v {
				return true
			}
		}
		return false
	}
	fm["raceList"] = func() []int {
		return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 17, 18, 19}
	}
	fm["shieldStr"] = xt.ShipShieldStr
	fm["ShipSpeedMax"] = xt.ShipSpeedMax
}
