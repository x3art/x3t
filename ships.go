package main

import (
	"log"
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
			err := st.tmpl.ExecuteTemplate(w, "ship", st.Ships[i])
			if err != nil {
				log.Print(err)
			}
			return
		}
	}

	http.NotFound(w, req)
}
