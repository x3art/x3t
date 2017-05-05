package main

//go:generate go-bindata -prefix assets assets/...

import (
	"bytes"
	"flag"
	"html/template"
	"log"
	"net/http"
	"x3t/xt"
)

type state struct {
	x     *xt.X
	Ships []xt.Ship
	Docks map[string]xt.TDock
	Suns  []xt.TSun
	U     xt.Universe
	tmpl  *template.Template
}

var rootTemplates = map[string]string{
	"/map":   "map",
	"/ships": "ships",
	"/about": "about",
}

func main() {
	flag.Parse()

	st := state{}

	st.x = xt.NewX(flag.Arg(0))

	st.Ships = st.x.GetShips()
	st.Docks = st.x.GetDocks()
	st.Suns = st.x.GetSuns()
	st.U = st.x.GetUniverse()

	fm := make(template.FuncMap)
	st.mapFuncs(fm)

	st.tmpl = template.New("")
	st.tmpl.Funcs(fm)
	if tmplDir, err := AssetDir("templates"); err == nil {
		for _, tn := range tmplDir {
			template.Must(st.tmpl.New(tn).Parse(string(MustAsset("templates/" + tn))))
		}
	}

	for n := range rootTemplates {
		t := rootTemplates[n]
		http.HandleFunc(n, func(w http.ResponseWriter, req *http.Request) {
			err := st.tmpl.ExecuteTemplate(w, t, st)
			if err != nil {
				log.Fatal(err)
			}
		})
	}

	http.HandleFunc("/ship/", st.ship)
	http.HandleFunc("/sector/", st.sector)

	if staticDir, err := AssetDir("static"); err == nil {
		for _, n := range staticDir {
			fn := "static/" + n
			http.HandleFunc("/"+fn, func(w http.ResponseWriter, req *http.Request) {
				ai, err := AssetInfo(fn)
				if err != nil {
					log.Fatal(err)
				}
				http.ServeContent(w, req, fn, ai.ModTime(), bytes.NewReader(MustAsset(fn)))
			})
		}
	}

	log.Printf("now")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
