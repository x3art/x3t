package xt

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"reflect"
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
			o.Decode(&s.Suns[len(s.Suns)-1])
		case 4: // Planets
		case 5: // Trading Dock
		case 6: // Factories Shipyards
		case 7: // Ships
		case 8, 9, 10, 11, 12, 13, 14, 15, 16: // Wares
		case 17: // Asteroids
			s.Asteroids = append(s.Asteroids, Asteroid{})
			o.Decode(&s.Asteroids[len(s.Asteroids)-1])
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
	Type   int `x3t:"o:atype"`
	Amount int `x3t:"o:aamount"`
	S      int `x3t:"o:s"`
	X      int `x3t:"o:x"`
	Y      int `x3t:"o:y"`
	Z      int `x3t:"o:z"`
	A      int `x3t:"o:a"`
	B      int `x3t:"o:b"`
	G      int `x3t:"o:g"`
	F      int `x3t:"o:f"`
}

type Sun struct {
	S     int `x3t:"o:s"`
	X     int `x3t:"o:x"`
	Y     int `x3t:"o:y"`
	Z     int `x3t:"o:z"`
	Color int `x3t:"o:color"`
	F     int `x3t:"o:f"`
}

type odec struct {
	i int
	k reflect.Kind
}
type odecoder map[string]odec

var ocache = map[reflect.Type]odecoder{}

func (o *O) Decode(data interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	dec := ocache[t]
	if dec == nil {
		dec = make(odecoder)
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get("x3t")
			tp := tagParse(tag)
			if field := tp["o"]; field != "" {
				dec[field] = odec{i, t.Field(i).Type.Kind()}
			}
		}
		ocache[t] = dec
	}
	for _, attr := range o.Attrs {
		if d, ok := dec[attr.Name.Local]; ok {
			switch d.k {
			case reflect.Int:
				i, err := strconv.Atoi(attr.Value)
				if err != nil {
					log.Fatal(err)
				}
				v.Field(d.i).SetInt(int64(i))
			default:
				log.Fatal("unknown field type")
			}
		} else {
			log.Printf("unknown attr %v: %v", attr.Name.Local, attr.Value)
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
