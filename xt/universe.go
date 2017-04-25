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
	F      int `x3t:"o:f"`
	X      int `x3t:"o:x"`
	Y      int `x3t:"o:y"`
	R      int `x3t:"o:r"`
	Size   int `x3t:"o:size"`
	M      int `x3t:"o:m"`
	P      int `x3t:"o:p"`
	Qtrade int `x3t:"o:qtrade"`
	Qfight int `x3t:"o:qfight"`
	Qthink int `x3t:"o:qthink"`
	Qbuild int `x3t:"o:qbuild"`

	Os []O `x3t:"os"`

	Suns []Sun

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
}

func (s *Sector) SunPercent() int {
	if len(s.Suns) == 1 {
		if s.Suns[0].S == 0 {
			return 100
		} else {
			return 150
		}
	} else {
		return 100 * len(s.Suns)
	}
}

func (s *Sector) Name(text Text) string {
	return text[7][1020000+100*(s.Y+1)+(s.X+1)]
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

type odecoder struct {
	fields   map[string]odec
	overflow int
}

type O struct {
	T     int        `xml:"t,attr"`
	Attrs []xml.Attr `xml:",any,attr"`
	Os    []O        `xml:"o"`
}

type Universe struct {
	Sectors []Sector
}

var ocache = map[reflect.Type]*odecoder{}

func (o *O) Decode(data interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	dec := ocache[t]
	if dec == nil {
		dec = &odecoder{fields: make(map[string]odec), overflow: -1}
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get("x3t")
			tp := tagParse(tag)
			if field := tp["o"]; field != "" {
				dec.fields[field] = odec{i, t.Field(i).Type.Kind()}
			}
			if tp["os"] != "" {
				dec.overflow = i
			}
		}
		ocache[t] = dec
	}
	for _, attr := range o.Attrs {
		if d, ok := dec.fields[attr.Name.Local]; ok {
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
	if dec.overflow != -1 {
		v.Field(dec.overflow).Set(reflect.ValueOf(o.Os))
	}
}

func GetUniverse(n string) Universe {
	f, err := os.Open(n)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)

	uo := O{}
	if err := d.Decode(&uo); err != nil {
		log.Fatal(err)
	}

	u := Universe{}
	for i := range uo.Os {
		o := &uo.Os[i]
		if o.T == 1 {
			u.Sectors = append(u.Sectors, Sector{})
			o.Decode(&u.Sectors[len(u.Sectors)-1])
		}
	}
	for i := range u.Sectors {
		u.Sectors[i].decodeOs()
	}

	return u
}
