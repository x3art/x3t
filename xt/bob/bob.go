package bob

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"reflect"
)

// Current understanding of the BOB format:
// Each section has an opening and a closing tag.
// The whole file is a section that starts with BOB1 and ends with \BOB

func Read(r io.Reader) {
	br := bufio.NewReader(r)
	err := sAny(br)
	if err != nil {
		log.Fatal(err)
	}
}

func sectLookup(s string) (func(*bufio.Reader) error, string) {
	// This is a function with a large switch instead of a map
	// because the compiler is silly and thinks there are circular dependencies.
	switch s {
	case "BOB1":
		return sBob, "/BOB"
	case "CUT1":
		return nil, "/CU1"
	case "INFO":
		return sInfo, "/INF"
	case "PATH":
		return nil, "/PAT"
	case "NAME":
		return nil, "/NAM"
	case "STAT":
		return nil, "/STA"
	case "NOTE":
		return nil, "/NOT"
	case "CONS":
		return nil, "/CON"
	case "MAT6":
		return sMat6, "/MAT"
	case "MAT5":
		return nil, "/MAT"
	case "BODY":
		return nil, "/BOD"
	case "POIN":
		return nil, "/POI"
	case "PART":
		return nil, "/PAR"
	case "BONE":
		return nil, "/BON"
	case "WEIG":
		return nil, "/WEI"
	}
	log.Fatalf("no reader for section %s", s)
	return nil, ""
}

func sect(r *bufio.Reader, s, e string, optional bool, f func() error) error {
	hdr := make([]byte, 4, 4)
	_, err := r.Read(hdr)
	if err != nil {
		return err
	}
	if string(hdr) != s {
		if optional {
			return nil
		}
		return fmt.Errorf("unexpected [%s], expected [%s]", hdr, s)
	}
	err = f()
	if err != nil {
		return err
	}
	_, err = r.Read(hdr)
	if string(hdr) != e {
		return fmt.Errorf("unexpected [%s], expected [%s]", hdr, s)
	}
	return nil
}

func sAny(r *bufio.Reader) error {
	s, err := r.Peek(4)
	if err != nil {
		return err
	}
	f, e := sectLookup(string(s))
	if f == nil {
		return fmt.Errorf("reader for %s not implemented", s)
	}
	return sect(r, string(s), e, false, func() error { return f(r) })
}

type bob struct {
	Info info
	Mat6 mat6
}

func sBob(r *bufio.Reader) error {
	b := bob{}
	err := decodeVal(r, &b)
	fmt.Printf("%v\n", b)
	return err
}

type info string

func (i *info) Decode(r *bufio.Reader) error {
	s := ""
	err := sect(r, "INFO", "/INF", true, func() error { return decodeVal(r, &s) })
	*i = info(s)
	return err
}

func sInfo(r *bufio.Reader) error {
	s := ""
	err := decodeVal(r, &s)
	if err != nil {
		return err
	}
	fmt.Printf("Info: %s\n", s)
	return nil
}

func decodeVal(r *bufio.Reader, data interface{}) error {
	return decode(r, reflect.Indirect(reflect.ValueOf(data)))
}

func decode(r *bufio.Reader, v reflect.Value) error {
	// Pretty simple, integer types are the right size and big
	// endian, strings are nul-terminated. no alignment
	// considerations.

	if v.CanAddr() {
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
			err := decode(r, v.Field(i))
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
		var l int16
		err := decodeVal(r, &l)
		if err != nil {
			return err
		}
		v.Set(reflect.MakeSlice(v.Type(), int(l), int(l)))
		for i := 0; i < int(l); i++ {
			err := decode(r, v.Index(i))
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			err := decode(r, v.Index(i))
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

func sMat6Pair(r *bufio.Reader, name string) {
	s, err := r.ReadBytes(0)
	if err != nil {
		log.Fatal(err)
	}
	var x int16
	err = binary.Read(r, binary.BigEndian, &x)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("mat6pair: %s: %s - %d\n", name, s, x)
}

type Mat1RGB struct {
	R, G, B int16
}

type Mat1Pair struct {
	Value, Strength int16
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
		return fmt.Errorf("unknown mat6 type")
	}
	return decodeVal(r, data)
}

type mat6 struct{}

func (m *mat6) Decode(r *bufio.Reader) error {
	return sect(r, "MAT6", "/MAT", false, func() error { return sMat6(r) })
}

func sMat6(r *bufio.Reader) error {
	var count int32
	err := decodeVal(r, &count)
	if err != nil {
		return err
	}
	for i := 0; i < int(count); i++ {
		var matHdr struct {
			Index int16
			Flags int32
		}
		err := decodeVal(r, &matHdr)
		if err != nil {
			return err
		}

		fmt.Printf("MAT6: hdr: %v %x\n", matHdr.Index, matHdr.Flags)
		if matHdr.Flags == matFlagBig {
			var big struct {
				Technique int16
				Effect    string
				Value     []mat6Value
			}
			err := decodeVal(r, &big)
			if err != nil {
				return err
			}
			fmt.Printf("big: %v\n", big)
		} else {
			// XXX - untested, but implemented because of earlier misunderstanding.
			textureFile, err := r.ReadBytes(0)
			if err != nil {
				return err
			}
			fmt.Printf("textureFile: %s\n", textureFile)
			p, _ := r.Peek(50)
			fmt.Printf("peek: %s %v\n", p, p)

			var small struct {
				Ambient, Diffuse, Specular Mat1RGB
				Transparency               int32
				SelfIllumination           int16
				Shininess                  Mat1Pair
				TextureValue               int16
			}
			err = binary.Read(r, binary.BigEndian, &small)
			fmt.Printf("small: %v\n", small)
			sMat6Pair(r, "enviromentMap")
			sMat6Pair(r, "bumpMap")
			sMat6Pair(r, "lightMap")
			sMat6Pair(r, "map4")
			sMat6Pair(r, "map5")
		}
	}
	return nil
}

/*
type body struct {
	bones // bone
	points // POINT section?
	parts  // slice part
	weights // slice WEIGHT sections?
}

func sBody(r *bufio.Reader) error {
	bodies := []body{}
}
*/
