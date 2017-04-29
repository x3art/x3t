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
	X      int `x3t:"o:x"`    // map coordinates
	Y      int `x3t:"o:y"`    // map coordinates
	R      int `x3t:"o:r"`    // Race
	Size   int `x3t:"o:size"` // Radius
	M      int `x3t:"o:m"`
	P      int `x3t:"o:p"`
	Qtrade int `x3t:"o:qtrade"`
	Qfight int `x3t:"o:qfight"`
	Qthink int `x3t:"o:qthink"`
	Qbuild int `x3t:"o:qbuild"`

	Suns       []Sun      `x3t:"ot:3"`
	Asteroids  []Asteroid `x3t:"ot:17"`
	Background Background `x3t:"ot:2"`
	Planets    []Planet   `x3t:"ot:4"`
	Docks      []Dock     `x3t:"ot:5"`
	Factories  []Factory  `x3t:"ot:6"`
	Gates      []Gate     `x3t:"ot:18"`
	Ships      []UShip    `x3t:"ot:7"`
	Specials   []Special  `x3t:"ot:20"`
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
	r, _ := text.Get(7, 1020000+100*(s.Y+1)+(s.X+1))
	return r
}

type pos struct {
	X int `x3t:"o:x"`
	Y int `x3t:"o:y"`
	Z int `x3t:"o:z"`
}

type rot struct {
	A int `x3t:"o:a"`
	B int `x3t:"o:b"`
	G int `x3t:"o:g"`
}

type Asteroid struct {
	Type   int `x3t:"o:atype"` // 0 - ore, 1 - silicon, 2 - nividium, 3 - ice
	Amount int `x3t:"o:aamount"`
	S      int `x3t:"o:s"`
	pos
	rot
	F int `x3t:"o:f"`
}

type Sun struct {
	S int `x3t:"o:s"`
	pos
	Color int `x3t:"o:color"`
	F     int `x3t:"o:f"`
}

type Background struct {
	S     int `x3t:"o:s"`
	Neb   int `x3t:"o:neb"`
	Stars int `x3t:"o:stars"`
}

type Planet struct {
	F int `x3t:"o:f"`
	S int `x3t:"o:s"`
	pos
	Color int `x3t:"o:color"`
	Fn    int `x3t:"o:fn"`
}

type race struct {
	R int `x3t:"o:r"`
}

type id struct {
	Id int `x3t:"o:id"`
}

type station struct {
	id
	F int `x3t:"o:f"`
	pos
	rot
	race
	CCs []CustomisableContainer `x3t:"ot:23"`
}

type Dock struct {
	S string `x3t:"o:s"`
	station
	N int `x3t:"o:n"`
}

type Factory struct {
	S string `x3t:"o:s"`
	station
}

type Gate struct {
	id
	Gid int `x3t:"o:gid"` // 0,1,2,3 - N,S,W,E
	pos
	rot
	// 0, 1, 2, 3 - normal gate N, S, W, E
	// 4 - destroyed
	// 5, 6, 7, 8 - Trans-orbital accelerator N, S, W, E
	// 9 - 20 - xenon hub gates
	S    int `x3t:"o:s"`
	Gx   int `x3t:"o:gx"`
	Gy   int `x3t:"o:gy"`
	Gtid int `x3t:"o:gtid"`
	F    int `x3t:"o:f"`
}

type UShip struct { // Name conflict, sigh.
	id
	S string `x3t:"o:s"`
	F int    `x3t:"o:f"`
	pos
	race
}

type Special struct {
	id
	S string `x3t:"o:s"`
	pos
	rot
	V int `x3t:"o:v"`
}

type CustomisableContainer struct {
	S        int    `x3t:"o:s"`
	Lasers   []Ware `x3t:"ot:8"`
	Shields  []Ware `x3t:"ot:9"`
	Missiles []Ware `x3t:"ot:10"`
	Energy   []Ware `x3t:"ot:11"`
	Novelty  []Ware `x3t:"ot:12"`
	Bio      []Ware `x3t:"ot:13"`
	Food     []Ware `x3t:"ot:14"`
	Mineral  []Ware `x3t:"ot:15"`
	Tech     []Ware `x3t:"ot:16"`
}

type Ware struct {
	id
	F int    `x3t:"o:f"`
	S string `x3t:"o:s"`
	I int    `x3t:"o:i"`
	pos
	N int `x3t:"o:n"`
}

type Universe struct {
	Sectors []Sector `x3t:"ot:1"`
}

func t(attrs []xml.Attr) int {
	for i := range attrs {
		if attrs[i].Name.Local == "t" {
			i, err := strconv.Atoi(attrs[i].Value)
			if err != nil {
				log.Fatal(err)
			}
			return i
		}
	}
	log.Fatal("no t")
	return -1
}

type odec struct {
	i []int
	k reflect.Kind
}

type odecoder struct {
	fields map[string]odec
	ts     map[int][]int

	complaints map[int]bool // to avoid spamming the output, just complain about each missing t once.
}

func (dec *odecoder) complain(st reflect.Type, ot int) {
	if dec.complaints[ot] {
		return
	}
	dec.complaints[ot] = true
	log.Printf("struct %v should hande ot: %d\n", st, ot)
}

func (dec *odecoder) embed(t reflect.Type, index []int) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			dec.embed(field.Type, append(index, i))
			continue
		}
		tp := tagParse(field.Tag.Get("x3t"))
		if ofield := tp["o"]; ofield != "" {
			dec.fields[ofield] = odec{append(index, i), field.Type.Kind()}
		}
		if ot := tp["ot"]; ot != "" {
			typ, err := strconv.Atoi(ot)
			if err != nil {
				log.Fatal(err)
			}
			dec.ts[typ] = append(index, i)
		}
	}
}

var ocache = map[reflect.Type]*odecoder{}

func decoder(t reflect.Type) *odecoder {
	dec := ocache[t]
	if dec == nil {
		dec = &odecoder{fields: make(map[string]odec), ts: make(map[int][]int)}
		dec.embed(t, []int{})
		ocache[t] = dec
	}
	return dec
}

func (dec *odecoder) attrs(v reflect.Value, attrs []xml.Attr) {
	for a := range attrs {
		attr := &attrs[a]
		if d, ok := dec.fields[attr.Name.Local]; ok {
			switch d.k {
			case reflect.String:
				v.FieldByIndex(d.i).SetString(attr.Value)
			case reflect.Int:
				i, err := strconv.Atoi(attr.Value)
				if err != nil {
					log.Fatalf("%v.%s: %v", v.Type(), attr.Name.Local, err)
				}
				v.FieldByIndex(d.i).SetInt(int64(i))
			default:
				log.Fatal("unknown field type")
			}
		} else if attr.Name.Local != "t" {
			log.Printf("unknown attr %v.%v: %v", v.Type(), attr.Name.Local, attr.Value)
		}
	}
}

func nextEl(d *xml.Decoder, n string) (bool, *xml.StartElement) {
	t, err := d.Token()
	if t == nil {
		return false, nil
	}
	if err != nil {
		log.Fatal(err)
	}
	switch t.(type) {
	case xml.StartElement:
		e := t.(xml.StartElement)
		if e.Name.Local != n {
			log.Fatalf("wrong start element: %v", t)
		}
		return true, &e
	case xml.EndElement:
		return false, nil
	case xml.ProcInst, xml.CharData, xml.Comment:
		return nextEl(d, n)
	default:
		log.Fatalf("unhandled token: %#v", t)
	}
	return false, nil
}

func elem(d *xml.Decoder, el *xml.StartElement, v reflect.Value) {
	dec := decoder(v.Type())
	dec.attrs(v, el.Attr)
	for {
		next, el := nextEl(d, "o")
		if !next {
			return
		}
		ot := t(el.Attr)
		if f, ok := dec.ts[ot]; ok {
			field := v.FieldByIndex(f)
			typ := field.Type()
			switch typ.Kind() {
			case reflect.Slice:
				field.Set(reflect.Append(field, reflect.Zero(typ.Elem())))
				elem(d, el, field.Index(field.Len()-1))
			case reflect.Struct:
				elem(d, el, field)
			default:
				log.Fatal("bad kind")
			}
		} else {
			dec.complain(v.Type(), ot)
		}
	}
}

func GetUniverse(n string) (u Universe) {
	f, err := os.Open(n)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	ok, el := nextEl(d, "universe")
	if !ok {
		log.Fatal("not universe")
	}

	elem(d, el, reflect.Indirect(reflect.ValueOf(&u)))

	return u
}

func (u Universe) SectorXY(x, y int) *Sector {
	// There's only a few hundred of them, no point in making this more efficient (for now).
	for i := range u.Sectors {
		if u.Sectors[i].X == x && u.Sectors[i].Y == y {
			return &u.Sectors[i]
		}
	}
	return nil
}
