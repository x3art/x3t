package bob

//go:generate go run ./gen/main.go . partX3

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"reflect"
	"strings"
)

/*
 * Whoever designed this "binary" format should take a hard look at
 * himself in the mirror. Mixing 32 bit and 16 bit array sizes and
 * special casing types we decode to by flags...
 *
 * We could almost use encoding/binary for this. If it weren't for the
 * bloody 0-terminated strings, they screw everything up.
 * Also, bufio would be nice, except that handling short reads from
 * bufio made things 3-4 times slower (why bufio gives us short reads
 * for 4 byte reads is...).
 *
 * This package is written with manual buffers and so many things
 * unrolled and not done generically because every single change from
 * the original generic/reflect approach has been carefully benchmarked
 * and going from 900ms to decode a mid-sized model to 30ms felt like
 * a good trade-off for the increased complexity of this code.
 *
 * There is still one big performance win to do here. We can know when a
 * slice element has a static size and use that to fetch data for the
 * whole element.
 */

// This keeps track of our reading. `buffer` is an internal buffer for
// future reads. `w` is a window into the buffer that keeps track of
// how much we've consumed.
type bobReader struct {
	buffer [4096]byte
	source io.Reader
	eof    bool
	w      []byte
}

var bobDec = tdec(reflect.TypeOf(Bob{}), 0)

type sTag [4]byte

func Read(r io.Reader) (*Bob, error) {
	br := &bobReader{source: r}

	b := &Bob{}
	err := br.sect(sTag{'B', 'O', 'B', '1'}, sTag{'/', 'B', 'O', 'B'}, false, func() error {
		return bobDec(br, b)
	})
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Data reader. We return a slice of an internal data buffer at least
// `l` bytes long. If the request amount is larger than the internal
// buffer the returned slice is allocated specifically for this
// request and doesn't use the buffer.
func (r *bobReader) data(l int, consume bool) ([]byte, error) {
	if len(r.w) < l {
		if l > len(r.buffer) {
			ret := make([]byte, l, l)
			copy(ret, r.w)
			resid := len(r.w)
			r.w = r.w[resid:]
			if resid == l {
				return ret, nil
			}
			n, err := r.source.Read(ret[resid:])
			if n != l-resid {
				err = io.EOF
			}
			if err != nil {
				if err == io.EOF {
					r.eof = true
				}
				return nil, err
			}
			return ret, nil
		}
		if r.eof {
			return nil, io.EOF
		}
		resid := len(r.w)
		if resid != 0 {
			copy(r.buffer[:], r.w)
		}
		n, err := r.source.Read(r.buffer[resid:])
		if err != nil {
			r.eof = err == io.EOF
			if r.eof && n+resid >= l {
				err = nil
			} else {
				return nil, err
			}
		}
		r.w = r.buffer[:n+resid]
	}
	ret := r.w
	if consume {
		r.eat(l)
	}
	return ret, nil
}

func (r *bobReader) eat(l int) {
	r.w = r.w[l:]
}

// The only time we peek at bytes forward is when sections are
// optional, but any time we don't find an optional section the next
// thing read will be either another section start or a section end.
func (r *bobReader) matchTag(expect sTag) (bool, error) {
	b, err := r.data(4, false)
	if err != nil {
		return false, err
	}
	match := b[0] == expect[0] && b[1] == expect[1] && b[2] == expect[2] && b[3] == expect[3]
	if match {
		r.eat(4)
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
				field := t.Field(i)
				if field.PkgPath != "" {
					continue
				}
				var sect, sectOptional bool
				var sectStart, sectEnd sTag
				for _, t := range strings.Split(field.Tag.Get("bobgen"), ",") {
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
				fdec := tdec(field.Type, nflags)
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
					if fields[i] == nil {
						continue
					}
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

func dec16(d []byte) int16 {
	_ = d[1]
	return int16(uint16(d[1]) | uint16(d[0])<<8)
}

func (r *bobReader) decode16() (int16, error) {
	d, err := r.data(2, true)
	if err != nil {
		return 0, err
	}
	return dec16(d), nil
}

func dec32(d []byte) int32 {
	_ = d[3]
	return int32(uint32(d[3]) | uint32(d[2])<<8 | uint32(d[1])<<16 | uint32(d[0])<<24)
}

func (r *bobReader) decode32() (int32, error) {
	d, err := r.data(4, true)
	if err != nil {
		return 0, err
	}
	return dec32(d), nil
}

func decf32(d []byte) float32 {
	return math.Float32frombits(uint32(d[3]) | uint32(d[2])<<8 | uint32(d[1])<<16 | uint32(d[0])<<24)
}

func (r *bobReader) decodef32() (float32, error) {
	d, err := r.data(4, true)
	if err != nil {
		return 0, err
	}
	return decf32(d), nil
}

func (r *bobReader) decodeString() (string, error) {
	b, err := r.data(1, false)
	if err != nil {
		return "", err
	}
	off := bytes.IndexByte(b, 0)
	if off != -1 {
		// trivial case
		s := string(b[:off])
		r.eat(off + 1)
		return s, nil
	}
	done := false
	ret := make([]byte, 0)
	for !done {
		b, err := r.data(1, false)
		if err != nil {
			return "", err
		}
		off := bytes.IndexByte(b, 0)
		if off != -1 {
			done = true
			ret = append(ret, b[:off]...)
			r.eat(off + 1)
		} else {
			ret = append(ret, b...)
			r.eat(len(b))
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

func (dec decd) decodeSlice(r *bobReader, v interface{}, l int) error {
	switch v := v.(type) {
	case *[]int16:
		*v = make([]int16, l, l)
		d, err := r.data(len(*v)*2, true)
		if err != nil {
			return err
		}
		for i := range *v {
			(*v)[i] = dec16(d[i*2:])
		}
		return nil
	case *[]int32:
		*v = make([]int32, l, l)
		d, err := r.data(len(*v)*4, true)
		if err != nil {
			return err
		}
		for i := range *v {
			(*v)[i] = dec32(d[i*4:])
		}
		return nil
	case *[]float32:
		*v = make([]float32, l, l)
		d, err := r.data(len(*v)*4, true)
		if err != nil {
			return err
		}
		for i := range *v {
			(*v)[i] = decf32(d[i*4:])
		}
		return nil
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

func decodeArray(r *bobReader, v interface{}) error {
	switch v := v.(type) {
	case *[10]int32:
		d, err := r.data(len(v)*4, true)
		if err != nil {
			return err
		}
		for i := range *v {
			(*v)[i] = dec32(d[i*4:])
		}
		return nil
	case *[4]int32:
		d, err := r.data(len(v)*4, true)
		if err != nil {
			return err
		}
		for i := range *v {
			(*v)[i] = dec32(d[i*4:])
		}
		return nil
	case *[6]float32:
		d, err := r.data(len(v)*4, true)
		if err != nil {
			return err
		}
		for i := range *v {
			(*v)[i] = decf32(d[i*4:])
		}
		return nil
	case *[3]int16:
		d, err := r.data(len(v)*2, true)
		if err != nil {
			return err
		}
		for i := range *v {
			(*v)[i] = dec16(d[i*2:])
		}
		return nil
	case *[2]int16:
		d, err := r.data(len(v)*2, true)
		if err != nil {
			return err
		}
		for i := range *v {
			(*v)[i] = dec16(d[i*2:])
		}
		return nil
	default:
		log.Fatalf("Special case array type  %T", v)
	}
	return nil
}

func decDecoder(r *bobReader, v interface{}) error {
	return v.(decoder).Decode(r)
}

type Bob struct {
	Info   string      `bobgen:"sect:INFO:/INF,optional"`
	Mat6   []material6 `bobgen:"sect:MAT6:/MAT,len32"`
	Bodies []body      `bobgen:"sect:BODY:/BOD"`
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
	d, err := r.data(sz*4, true)
	if err != nil {
		return err
	}
	for i := 0; i < sz; i++ {
		p.values[i] = dec32(d[i*4:])
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
	Faces         [][4]int32 `bobgen:"len32"`
}

type faceListX3 struct {
	MaterialIndex int32
	Faces         [][4]int32 `bobgen:"len32"`
	UVList        []uv       `bobgen:"len32"`
}

type partX3 struct {
	FacesX3 []faceListX3
	X3Vals  [10]int32
}

type partNotX3 struct {
	Faces []faceList
}

type part struct {
	Flags int32
	p     interface{}
}

var px3Dec = tdec(reflect.TypeOf(partX3{}), 0)
var pnx3Dec = tdec(reflect.TypeOf(partNotX3{}), 0)

func (p *part) Decode(r *bobReader) error {
	f, err := r.decode32()
	if err != nil {
		return err
	}
	p.Flags = f
	if (f & 0x10000000) != 0 {
		px := partX3{}
		err = px3Dec(r, &px)
		p.p = px
	} else {
		px := partNotX3{}
		err = pnx3Dec(r, &px)
		p.p = px
	}
	return err
}

type body struct {
	Size    int32
	Flags   int32
	Bones   []string `bobgen:"sect:BONE:/BON,len32,optional"`
	Points  []point  `bobgen:"sect:POIN:/POI,len32,optional"`
	Weights []weight `bobgen:"sect:WEIG:/WEI,len32,optional"`
	Parts   []part   `bobgen:"sect:PART:/PAR,len32,optional"`
}
