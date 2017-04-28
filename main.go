package main

//go:generate go-bindata -prefix assets assets/...

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"x3t/xt"
)

var textFile = flag.String("-strings", "data/0001-L044.xml", "strings file")
var shipsFile = flag.String("-ships", "data/TShips.txt", "ships file")
var cockpitsFile = flag.String("-cockpits", "data/TCockpits.txt", "cockpits file")
var lasersFile = flag.String("-lasers", "data/TLaser.txt", "lasers file")
var universeFile = flag.String("-universe", "data/x3_universe.xml", "universe file")

type state struct {
	text     xt.Text
	Ships    []xt.Ship
	cockpits []xt.Cockpit
	lasers   []xt.Laser
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

	f := xf.Open("addon/types/TWareN.txt")
	a, _ := ioutil.ReadAll(f)
	fmt.Printf("%s\n", a)
	return

	st := state{}

	st.text = xt.GetText(*textFile)

	st.Ships = xt.GetShips(*shipsFile, st.text)
	// st.cockpits = xt.GetCockpits(*cockpitsFile, text)
	// st.lasers = xt.GetLasers(*lasersFile, text)
	st.U = xt.GetUniverse(*universeFile)

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
