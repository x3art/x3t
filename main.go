package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"strings"
	"x3t/xt"
)

var textFile = flag.String("-strings", "data/0001-L044.xml", "strings file")
var shipsFile = flag.String("-ships", "data/TShips.txt", "ships file")
var cockpitsFile = flag.String("-cockpits", "data/TCockpits.txt", "cockpits file")
var lasersFile = flag.String("-lasers", "data/TLaser.txt", "lasers file")
var universeFile = flag.String("-universe", "data/x3_universe.xml", "universe file")

type state struct {
	text     xt.Text
	ships    []xt.Ship
	cockpits []xt.Cockpit
	lasers   []xt.Laser
	u        xt.Universe
	tmpl     *template.Template
}

func main() {
	flag.Parse()

	st := state{}

	st.text = xt.GetText(*textFile)

	// st.ships = xt.GetShips(*shipsFile, text)
	// st.cockpits = xt.GetCockpits(*cockpitsFile, text)
	// st.lasers = xt.GetLasers(*lasersFile, text)
	st.u = xt.GetUniverse(*universeFile)

	fm := make(template.FuncMap)
	fm["sectorName"] = func(s xt.Sector) []string {
		sp := strings.Split(s.Name(st.text), " ")
		ret := make([]string, 0)
		for _, e := range sp {
			// If two substrings are shorter than 12, combine them
			if len(ret) != 0 && len(ret[len(ret)-1])+len(e) < 12 {
				ret[len(ret)-1] = ret[len(ret)-1] + " " + e
			} else {
				ret = append(ret, e)
			}
		}
		return ret
	}

	st.tmpl = tmpls.Compile(fm)

	//	http.HandleFunc("/ship/", st.ship)
	//	http.HandleFunc("/ships", st.shiplist)
	http.HandleFunc("/map", st.showMap)

	log.Printf("now")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

var _ = tmpls.Add("all", `
{{- define "header" -}}
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<title> foobar </title>
</head>
<body>
{{- end -}}
{{- define "footer" -}}
</body>
</html>
{{- end -}}
`)
