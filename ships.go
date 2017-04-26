package main

import (
	"net/http"
	"strings"
)

var _ = tmpls.Add("shiplist", `
{{template "header"}}
 <ul>
{{- range .}}
  <li><a href="/ship/{{.Description}}{{if .Variation}}/{{.Variation}}{{end}}">{{.Description}} {{.Variation}}</a>
{{- end}}
 </ul>
{{template "footer"}}
`)

func (st *state) shiplist(w http.ResponseWriter, req *http.Request) {
	st.tmpl.ExecuteTemplate(w, "shiplist", st.ships)
}

var _ = tmpls.Add("ship", `
{{template "header"}}
  {{.Description}} {{.Variation}}<br/>
  Cargo: {{.CargoMin}} - {{.CargoMax}}<br/>
{{template "footer"}}
`)

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

	for i := range st.ships {
		if st.ships[i].Description == name && st.ships[i].Variation == variation {
			st.tmpl.ExecuteTemplate(w, "ship", st.ships[i])
			return
		}
	}

	http.NotFound(w, req)
}
