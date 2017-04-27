package main

import (
	"html/template"
	"strings"
	"x3t/xt"
)

func (st *state) mapFuncs(fm template.FuncMap) {
	fm["sectorName"] = func(s xt.Sector) []string {
		ret := make([]string, 0)
		for _, e := range strings.Split(s.Name(st.text), " ") {
			// If two substrings are shorter than 11, combine them
			if len(ret) != 0 && len(ret[len(ret)-1])+len(e) < 11 {
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
		if len(s.Docks) > 0 {
			ret = append(ret, "dock")
		}
		return ret
	}
}
