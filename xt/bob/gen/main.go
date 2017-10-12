package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type typeInfo interface {
	// Size returns a non-zero size of this type if it can be
	// determined before decoding, otherwise it returns 0.
	Size() int
	// Decode generates the code needed to decode from buf into
	// dest. buf may only be used on types with non-zero Size.
	// nilErr specifies if the function we're in returns nil, err
	Decode(out *gen, dest, buf string, nilErr bool)
	// name of this type.
	Name() string

	// name of this type that can be used in a func name.
	Fname() string

	// If this needs to generate a func to decode this type, this
	// does it.
	Func(out *gen)
}

type gen struct {
	structs   map[string]*ast.StructType
	targets   map[string]typeInfo // contains types that need a decoder func generated.
	generated map[string]bool     // marks generated types so that we don't generate a func twice.

	indent int
	out    io.Writer
}

func (out *gen) o(f string, a ...interface{}) *gen {
	for i := 0; i < out.indent; i++ {
		fmt.Fprint(out.out, "\t")
	}
	fmt.Fprintf(out.out, f, a...)
	return out
}

func (out *gen) i(i int) *gen {
	out.indent += i
	return out
}

func (out *gen) errRet(addnil bool) *gen {
	out.o("if err != nil {\n").i(1)
	if addnil {
		out.o("return nil, err\n")
	} else {
		out.o("return err\n")
	}
	out.i(-1).o("}\n")
	return out
}

type nofunc struct{}

func (_ nofunc) Func(out *gen) {
}

type decBuf struct {
}

type typeI32 struct {
	nofunc
}

func (t typeI32) Size() int {
	return 4
}

func (t typeI32) Decode(out *gen, dest, buf string, nilErr bool) {
	out.o("%s = dec32(%s)\n", dest, buf)
}

func (t typeI32) Name() string {
	return "int32"
}

func (t typeI32) Fname() string {
	return "i32"
}

type typeI16 struct {
	nofunc
}

func (t typeI16) Size() int {
	return 2
}

func (t typeI16) Decode(out *gen, dest, buf string, nilErr bool) {
	out.o("%s = dec16(%s)\n", dest, buf)
}

func (t typeI16) Name() string {
	return "int16"
}

func (t typeI16) Fname() string {
	return "i16"
}

type typeF32 struct {
	nofunc
}

func (t typeF32) Size() int {
	return 4
}

func (t typeF32) Decode(out *gen, dest, buf string, nilErr bool) {
	out.o("%s = decf32(%s)\n", dest, buf)
}

func (t typeF32) Name() string {
	return "float32"
}

func (t typeF32) Fname() string {
	return "f32"
}

type typeSlice struct {
	el    typeInfo
	len32 bool
}

func (t typeSlice) Size() int {
	return 0
}

func (t typeSlice) Decode(out *gen, dest, buf string, nilErr bool) {
	out.targets[t.Name()] = t
	out.o("%s, err = dec_%s(r)\n", dest, t.Fname())
	out.errRet(nilErr)
}

func (t typeSlice) Name() string {
	return fmt.Sprintf("[]%s", t.el.Name())
}

func (t typeSlice) Fname() string {
	return "sl_" + t.el.Fname()
}

func (t typeSlice) Func(out *gen) {
	fname := "dec_" + t.Fname()
	out.o("\nfunc %s(r *bobReader) ([]%s, error) {\n", fname, t.el.Name()).i(1)
	out.o("var err error\n")
	if t.len32 {
		out.o("l32, err := r.decode32()\n")
		out.errRet(true)
		out.o("l := int(l32)\n")
	} else {
		out.o("l16, err := r.decode16()\n")
		out.errRet(true)
		out.o("l := int(l16)\n")
	}

	out.o("ret := make([]%s, l)\n", t.el.Name())

	elsz := t.el.Size()
	if elsz != 0 {
		out.o("b, err := r.data(l*%d, true)\n", elsz)
		out.errRet(true)
		// zero sized elements will not use buf.
	}
	ivar := "i_" + t.Fname()
	out.o("for %s := range ret {\n", ivar).i(1)
	t.el.Decode(out, fmt.Sprintf("ret[%s]", ivar), fmt.Sprintf("b[%s*%d:]", ivar, elsz), true)
	out.i(-1).o("}\n")
	out.o("return ret, nil\n")
	out.i(-1).o("}\n")
}

type typeArr struct {
	el typeInfo
	sz int
}

func (t typeArr) Size() int {
	return t.el.Size() * t.sz
}

func (t typeArr) Decode(out *gen, dest, buf string, nilErr bool) {
	ivar := "i_" + t.Fname()
	out.o("for %s := range %s {\n", ivar, dest).i(1)
	elsz := t.el.Size()
	if elsz == 0 {
		panic("no support for 0 size array elements yet")
	}
	t.el.Decode(out, fmt.Sprintf("%s[%s]", dest, ivar), fmt.Sprintf("%s[%s*%d:]", buf, ivar, elsz), nilErr)
	out.i(-1).o("}\n")
}

func (t typeArr) Name() string {
	return fmt.Sprintf("[%d]%s", t.sz, t.el.Name())
}

func (t typeArr) Fname() string {
	return fmt.Sprintf("arr%d_%s", t.sz, t.el.Name())
}

func (t typeArr) Func(_ *gen) {
}

type structField struct {
	name string
	t    typeInfo
	i    int
	tag  string
}

type typeStruct struct {
	name   string
	fields []structField
}

func (t typeStruct) Size() int {
	sz := 0
	for i := range t.fields {
		s := t.fields[i].t.Size()
		if s == 0 {
			return 0
		}
		sz += s
	}
	return sz
}

func (t typeStruct) Decode(out *gen, dest, buf string, nilErr bool) {
	out.targets[t.Name()] = t
	bufdecoder := t.Size() != 0
	if bufdecoder {
		out.o("%s.decodeBuf(%s)\n", dest, buf)
	} else {
		out.o("err = %s.Decode(r)\n", dest)
		out.errRet(nilErr)
	}
}

func (t typeStruct) Name() string {
	return t.name
}

func (t typeStruct) Fname() string {
	return t.name
}

func (t typeStruct) Func(out *gen) {
	bufdecoder := t.Size() != 0
	if bufdecoder {
		out.o("\nfunc (x *%s) decodeBuf(b []byte) {\n", t.name).i(1)
		off := 0
		for i := range t.fields {
			t.fields[i].t.Decode(out, "x."+t.fields[i].name, fmt.Sprintf("b[%d:]", off), false)
			off += t.fields[i].t.Size()
		}
	} else {
		out.o("\nfunc (x *%s) Decode(r *bobReader) error {\n", t.name).i(1)
		sizeGroups := []int{}
		sz := 0
		needbuf := false
		for i := range t.fields {
			s := t.fields[i].t.Size()
			if s == 0 {
				sizeGroups = append(sizeGroups, sz)
				sz = 0
			} else {
				needbuf = true
				sz += s
			}
		}
		sizeGroups = append(sizeGroups, sz)

		if needbuf {
			out.o("var buf []byte\n")
			out.o("var err error\n")
		}

		nextbuf := true // next non-zero field needs new buf.
		bufoff := 0
		for i := range t.fields {
			s := t.fields[i].t.Size()
			if s == 0 {
				sizeGroups = sizeGroups[1:]
				nextbuf = true
			} else if nextbuf {
				out.o("buf, err = r.data(%d, true)\n", sizeGroups[0])
				out.errRet(false)
				sizeGroups = sizeGroups[1:]
				nextbuf = false
			}
			t.fields[i].t.Decode(out, "x."+t.fields[i].name, fmt.Sprintf("buf[%d:]", bufoff), false)
			bufoff += t.fields[i].t.Size()
		}
		out.o("return nil\n")
	}
	out.i(-1).o("}\n")
}

type eo struct {
	sect, sectOptional bool
	len32              bool
	sectStart, sectEnd [4]byte
}

func exprOpts(tag string) (ret *eo) {
	for _, t := range strings.Split(tag, ",") {
		if ret == nil {
			ret = &eo{}
		}
		if t == "len32" {
			ret.len32 = true
		} else if strings.HasPrefix(t, "sect") {
			x := strings.Split(t, ":")
			if len(x) != 3 {
				panic(fmt.Errorf("sect tag bad: [%s]", t))
			}
			ret.sect = true
			copy(ret.sectStart[:], x[1])
			copy(ret.sectEnd[:], x[2])
		} else if t == "optional" {
			ret.sectOptional = true
		}
	}
	return
}

func (s *gen) resolveExpr(e ast.Expr, opts *eo) typeInfo {
	switch t := e.(type) {
	case *ast.Ident:
		switch t.Name {
		case "int16":
			return typeI16{}
		case "int32":
			return typeI32{}
		case "float32":
			return typeF32{}
		case "string":
			panic("not yet")
		default:
			return s.resolveStruct(t.Name)
		}
	case *ast.ArrayType:
		if t.Len == nil {
			len32 := opts != nil && opts.len32
			return typeSlice{s.resolveExpr(t.Elt, nil), len32}
		}
		ls := t.Len.(*ast.BasicLit).Value
		l, err := strconv.Atoi(ls)
		if err != nil {
			log.Fatal("bad array size: %s", ls)
		}
		return typeArr{s.resolveExpr(t.Elt, nil), l}
	default:
		log.Fatal("resolveExpr, unknown: %v", e)
	}
	return nil
}

func (s *gen) resolveStruct(t string) typeInfo {
	if x := s.targets[t]; x != nil {
		return x
	}
	st := s.structs[t]
	if st == nil {
		log.Fatal("unknown type name: %s", t)
	}
	ret := typeStruct{name: t}
	for i, f := range st.Fields.List {
		if f.Names[0] == nil {
			log.Fatal("field without name not supported yet")
		}
		// do something with: if f.Names[0].IsExported
		name := f.Names[0].String()
		tag := ""
		if f.Tag != nil {
			ts, err := strconv.Unquote(f.Tag.Value)
			if err != nil {
				log.Fatalf("unquote: %v", err)
			}
			tag, _ = reflect.StructTag(ts).Lookup("bobgen")
		}

		ret.fields = append(ret.fields, structField{
			name, s.resolveExpr(f.Type, exprOpts(tag)), i, tag,
		})

	}
	s.targets[t] = ret
	return ret
}

func (s *gen) parseFile(fs *token.FileSet, fname string) {
	f, err := parser.ParseFile(fs, fname, nil, 0)
	if err != nil {
		log.Fatalf("parser.ParseFile(%s): %v", fname, err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		typ, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := typ.Type.(*ast.StructType)
		if !ok {
			return true
		}
		s.structs[typ.Name.Name] = st
		/*
			for _, f := range st.Fields.List {
				if f.Tag == nil {
					continue
				}
				ts, err := strconv.Unquote(f.Tag.Value)
				if err != nil {
					log.Fatalf("unquote: %v", err)
				}
				tag := reflect.StructTag(ts)
				tv, ok := tag.Lookup("bobgen")
				if !ok {
					continue
				}
				_ = tv
				fType, ok := f.Type.(*ast.Ident)
				if !ok {
					// XXX - definitely need to handle slices and arrays here.
					continue
				}
				if len(f.Names) != 1 {
					log.Fatalf("%s embed field name problem: %v", typ.Name.Name, f.Names)
				}
				s.targets[fType.Name] = nil
			}
		*/
		return true
	})
}

var outFname = flag.String("o", "", "output file name")

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: bobgen <package dir> <structs...>\n")
		flag.PrintDefaults()
		os.Exit(1)

	}
	dname := flag.Arg(0)

	fs := token.NewFileSet()
	pkg, err := build.Default.ImportDir(dname, 0)
	if err != nil {
		log.Fatalf("cannot process directory '%s': %v", dname, err)
	}

	outName := strings.TrimSuffix(pkg.Name, ".go") + "_generated.go"
	if *outFname != "" {
		outName = *outFname
	}

	out, err := os.OpenFile(outName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("open(%s): %v", outName, err)
	}
	defer out.Close()
	s := &gen{
		structs:   make(map[string]*ast.StructType),
		targets:   make(map[string]typeInfo),
		generated: make(map[string]bool),
		indent:    0,
		out:       out,
	}

	for i := 1; i < flag.NArg(); i++ {
		s.targets[flag.Arg(i)] = nil
	}

	for _, fname := range pkg.GoFiles {
		if fname == outName {
			continue // Skip the generated file.
		}
		s.parseFile(fs, filepath.Join(dname, fname))
	}

	for _, fname := range pkg.TestGoFiles {
		s.parseFile(fs, filepath.Join(dname, fname))
	}

	s.o("package %s\n\n// Automatically generated by bobgen, do not edit.\n", pkg.Name)

	for {
		done := true
		for k := range s.targets {
			if s.generated[k] {
				continue
			}
			done = false
			if s.targets[k] == nil {
				s.resolveStruct(k).Func(s)
			} else {
				s.targets[k].Func(s)
			}
			s.generated[k] = true
		}
		if done {
			break
		}
	}

}
