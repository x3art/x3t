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
}
