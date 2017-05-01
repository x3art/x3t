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
	text     xt.Text
	Ships    []xt.Ship
	cockpits []xt.Cockpit
	lasers   []xt.Laser
	Docks    map[string]xt.TDock
	Suns     []xt.TSun
	U        xt.Universe
	tmpl     *template.Template
}

var rootTemplates = map[string]string{
	"/map":   "map",
	"/ships": "ships",
	"/about": "about",
}

func main() {
	flag.Parse()

	xf := xt.XFiles(flag.Arg(0))

	st := state{}

	st.text = xt.GetText(xf)

	st.Ships = xt.GetShips(xf, st.text)
	// st.cockpits = xt.GetCockpits(xf, text)
	// st.lasers = xt.GetLasers(xf, text)
	st.Docks = make(map[string]xt.TDock)
	for _, d := range xt.GetDocks(xf, st.text) {
		st.Docks[d.ObjectID] = d
	}
	st.U = xt.GetUniverse(xf)
	st.Suns = xt.GetSuns(xf, st.text)

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
