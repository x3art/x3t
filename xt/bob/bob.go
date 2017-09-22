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
	b := bob{}
	err := sect(br, "BOB1", "/BOB", false, func() error { return decodeVal(br, &b) })
	fmt.Printf("%v\n", b)
	if err != nil {
		log.Fatal(err)
	}
	return
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
		return fmt.Errorf("unexpected [%s], expected [%s]", hdr, e)
	}
	return nil
}

type bob struct {
	Info info
	Mat6 mat6
	Body body
}

type info string

func (i *info) Decode(r *bufio.Reader) error {
	return sect(r, "INFO", "/INF", true, func() error { return decodeValRaw(r, i) })
}

func decodeVal(r *bufio.Reader, data interface{}) error {
	return decode(r, false, reflect.Indirect(reflect.ValueOf(data)))
}

func decodeValRaw(r *bufio.Reader, data interface{}) error {
	return decode(r, true, reflect.Indirect(reflect.ValueOf(data)))
}

func decode(r *bufio.Reader, skipMethod bool, v reflect.Value) error {
	// Pretty simple, integer types are the right size and big
	// endian, strings are nul-terminated. no alignment
	// considerations.

	if !skipMethod && v.CanAddr() {
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
			err := decode(r, false, v.Field(i))
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
			err := decode(r, false, v.Index(i))
			if err != nil {
				return err
			}
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			err := decode(r, false, v.Index(i))
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

func (m *mat6) Decode(r *bufio.Reader) error {
	return sect(r, "MAT6", "/MAT", false, func() error {
		var count int32
		err := decodeVal(r, &count)
		if err != nil {
			return err
		}
		for i := 0; i < int(count); i++ {
			m := material6{}
			err := decodeVal(r, &m)
			if err != nil {
				return err
			}
			fmt.Printf("mat6: %v %v\n", m.matHdr, m.mat)
		}
		return nil
	})
}

type body struct {
	/*
		bones // bone
		points // POINT section?
		parts  // slice part
		weights // slice WEIGHT sections?
	*/
}

func (b *body) Decode(r *bufio.Reader) error {
	return sect(r, "BODY", "/BOD", false, func() error { return decodeValRaw(r, b) })
}
