package main

import (
	"flag"
	"html/template"
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
	st.mapFuncs(fm)
	st.tmpl = tmpls.Compile(fm)

	//	http.HandleFunc("/ship/", st.ship)
	//	http.HandleFunc("/ships", st.shiplist)
	http.HandleFunc("/map", st.showMap)

	http.HandleFunc("/js/svg-pan-zoom.min.js", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "js/svg-pan-zoom.min.js")
	})

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
