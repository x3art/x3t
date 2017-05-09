package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"x3t/xt"
)

func (st *state) ship(w http.ResponseWriter, req *http.Request) {
	s := strings.SplitN(strings.TrimPrefix(req.URL.Path, "/ship/"), "/", 2)
	var name, variation string

	switch len(s) {
	case 1:
		name = s[0]
	case 2:
		name = s[0]
		variation = s[1]
	}

	for i := range st.Ships {
		if st.Ships[i].Description == name && st.Ships[i].Variation == variation {
			err := st.tmpl.ExecuteTemplate(w, "ship", st.Ships[i])
			if err != nil {
				log.Print(err)
			}
			return
		}
	}

	http.NotFound(w, req)
}

type shipsReq struct {
	Ships []*xt.Ship
}

type shipFilter interface {
	Match(*xt.Ship) bool
}

type sfClass string

func (c sfClass) Match(s *xt.Ship) bool {
	return s.ClassDescription == string(c) || s.ClassDescription == "OBJ_SHIP_"+string(c)
}

func sfClassInit(s string) shipFilter {
	return sfClass(s)
}

type sfUnion []shipFilter

func (u sfUnion) Match(s *xt.Ship) bool {
	for _, m := range u {
		if m.Match(s) {
			return true
		}
	}
	return false
}

type sfIntersection []shipFilter

func (i sfIntersection) Match(s *xt.Ship) bool {
	for _, m := range i {
		if !m.Match(s) {
			return false
		}
	}
	return true
}

type sfTrue struct{}

func (_ sfTrue) Match(s *xt.Ship) bool {
	return true
}

type sfInit func(string) shipFilter

var shipFilters = map[string]sfInit{
	"class": sfClassInit,
}

func (st *state) ships(w http.ResponseWriter, req *http.Request) {
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
				un = append(un, sfinit(vals[i]))
			}
			inter = append(inter, un)
		} else {
			http.NotFound(w, req)
		}
	}

	sr := shipsReq{}
	for i := range st.Ships {
		s := &st.Ships[i]
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
}
