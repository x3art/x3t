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

func (st *state) ships(w http.ResponseWriter, req *http.Request) {
	sr := shipsReq{}
	for i := range st.Ships {
		sr.Ships = append(sr.Ships, &st.Ships[i])
	}
	err := st.tmpl.ExecuteTemplate(w, "ships", st)
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
