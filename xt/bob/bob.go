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

var bobTinfo = tinfo(reflect.TypeOf(Bob{}), 0)

func Read(r io.Reader) {
	br := &bobReader{source: r}

	b := Bob{}
	t := time.Now()
	err := br.sect(sTag{'B', 'O', 'B', '1'}, sTag{'/', 'B', 'O', 'B'}, false, func() error {
		return bobTinfo.decodev(br, &b)
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

type typeInfo struct {
	flags  uint
	kind   reflect.Kind
	slTi   *typeInfo
	fields []fieldSpecial
}

type fieldSpecial struct {
	sect         bool
	sectStart    sTag
	sectEnd      sTag
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
					ret.fields[i].sect = true
					copy(ret.fields[i].sectStart[:], x[1])
					copy(ret.fields[i].sectEnd[:], x[2])
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

func (r *bobReader) arrsl(v interface{}) error {
	var err error
	switch v := v.(type) {
	case []int16:
		for i := range v {
			v[i], err = r.decode16()
		}
	case []int32:
		for i := range v {
			v[i], err = r.decode32()
		}
	case []float32:
		for i := range v {
			v[i], err = r.decodef32()
		}
	default:
		log.Fatalf("unknown array slice %T", v)
	}
	return err
}

func (ti *typeInfo) decodev(r *bobReader, v interface{}) error {
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
			x, err := r.decode32()
			if err != nil {
				return err
			}
			sliceLen = int(x)
		} else {
			x, err := r.decode16()
			if err != nil {
				return err
			}
			sliceLen = int(x)
		}
		switch v := v.(type) {
		case *[]int16:
			*v = make([]int16, sliceLen, sliceLen)
			return r.arrsl(*v)
		case *[]int32:
			*v = make([]int32, sliceLen, sliceLen)
			return r.arrsl(*v)
		case *[]float32:
			*v = make([]float32, sliceLen, sliceLen)
			return r.arrsl(*v)
		default:
			val := reflect.Indirect(reflect.ValueOf(v))
			val.Set(reflect.MakeSlice(val.Type(), sliceLen, sliceLen))
			for i := 0; i < sliceLen; i++ {
				err := ti.slTi.decodev(r, val.Index(i).Addr().Interface())
				if err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.Struct:
		val := reflect.Indirect(reflect.ValueOf(v))
		for i := 0; i < val.NumField(); i++ {
			f := &ti.fields[i]
			if f.sect {
				err := r.sect(f.sectStart, f.sectEnd, f.sectOptional, func() error {
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
		switch v := v.(type) {
		case *[10]int32:
			return r.arrsl(v[:])
		case *[4]int32:
			return r.arrsl(v[:])
		case *[6]float32:
			return r.arrsl(v[:])
		default:
			log.Fatalf("Special case array type  %T", v)
		}
		/*
			// This is no longer necessary, it should have been special cased above.
			val := reflect.Indirect(reflect.ValueOf(v))
			return r.arrsl(val.Slice(0, val.Len()).Interface())
		*/
	}

	switch v := v.(type) {
	case *int16:
		*v, err = r.decode16()
	case *int32:
		*v, err = r.decode32()
	case *string:
		*v, err = r.decodeString()
	case *float32:
		var x int32
		x, err = r.decode32()
		*v = math.Float32frombits(uint32(x))
	default:
		log.Fatalf("unknown type %T", v)
	}
	return err
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
		err = r.arrsl(m.f4[:])
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

var m6bigTinfo = tinfo(reflect.TypeOf(mat6big{}), 0)
var m6smallTinfo = tinfo(reflect.TypeOf(mat6small{}), 0)

func (m *material6) Decode(r *bobReader) error {
	var err error
	m.Index, _ = r.decode16()
	m.Flags, err = r.decode32()
	if err != nil {
		return err
	}
	if m.Flags == matFlagBig {
		m.mat = &mat6big{}
		return m6bigTinfo.decodev(r, m.mat)
	} else {
		m.mat = &mat6small{}
		return m6smallTinfo.decodev(r, m.mat)
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
	return r.arrsl(p.values[:sz])
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

var px3Tinfo = tinfo(reflect.TypeOf(partX3{}), 0)
var pnx3Tinfo = tinfo(reflect.TypeOf(partNotX3{}), 0)

func (p *part) Decode(r *bobReader) error {
	f, err := r.decode32()
	if err != nil {
		return err
	}
	p.flags = f
	if (p.flags & 0x10000000) != 0 {
		err = px3Tinfo.decodev(r, &p.x3)
	} else {
		err = pnx3Tinfo.decodev(r, &p.notx3)
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
