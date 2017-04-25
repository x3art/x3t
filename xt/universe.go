package xt

import (
	"encoding/xml"
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

	Suns []Sun `x3t:"ot:3"`

	Asteroids []Asteroid `x3t:"ot:17"`
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

type Universe struct {
	Sectors []Sector `x3t:"ot:1"`
}

type odec struct {
	i int
	k reflect.Kind
}

type odecoder struct {
	fields   map[string]odec
	ts       map[int]int
	overflow int
}

type O struct {
	T     int        `xml:"t,attr"`
	Attrs []xml.Attr `xml:",any,attr"`
	Os    []O        `xml:"o"`
}

var ocache = map[reflect.Type]*odecoder{}

type complaint struct {
	st reflect.Type
	ot int
}

var complainOnce = map[complaint]bool{}

func complain(st reflect.Type, ot int) {
	c := complaint{st, ot}
	if complainOnce[c] {
		return
	}
	complainOnce[c] = true
	log.Printf("struct %v should hande ot: %d\n", st, ot)
}

func (o *O) Decode(data interface{}) {
	v := reflect.Indirect(reflect.ValueOf(data))
	t := v.Type()
	dec := ocache[t]
	if dec == nil {
		dec = &odecoder{fields: make(map[string]odec), ts: make(map[int]int), overflow: -1}
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag.Get("x3t")
			tp := tagParse(tag)
			if field := tp["o"]; field != "" {
				dec.fields[field] = odec{i, t.Field(i).Type.Kind()}
			}
			if tp["os"] != "" {
				dec.overflow = i
			}
			if ot := tp["ot"]; ot != "" {
				typ, err := strconv.Atoi(ot)
				if err != nil {
					log.Fatal(err)
				}
				dec.ts[typ] = i
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
	for i := range o.Os {
		if f, ok := dec.ts[o.Os[i].T]; ok {
			field := v.Field(f)
			field = reflect.Append(field, reflect.Zero(field.Type().Elem()))
			v.Field(f).Set(field)
			o.Os[i].Decode(field.Index(field.Len() - 1).Addr().Interface())
		} else {
			complain(t, o.Os[i].T)
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
	uo.Decode(&u)

	return u
}
