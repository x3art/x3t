package xt

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Sector struct {
	F      int `xml:"f,attr"`
	X      int `xml:"x,attr"`
	Y      int `xml:"y,attr"`
	R      int `xml:"r,attr"`
	Size   int `xml:"size,attr"`
	M      int `xml:"m,attr"`
	P      int `xml:"p,attr"`
	Qtrade int `xml:"qtrade,attr"`
	Qfight int `xml:"qfight,attr"`
	Qthink int `xml:"qthink,attr"`
	Qbuild int `xml:"qbuild,attr"`

	Os []O `xml:"o"`

	Suns       []Sun
	SunPercent int

	Asteroids []Asteroid
}

func (s *Sector) decodeOs() {
	for i := range s.Os {
		o := &s.Os[i]
		switch o.T {
		case 2: // Background
		case 3:
			s.Suns = append(s.Suns, Sun{})
			s.Suns[len(s.Suns)-1].Decode(o.Attrs)
		case 4: // Planets
		case 5: // Trading Dock
		case 6: // Factories Shipyards
		case 7: // Ships
		case 8, 9, 10, 11, 12, 13, 14, 15, 16: // Wares
		case 17: // Asteroids
			s.Asteroids = append(s.Asteroids, Asteroid{})
			s.Asteroids[len(s.Asteroids)-1].Decode(o.Attrs)
		case 18: // Gates
		case 20: // Specials
		default:
			fmt.Printf("unknown type %d\n", o.T)
		}
	}
	if len(s.Suns) == 1 {
		if s.Suns[0].S == 0 {
			s.SunPercent = 100
		} else {
			s.SunPercent = 150
		}
	} else {
		s.SunPercent = 100 * len(s.Suns)
	}
}

func (s *Sector) Name(text Text) string {
	return text[7][1020000+100*(s.Y+1)+(s.X+1)]
}

type O struct {
	T     int        `xml:"t,attr"`
	Attrs []xml.Attr `xml:",any,attr"`
	Os    []O        `xml:"o"`
}

type Universe struct {
	Sectors []Sector `xml:"o"`
}

type Asteroid struct {
	Type   int
	Amount int
}

func (a *Asteroid) Decode(attrs []xml.Attr) {
	for _, attr := range attrs {
		i, err := strconv.Atoi(attr.Value)
		if err != nil {
			log.Fatal(err)
		}
		switch attr.Name.Local {
		case "s":
		case "x":
		case "y":
		case "z":
		case "a":
		case "b":
		case "g":
		case "atype":
			a.Type = i
		case "aamount":
			a.Amount = i
		default:
			log.Print("Unknown attr k: ", attr.Name.Local, attr.Value)
		}
	}
}

type Sun struct {
	S     int
	X     int
	Y     int
	Z     int
	Color int
	F     int
}

func (s *Sun) Decode(attrs []xml.Attr) {
	for _, attr := range attrs {
		i, err := strconv.Atoi(attr.Value)
		if err != nil {
			log.Fatal(err)
		}
		switch attr.Name.Local {
		case "s":
			s.S = i
		case "x":
			s.X = i
		case "y":
			s.Y = i
		case "z":
			s.Z = i
		case "color":
			s.Color = i
		case "f":
			s.F = i
		default:
			log.Print("Unknown attr k: ", attr.Name.Local, attr.Value)
		}
	}
}

func GetUniverse(n string) Universe {
	f, err := os.Open(n)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	u := Universe{}
	if err := d.Decode(&u); err != nil {
		log.Fatal(err)
	}

	for i := range u.Sectors {
		u.Sectors[i].decodeOs()
	}

	return u
}
