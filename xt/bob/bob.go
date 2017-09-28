package bob

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"reflect"
	"strings"
	"time"
)

/*
 * Whoever designed this "binary" format should take a hard look at
 * himself in the mirror. Mixing 32 bit and 16 bit array sizes and
 * special casing types we decode to by flags...
 */

func Read(r io.Reader) {
	br := bufio.NewReader(r)
	b := Bob{}
	t := time.Now()
	err := sect(br, "BOB1", "/BOB", false, func() error { return decodeVal(br, &b) })
	fmt.Printf("T: %v\n", time.Since(t))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", b.Info)
	return
}

func sect(r *bufio.Reader, s, e string, optional bool, f func() error) error {
	hdr, err := r.Peek(4)
	if err != nil {
		return err
	}
	if string(hdr) != s {
		if optional {
			return nil
		}
		return fmt.Errorf("unexpected [%s], expected [%s]", hdr, s)
	} else {
		_, err = r.Read(hdr)
		if err != nil {
			return err
		}
	}
	err = f()
	if err != nil {
		return err
	}
	_, err = r.Read(hdr)
	if string(hdr) != e {
		return fmt.Errorf("unexpected [%s]%v, expected [%s]", hdr, hdr, e)
	}
	return nil
}

func decodeVal(r *bufio.Reader, data interface{}) error {
	return tinfo(reflect.TypeOf(data).Elem(), 0).decodev(r, data)
}

const (
	len32 = uint(1 << iota)
)

type decoder interface {
	Decode(*bufio.Reader) error
}

type typeInfo struct {
	flags  uint
	kind   reflect.Kind
	slTi   *typeInfo
	fields []fieldSpecial
}

type fieldSpecial struct {
	sectStart    string
	sectEnd      string
	sectOptional bool
	ti           *typeInfo
}

var fsCache = map[reflect.Type]*typeInfo{}

func tinfo(t reflect.Type, flags uint) *typeInfo {
	if c, ok := fsCache[t]; ok {
		return c
	}
	ret := &typeInfo{flags: flags}
	ret.kind = t.Kind()
	switch ret.kind {
	case reflect.Struct:
		n := t.NumField()
		ret.fields = make([]fieldSpecial, n)
		for i := 0; i < n; i++ {
			flags := uint(0)
			for _, t := range strings.Split(t.Field(i).Tag.Get("x3t"), ",") {
				if t == "len32" {
					flags |= len32
				} else if strings.HasPrefix(t, "sect") {
					x := strings.Split(t, ":")
					if len(x) != 3 {
						panic(fmt.Errorf("sect tag bad: [%s]", t))
					}
					ret.fields[i].sectStart = x[1]
					ret.fields[i].sectEnd = x[2]
				} else if t == "optional" {
					ret.fields[i].sectOptional = true
				}
			}
			ret.fields[i].ti = tinfo(t.Field(i).Type, flags)
		}
	case reflect.Slice, reflect.Array:
		ret.slTi = tinfo(t.Elem(), 0)
	}
	fsCache[t] = ret
	return ret
}

func decode16(r *bufio.Reader) (int16, error) {
	d := make([]byte, 2, 2)
	_, err := io.ReadFull(r, d)
	if err != nil {
		return 0, err
	}
	return int16(uint16(d[1]) | uint16(d[0])<<8), nil
}

func decode32(r *bufio.Reader) (int32, error) {
	d := make([]byte, 4)
	_, err := io.ReadFull(r, d)
	if err != nil {
		return 0, err
	}
	return int32(uint32(d[3]) | uint32(d[2])<<8 | uint32(d[1])<<16 | uint32(d[0])<<24), nil
}

func (ti *typeInfo) decodev(r *bufio.Reader, v interface{}) error {
	// Pretty simple, integer types are the right size and big
	// endian, strings are nul-terminated. no alignment
	// considerations.

	if dc, ok := v.(decoder); ok {
		return dc.Decode(r)
	}

	var err error

	switch ti.kind {
	case reflect.Slice:
		var sliceLen int
		if (ti.flags & len32) != 0 {
			x, err := decode32(r)
			if err != nil {
				return err
			}
			sliceLen = int(x)
		} else {
			x, err := decode16(r)
			if err != nil {
				return err
			}
			sliceLen = int(x)
		}
		switch v := v.(type) {
		case *[]int16:
			*v = make([]int16, sliceLen, sliceLen)
			for i := range *v {
				(*v)[i], err = decode16(r)
			}
		case *[]int32:
			*v = make([]int32, sliceLen, sliceLen)
			for i := range *v {
				(*v)[i], err = decode32(r)
			}
		default:
			val := reflect.Indirect(reflect.ValueOf(v))
			val.Set(reflect.MakeSlice(val.Type(), sliceLen, sliceLen))
			for i := 0; i < sliceLen; i++ {
				err = ti.slTi.decodev(r, val.Index(i).Addr().Interface())
			}
		}
		return err
	case reflect.Struct:
		val := reflect.Indirect(reflect.ValueOf(v))
		for i := 0; i < val.NumField(); i++ {
			f := &ti.fields[i]
			if f.sectStart != "" {
				err := sect(r, f.sectStart, f.sectEnd, f.sectOptional, func() error {
					return f.ti.decodev(r, val.Field(i).Addr().Interface())
				})
				if err != nil {
					return err
				}
			} else {
				err := f.ti.decodev(r, val.Field(i).Addr().Interface())
				if err != nil {
					return err
				}
			}
		}
		return nil
	case reflect.Array:
		val := reflect.Indirect(reflect.ValueOf(v))
		// XXX - should we just slice it and let this function deal with it?
		for i := 0; i < val.Len(); i++ {
			err = ti.slTi.decodev(r, val.Index(i).Addr().Interface())
		}
		return err
	}

	switch v := v.(type) {
	case *int16:
		*v, err = decode16(r)
	case *int32:
		*v, err = decode32(r)
	case *string:
		s, err := r.ReadBytes(0)
		if err != nil {
			return err
		}
		*v = string(s)
	case *float32:
		var x int32
		x, err = decode32(r)
		*v = math.Float32frombits(uint32(x))
	default:
		panic("unknown type")
	}
	return err
}

type Bob struct {
	Info   string      `x3t:"sect:INFO:/INF,optional"`
	Mat6   []material6 `x3t:"sect:MAT6:/MAT,len32"`
	Bodies []body      `x3t:"sect:BODY:/BOD"`
}

type mat6Value struct {
	Hdr struct {
		Name string
		Type int16
	}
	b  int32
	i  int32
	f  float32
	f4 [4]float32
	s  string
}

func (m *mat6Value) Decode(r *bufio.Reader) error {
	err := decodeVal(r, &m.Hdr)
	if err != nil {
		return err
	}
	var data interface{}
	// XXX - make constants, not magic numbers here.
	switch m.Hdr.Type {
	case 0:
		data = &m.i
	case 1:
		data = &m.b
	case 2:
		data = &m.f
	case 5:
		data = &m.f4
	case 8:
		data = &m.s
	default:
		return fmt.Errorf("unknown mat6 type %x", m.Hdr.Type)
	}
	return decodeVal(r, data)
}

type Mat6Pair struct {
	Name  string
	Value int16
}

type mat6big struct {
	Technique int16
	Effect    string
	Value     []mat6Value
}

type mat6small struct {
	TextureFile                string
	Ambient, Diffuse, Specular [3]int16
	Transparency               int32
	SelfIllumination           int16
	Shininess                  [2]int16
	TextureValue               int16
	EnvironmentMap             Mat6Pair
	BumpMap                    Mat6Pair
	LightMap                   Mat6Pair
	Map4                       Mat6Pair
	Map5                       Mat6Pair
}

const matFlagBig = 0x2000000

type material6 struct {
	matHdr struct {
		Index int16
		Flags int32
	}
	mat interface{}
}

func (m *material6) Decode(r *bufio.Reader) error {
	err := decodeVal(r, &m.matHdr)
	if err != nil {
		return err
	}
	if m.matHdr.Flags == matFlagBig {
		m.mat = &mat6big{}
	} else {
		m.mat = &mat6small{}
	}
	return decodeVal(r, m.mat)
}

type point struct {
	hdr struct {
		Type int16
	}
	values []int32
}

func (p *point) Decode(r *bufio.Reader) error {
	err := decodeVal(r, &p.hdr)
	if err != nil {
		return err
	}
	sz := 0
	switch p.hdr.Type {
	case 0x1f:
		sz = 11
	case 0x1b:
		sz = 9
	case 0x19:
		sz = 7
	default:
		return fmt.Errorf("unknown point type %d", p.hdr.Type)
	}
	p.values = make([]int32, sz)
	for i := range p.values {
		err := decodeVal(r, &p.values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

type weight struct {
	Weights []struct {
		Idx   int16
		Coeff int32
	}
}

type uv struct {
	Idx    int32
	Values [6]float32
}

type faceList struct {
	MaterialIndex int32
	Faces         [][4]int32 `x3t:"len32"`
}

type faceListX3 struct {
	MaterialIndex int32
	Faces         [][4]int32 `x3t:"len32"`
	UVList        []uv       `x3t:"len32"`
}

type part struct {
	Hdr struct {
		Flags int32
	}
	x3 struct {
		FacesX3 []faceListX3
		X3Vals  [10]int32
	}
	notx3 struct {
		Faces []faceList
	}
}

func (p *part) Decode(r *bufio.Reader) error {
	err := decodeVal(r, &p.Hdr)
	if err != nil {
		return err
	}
	if (p.Hdr.Flags & 0x10000000) != 0 {
		err = decodeVal(r, &p.x3)
	} else {
		err = decodeVal(r, &p.notx3)
	}
	return err
}

type body struct {
	Size    int32
	Flags   int32
	Bones   []string `x3t:"sect:BONE:/BON,len32,optional"`
	Points  []point  `x3t:"sect:POIN:/POI,len32,optional"`
	Weights []weight `x3t:"sect:WEIG:/WEI,len32,optional"`
	Parts   []part   `x3t:"sect:PART:/PAR,len32,optional"`
}
