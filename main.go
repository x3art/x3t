package main

//go:generate go-bindata -prefix assets assets/...

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"image"
	"log"
	"net/http"
	"time"
	"x3t/xt"

	"image/png"

	_ "github.com/lukegb/dds"
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

func calc(vals ...interface{}) (int, error) {
	s := make([]int, 0)
	for _, vi := range vals {
		switch v := vi.(type) {
		case int:
			s = append(s, v)
		case string:
			l := len(s)
			if l < 2 {
				return 0, fmt.Errorf("calc stack underflow")
			}
			a, b := s[l-2], s[l-1]
			s = s[:l-1]
			r := &s[l-2]
			switch v {
			case "+":
				*r = a + b
			case "-":
				*r = a - b
			case "*":
				*r = a * b
			case "/":
				*r = a / b
			}
		default:
			return 0, fmt.Errorf("bad type %v", vi)
		}
	}
	if len(s) != 1 {
		return 0, fmt.Errorf("calc stack overflow")
	}
	return s[0], nil
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
	fm["calc"] = calc

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

	http.HandleFunc("/foo.png", func(w http.ResponseWriter, req *http.Request) {
		img, _, err := image.Decode(st.x.Open("dds/interface_icons_XT_diff.dds"))
		if err != nil {
			log.Fatal(err)
		}
		b := bytes.NewBuffer(nil)
		err = png.Encode(b, img)
		if err != nil {
			log.Fatal(err)
		}
		// xxx - it wold be nice if xf could return a ModTime.
		http.ServeContent(w, req, "/foo.png", time.Time{}, bytes.NewReader(b.Bytes()))
	})

	log.Printf("now")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
