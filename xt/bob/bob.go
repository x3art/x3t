package bob

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"reflect"
)

/*
 * Whoever designed this "binary" format should take a hard look at
 * himself in the mirror. Mixing 32 bit and 16 bit array sizes and
 * special casing types we decode to by flags...
 */

func Read(r io.Reader) {
	br := bufio.NewReader(r)
	b := bob{}
	err := sect(br, "BOB1", "/BOB", false, func() error { return decodeVal(br, 0, &b) })
	//	fmt.Printf("%v\n", b)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func sect(r *bufio.Reader, s, e string, optional bool, f func() error) error {
	hdr, err := r.Peek(4)
	if err != nil {
		return err
	}
	if string(hdr) != s {
		if optional {
			log.Printf("unoptional(%s) %v/%s", s, hdr, hdr)
			return nil
		}
		return fmt.Errorf("unexpected [%s], expected [%s]", hdr, s)
	} else {
		_, err = r.Read(hdr)
		if err != nil {
			return err
		}
	}
	log.Printf("start %s", s)
	err = f()
	if err != nil {
		return err
	}
	_, err = r.Read(hdr)
	if string(hdr) != e {
		log.Print(r.Peek(100))
		return fmt.Errorf("unexpected [%s]%v, expected [%s]", hdr, hdr, e)
	}
	log.Printf("end %s", s)
	return nil
}

type bob struct {
	Info   info
	Mat6   mat6
	Bodies bodies
}

type info string

func (i *info) Decode(r *bufio.Reader) error {
	return sect(r, "INFO", "/INF", true, func() error { return decodeVal(r, skipMethod, i) })
}

func decodeVal(r *bufio.Reader, flags uint, data interface{}) error {
	return decode(r, flags, reflect.Indirect(reflect.ValueOf(data)))
}

const (
	skipMethod = uint(1 << iota)
	len32
)

func decode(r *bufio.Reader, flags uint, v reflect.Value) error {
	// Pretty simple, integer types are the right size and big
	// endian, strings are nul-terminated. no alignment
	// considerations.

	if (flags&skipMethod) == 0 && v.CanAddr() {
		if dec := v.Addr().MethodByName("Decode"); dec.IsValid() {
			ret := dec.Call([]reflect.Value{reflect.ValueOf(r)})
			if len(ret) != 1 {
				return fmt.Errorf("Decode bad ret: %v", ret)
			}
			r := ret[0].Interface()
			if r != nil {
				return r.(error)
			}
			return nil
		}
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			err := decode(r, 0, v.Field(i))
			if err != nil {
				return err
			}
		}
	case reflect.String:
		s, err := r.ReadBytes(0)
		if err != nil {
			return err
		}
		v.SetString(string(s[:len(s)-1]))
	case reflect.Slice:
		l := 0
		if (flags & len32) != 0 {
			var x int32
			err := decodeVal(r, 0, &x)
			if err != nil {
				return err
			}
			l = int(x)
		} else {
			var x int16
			err := decodeVal(r, 0, &x)
			if err != nil {
				return err
			}
			l = int(x)
		}
		v.Set(reflect.MakeSlice(v.Type(), l, l))
		for i := 0; i < l; i++ {
			err := decode(r, 0, v.Index(i))
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			err := decode(r, 0, v.Index(i))
			if err != nil {
				return err
			}
		}
	default:
		err := binary.Read(r, binary.BigEndian, v.Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

const matFlagBig = 0x2000000

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
	err := decodeVal(r, 0, &m.Hdr)
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
		return fmt.Errorf("unknown mat6 type")
	}
	return decodeVal(r, 0, data)
}

type Mat1RGB struct {
	R, G, B int16
}

type Mat1Pair struct {
	Value, Strength int16
}

type Mat6Pair struct {
	Name  string
	Value int16
}

type material6 struct {
	matHdr struct {
		Index int16
		Flags int32
	}
	mat interface{}
}

type mat6big struct {
	Technique int16
	Effect    string
	Value     []mat6Value
}

type mat6small struct {
	TextureFile                string
	Ambient, Diffuse, Specular Mat1RGB
	Transparency               int32
	SelfIllumination           int16
	Shininess                  Mat1Pair
	TextureValue               int16
	EnvironmentMap             Mat6Pair
	BumpMap                    Mat6Pair
	LightMap                   Mat6Pair
	Map4                       Mat6Pair
	Map5                       Mat6Pair
}

func (m *material6) Decode(r *bufio.Reader) error {
	err := decodeVal(r, 0, &m.matHdr)
	if err != nil {
		return err
	}
	if m.matHdr.Flags == matFlagBig {
		m.mat = &mat6big{}
	} else {
		m.mat = &mat6small{}
	}
	return decodeVal(r, 0, m.mat)
}

type mat6 []material6

func (m *mat6) Decode(r *bufio.Reader) error {
	return sect(r, "MAT6", "/MAT", false, func() error { return decodeVal(r, skipMethod|len32, m) })
}

type bones []string

func (b *bones) Decode(r *bufio.Reader) error {
	return sect(r, "BONE", "/BON", true, func() error { return decodeVal(r, skipMethod|len32, b) })
}

type point struct {
	hdr struct {
		Type int16
	}
	values []int32
}

func (p *point) Decode(r *bufio.Reader) error {
	err := decodeVal(r, 0, &p.hdr)
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
		err := decodeVal(r, 0, &p.values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

type points []point

func (p *points) Decode(r *bufio.Reader) error {
	return sect(r, "POIN", "/POI", true, func() error { return decodeVal(r, skipMethod|len32, p) })
}

type weight struct {
	Weights []struct {
		Idx   int16
		Coeff int32
	}
}

type weights []weight

func (w *weights) Decode(r *bufio.Reader) error {
	return sect(r, "WEIG", "/WEI", true, func() error { return decodeVal(r, skipMethod|len32, w) })
}

type uv struct {
	Idx    int32
	Values [6]float32
}

type uvlist []uv

func (u *uvlist) Decode(r *bufio.Reader) error {
	return decodeVal(r, skipMethod|len32, u)
}

type faces [][4]int32

func (f *faces) Decode(r *bufio.Reader) error {
	return decodeVal(r, skipMethod|len32, f)
}

type faceList struct {
	MaterialIndex int32
	Faces         faces
}

type faceListX3 struct {
	MaterialIndex int32
	Faces         faces
	UVList        uvlist
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
	err := decodeVal(r, 0, &p.Hdr)
	if err != nil {
		return err
	}
	if (p.Hdr.Flags & 0x10000000) != 0 {
		err = decodeVal(r, 0, &p.x3)
	} else {
		err = decodeVal(r, 0, &p.notx3)
	}
	return err
}

type parts []part

func (p *parts) Decode(r *bufio.Reader) error {
	return sect(r, "PART", "/PAR", true, func() error { return decodeVal(r, skipMethod|len32, p) })
}

type body struct {
	Size    int32
	Flags   int32
	Bones   bones
	Points  points
	Weights weights
	Parts   parts
	/*
		bones // bone
		points // POINT section?
		parts  // slice part
		weights // slice WEIGHT sections?
	*/
}

func (b *body) Decode(r *bufio.Reader) error {
	err := decodeVal(r, skipMethod, b)
	fmt.Printf("body %x\n", b.Size)
	return err
}

type bodies []body

func (b *bodies) Decode(r *bufio.Reader) error {
	return sect(r, "BODY", "/BOD", false, func() error { return decodeVal(r, skipMethod, b) })
}
