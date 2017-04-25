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
}

func main() {
	flag.Parse()

	text := xt.GetText(*textFile)

	/*
		u := xt.GetUniverse(*universeFile)
		for i := range u.Sectors {
			s := &u.Sectors[i]
			if s.X == 8 && s.Y == 6 {
				fmt.Printf("%s: %d | %v\n", s.Name(text), s.SunPercent(), s.Suns)
			}
		}
	*/
	st := state{
		text:     text,
		ships:    xt.GetShips(*shipsFile, text),
		cockpits: xt.GetCockpits(*cockpitsFile, text),
		lasers:   xt.GetLasers(*lasersFile, text),
		u:        xt.GetUniverse(*universeFile),
	}

	http.HandleFunc("/ship/", st.ship)
	http.HandleFunc("/ships/", st.shiplist)
	log.Fatal(http.ListenAndServe(":8080", nil))

	/*
		ship := flag.Arg(0)

		s, _ := json.MarshalIndent(ships[ship], "", "\t")
		fmt.Printf("%s", s)
		s, _ = json.MarshalIndent(cockpits[ships[ship].TurretDescriptor[0].CIndex], "", "\t")
		fmt.Printf("%s", s)

		l := cockpits[ships[ship].TurretDescriptor[0].CIndex].LaserMask
		for i := uint(0); i < 64; i++ {
			if l&(1<<i) != 0 {
				fmt.Println(lasers[i].Description)
			}
		}
	*/
}

var tmpl = template.Must(template.New("all").Parse(`
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
`))

var _ = template.Must(tmpl.New("shiplist").Parse(`
{{template "header"}}
 <ul>
{{- range .}}
  <li>{{.Description}} {{.Variation}}
{{- end}}
 </ul>
{{template "footer"}}
`))

func (st *state) shiplist(w http.ResponseWriter, req *http.Request) {
	tmpl.ExecuteTemplate(w, "shiplist", st.ships)
}

var _ = template.Must(tmpl.New("ship").Parse(`
{{template "header"}}
  {{.Description}} {{.Variation}}<br/>
  Cargo: {{.CargoMin}} - {{.CargoMax}}<br/>
{{template "footer"}}
`))

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
			tmpl.ExecuteTemplate(w, "ship", st.ships[i])
			return
		}
	}

	http.NotFound(w, req)
}
