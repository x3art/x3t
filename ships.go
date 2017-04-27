package main

import (
	"net/http"
	"strings"
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
			st.tmpl.ExecuteTemplate(w, "ship", st.Ships[i])
			return
		}
	}

	http.NotFound(w, req)
}
