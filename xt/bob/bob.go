package bob

import (
	"bytes"
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
 *
 * We could almost use encoding/binary for this. If it weren't for the
 * bloody 0-terminated strings, they screw everything up.
 * Also, bufio would be nice, except that handling short reads from
 * bufio made things 3-4 time slower (why bufio gives us short reads
 * for 4 byte reads is...).
 */

type sTag [4]byte

type bobReader struct {
	source io.Reader
	buffer [4096]byte
	w      []byte
	eof    bool
}

var bobDec = tdec(reflect.TypeOf(Bob{}), 0)

func Read(r io.Reader) {
	br := &bobReader{source: r}

	b := Bob{}
	t := time.Now()
	err := br.sect(sTag{'B', 'O', 'B', '1'}, sTag{'/', 'B', 'O', 'B'}, false, func() error {
		return bobDec(br, &b)
	})
	fmt.Printf("T: %v\n", time.Since(t))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", b.Info)
	return
}

// Ensure that there are at least l bytes in the buffer.
func (r *bobReader) ensure(l int) error {
	if len(r.w) >= l {
		return nil
	}
	if r.eof {
		return io.EOF
	}
	resid := len(r.w)
	if resid != 0 {
		copy(r.buffer[:resid], r.w)
	}
	n, err := r.source.Read(r.buffer[resid:])
	if err == io.EOF {
		r.eof = true
		if n+resid < l {
			return io.EOF
		}
		err = nil
	}
	if err != nil {
		return err
	}
	r.w = r.buffer[:n+resid]
	return nil
}

// The only time we peek at bytes forward is when sections are
// optional, but any time we don't find an optional section the next
// thing read will be either another section start or a section end.
func (r *bobReader) matchTag(expect sTag) (bool, error) {
	err := r.ensure(4)
	if err != nil {
		return false, err
	}
	match := r.w[0] == expect[0] && r.w[1] == expect[1] && r.w[2] == expect[2] && r.w[3] == expect[3]
	if match {
		r.w = r.w[4:]
	}
	return match, nil
}

func (r *bobReader) sect(s, e sTag, optional bool, f func() error) error {
	match, err := r.matchTag(s)
	if err != nil {
		return err
	}
	if !match {
		if optional {
			return nil
		}
		return fmt.Errorf("unexpected [%s], expected [%s]", r.w[:4], s)
	}
	err = f()
	if err != nil {
		return err
	}
	match, err = r.matchTag(e)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("unexpected [%s]%v, expected [%s]", r.w[:4], r.w[:4], e)
	}
	return nil
}

const (
	len32 = uint(1 << iota)
)

type decoder interface {
	Decode(*bobReader) error
}

type decd func(*bobReader, interface{}) error

type typeInfo struct {
	dec decd
}

var fsCache = map[reflect.Type]decd{}

func tdec(t reflect.Type, flags uint) decd {
	if c, ok := fsCache[t]; ok {
		return c
	}
	var ret decd
	if reflect.PtrTo(t).Implements(reflect.TypeOf((*decoder)(nil)).Elem()) {
		ret = decDecoder
	} else {
		switch t.Kind() {
		case reflect.Struct:
			n := t.NumField()
			fields := make([]decd, n)
			for i := 0; i < n; i++ {
				nflags := uint(0)
				var sect, sectOptional bool
				var sectStart, sectEnd sTag
				for _, t := range strings.Split(t.Field(i).Tag.Get("x3t"), ",") {
					if t == "len32" {
						nflags |= len32
					} else if strings.HasPrefix(t, "sect") {
						x := strings.Split(t, ":")
						if len(x) != 3 {
							panic(fmt.Errorf("sect tag bad: [%s]", t))
						}
						sect = true
						copy(sectStart[:], x[1])
						copy(sectEnd[:], x[2])
					} else if t == "optional" {
						sectOptional = true
					}
				}
				fdec := tdec(t.Field(i).Type, nflags)
				if sect {
					fields[i] = func(r *bobReader, v interface{}) error {
						return r.sect(sectStart, sectEnd, sectOptional, func() error {
							return fdec(r, v)
						})
					}
				} else {
					fields[i] = fdec
				}
			}
			ret = func(r *bobReader, v interface{}) error {
				val := reflect.Indirect(reflect.ValueOf(v))
				for i := range fields {
					err := fields[i](r, val.Field(i).Addr().Interface())
					if err != nil {
						return err
					}
				}
				return nil
			}
		case reflect.Slice:
			slDec := tdec(t.Elem(), 0)
			if (flags & len32) != 0 {
				ret = slDec.decodeSlice32
			} else {
				ret = slDec.decodeSlice16
			}
		case reflect.Array:
			ret = decodeArray
		case reflect.Int16:
			ret = func(r *bobReader, v interface{}) (err error) {
				*(v.(*int16)), err = r.decode16()
				return
			}
		case reflect.Int32:
			ret = func(r *bobReader, v interface{}) (err error) {
				*(v.(*int32)), err = r.decode32()
				return
			}
		case reflect.String:
			ret = func(r *bobReader, v interface{}) (err error) {
				*(v.(*string)), err = r.decodeString()
				return
			}
		case reflect.Float32:
			ret = func(r *bobReader, v interface{}) (err error) {
				*(v.(*float32)), err = r.decodef32()
				return
			}
		default:
			log.Fatalf("unknown type %s", t.Name())
		}
	}
	fsCache[t] = ret
	return ret
}

func (r *bobReader) decode16() (int16, error) {
	err := r.ensure(2)
	if err != nil {
		return 0, err
	}
	_ = r.w[1]
	ret := int16(uint16(r.w[1]) | uint16(r.w[0])<<8)
	r.w = r.w[2:]
	return ret, nil
}

func (r *bobReader) decode32() (int32, error) {
	err := r.ensure(4)
	if err != nil {
		return 0, err
	}
	_ = r.w[3]
	ret := int32(uint32(r.w[3]) | uint32(r.w[2])<<8 | uint32(r.w[1])<<16 | uint32(r.w[0])<<24)
	r.w = r.w[4:]
	return ret, nil
}

func (r *bobReader) decodef32() (float32, error) {
	x, err := r.decode32()
	return math.Float32frombits(uint32(x)), err
}

func (r *bobReader) decodeString() (string, error) {
	err := r.ensure(1)
	if err != nil {
		return "", err
	}
	off := bytes.IndexByte(r.w, 0)
	if off != -1 {
		// trivial case
		s := string(r.w[:off])
		r.w = r.w[off+1:]
		return s, nil
	}
	done := false
	ret := make([]byte, 0)
	for !done {
		err := r.ensure(1)
		if err != nil {
			return "", err
		}
		off := bytes.IndexByte(r.w, 0)
		if off != -1 {
			done = true
			ret = append(ret, r.w[:off]...)
			r.w = r.w[off+1:]
		} else {
			ret = append(ret, r.w...)
			r.w = r.w[len(r.w):]
		}
	}
	return string(ret), nil
}

func (dec decd) decodeSlice32(r *bobReader, v interface{}) error {
	l, err := r.decode32()
	if err != nil {
		return err
	}
	return dec.decodeSlice(r, v, int(l))
}

func (dec decd) decodeSlice16(r *bobReader, v interface{}) error {
	l, err := r.decode16()
	if err != nil {
		return err
	}
	return dec.decodeSlice(r, v, int(l))
}

func (dec decd) decodeSlice(r *bobReader, v interface{}, l int) (err error) {
	switch v := v.(type) {
	case *[]int16:
		*v = make([]int16, l, l)
		for i := range *v {
			(*v)[i], err = r.decode16()
		}
		return
	case *[]int32:
		*v = make([]int32, l, l)
		for i := range *v {
			(*v)[i], err = r.decode32()
		}
		return
	case *[]float32:
		*v = make([]float32, l, l)
		for i := range *v {
			(*v)[i], err = r.decodef32()
		}
		return
	default:
		val := reflect.Indirect(reflect.ValueOf(v))
		val.Set(reflect.MakeSlice(val.Type(), l, l))
		for i := 0; i < l; i++ {
			err := dec(r, val.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func decodeArray(r *bobReader, v interface{}) (err error) {
	switch v := v.(type) {
	case *[10]int32:
		for i := range *v {
			(*v)[i], err = r.decode32()
		}
		return
	case *[4]int32:
		for i := range *v {
			(*v)[i], err = r.decode32()
		}
		return
	case *[6]float32:
		for i := range *v {
			(*v)[i], err = r.decodef32()
		}
		return
	default:
		log.Fatalf("Special case array type  %T", v)
	}
	return nil
}

func decDecoder(r *bobReader, v interface{}) error {
	return v.(decoder).Decode(r)
}

type Bob struct {
	Info   string      `x3t:"sect:INFO:/INF,optional"`
	Mat6   []material6 `x3t:"sect:MAT6:/MAT,len32"`
	Bodies []body      `x3t:"sect:BODY:/BOD"`
}

type mat6Value struct {
	Name string
	Type int16
	b    int32
	i    int32
	f    float32
	f4   [4]float32
	s    string
}

func (m *mat6Value) Decode(r *bobReader) error {
	var err error
	m.Name, _ = r.decodeString()
	m.Type, err = r.decode16()
	if err != nil {
		return err
	}
	// XXX - make constants, not magic numbers here.
	switch m.Type {
	case 0:
		m.i, err = r.decode32()
	case 1:
		m.b, err = r.decode32()
	case 2:
		m.f, err = r.decodef32()
	case 5:
		for i := range m.f4 {
			m.f4[i], err = r.decodef32()
		}
	case 8:
		m.s, err = r.decodeString()
	default:
		return fmt.Errorf("unknown mat6 type %x", m.Type)
	}
	return err
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
	Index int16
	Flags int32
	mat   interface{}
}

var m6bigDec = tdec(reflect.TypeOf(mat6big{}), 0)
var m6smallDec = tdec(reflect.TypeOf(mat6small{}), 0)

func (m *material6) Decode(r *bobReader) error {
	var err error
	m.Index, _ = r.decode16()
	m.Flags, err = r.decode32()
	if err != nil {
		return err
	}
	if m.Flags == matFlagBig {
		m.mat = &mat6big{}
		return m6bigDec(r, m.mat)
	} else {
		m.mat = &mat6small{}
		return m6smallDec(r, m.mat)
	}
}

type point struct {
	typ    int16
	values [11]int32
}

func (p *point) Decode(r *bobReader) error {
	t, err := r.decode16()
	if err != nil {
		return err
	}
	p.typ = t
	sz := 0
	switch p.typ {
	case 0x1f:
		sz = 11
	case 0x1b:
		sz = 9
	case 0x19:
		sz = 7
	default:
		return fmt.Errorf("unknown point type %d", p.typ)
	}
	for i := 0; i < sz; i++ {
		p.values[i], err = r.decode32()
	}
	return err
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

type partX3 struct {
	FacesX3 []faceListX3
	X3Vals  [10]int32
}

type partNotX3 struct {
	Faces []faceList
}

type part struct {
	flags int32
	x3    partX3
	notx3 partNotX3
}

var px3Dec = tdec(reflect.TypeOf(partX3{}), 0)
var pnx3Dec = tdec(reflect.TypeOf(partNotX3{}), 0)

func (p *part) Decode(r *bobReader) error {
	f, err := r.decode32()
	if err != nil {
		return err
	}
	p.flags = f
	if (p.flags & 0x10000000) != 0 {
		err = px3Dec(r, &p.x3)
	} else {
		err = pnx3Dec(r, &p.notx3)
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
