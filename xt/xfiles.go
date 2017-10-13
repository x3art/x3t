package xt

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// Access files according to the rules as I understand them.
// cat/dat files in number order, then actual directories.
// Latter overriding the earlier.

type Xdata interface {
	Open() io.ReadCloser
}

type Xfiles struct {
	f map[string]map[string]Xdata // [directory][file]
}

func XFiles(dir string) Xfiles {
	ret := Xfiles{f: make(map[string]map[string]Xdata)}
	// 01.{cat,dat}, 02.{cat,dat}, etc. stop at the first that doesn't exist.
	// XXX - how are the non-addon directory cat files involved here?
	for i := 1; ret.parseCD(filepath.Join(dir, "addon", fmt.Sprintf("%.2d", i))); i++ {
	}
	// Now, the normal files.
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		relpath, err := filepath.Rel(dir, path)
		if err != nil {
			log.Fatal(err)
		}
		ret.add(relpath, fs(path))
		return nil
	})
	return ret
}

func (xf *Xfiles) Open(fname string) io.ReadCloser {
	a := strings.LastIndex(fname, "/")
	if dir := xf.f[fname[:a]]; dir == nil {
		return nil
	} else if f := dir[fname[a+1:]]; f == nil {
		return nil
	} else {
		return f.Open()
	}

}

func (xf *Xfiles) Map(f func(string, string)) {
	for dir := range xf.f {
		for fn := range xf.f[dir] {
			f(dir, fn)
		}
	}
}

// If a file has the suffix .pck in a certain directory, what suffix should it have?
var pckMap = map[string]string{
	"types":    "txt",
	"t":        "xml",
	"scripts":  "xml",
	"maps":     "xml",
	"director": "xml",
	"dds":      "dds",
	// XXX - no idea about the directories below, just shut up the warning for now.
	"v":         "wtf",
	"cutscenes": "wtf",
	"mov":       "mov",
	".":         "wtf",
}

// Must be called with native paths, we'll convert back to slashes.
func (xf *Xfiles) add(fn string, xd Xdata) {
	d, f := filepath.Split(fn)
	switch filepath.Ext(f) {
	case ".pck":
		base := filepath.Base(d)
		if pm := pckMap[base]; pm == "" {
			log.Printf("Path '%s' (%s) has a .pck file without mapping", fn, base)
		} else {
			f = strings.TrimSuffix(f, "pck") + pm
		}
		xd = pck{xd}
	case ".pbd":
		xd = pck{xd}
	case ".bob":
		xd = bobOpener{xd}
	}
	d = strings.TrimSuffix(filepath.ToSlash(d), "/")
	if xf.f[d] == nil {
		xf.f[d] = make(map[string]Xdata)
	}
	xf.f[d][f] = xd

}

var pathRe = regexp.MustCompile(`(.+) ([0-9]+)`)

func (xf *Xfiles) parseCD(basename string) bool {
	fc, err := os.Open(basename + ".cat")
	if err != nil {
		return false
	}
	defer fc.Close()
	fd, err := os.Open(basename + ".dat")
	if err != nil {
		log.Fatalf("cat(%s) without dat: %v", basename, err)
	}

	// We deliberately leak the dat file descriptors. There won't
	// be that many of them, so it's not worth the effort to
	// figure out the logic of when they should be opened and
	// closed.

	s := bufio.NewScanner(&stupidDescramblerOff{r: fc, cookie: 219, addOff: true})
	s.Scan() // throw away the first line
	off := int64(0)
	for s.Scan() {
		split := pathRe.FindStringSubmatch(s.Text())
		if len(split) != 3 {
			log.Printf("can't parse .cat line: '%s' (%v)", s.Text(), split)
			continue
		}
		i, err := strconv.ParseInt(string(split[2]), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		xf.add(filepath.FromSlash(split[1]), cd{fd, off, i})
		off += i
	}
	return true
}

type stupidDescrambler struct {
	io.ReadCloser
	c64 uint64
}

func newSD(r io.ReadCloser, c byte) io.ReadCloser {
	c64 := uint64(c)
	c64 |= c64 << 8
	c64 |= c64 << 16
	c64 |= c64 << 32
	return &stupidDescrambler{r, c64}
}

const wordSize = 8
const supportsUnaligned = runtime.GOARCH == "amd64"

func fastXOR(b []byte, c64 uint64) {
	n := len(b)

	w := n / wordSize

	bw := *(*[]uint64)(unsafe.Pointer(&b))
	for i := 0; i < w; i++ {
		bw[i] ^= c64
	}
	for i := (n - n%wordSize); i < n; i++ {
		b[i] ^= byte(c64 & 0xff)
	}
}

func safeXOR(b []byte, c byte) {
	for i := range b {
		b[i] ^= c
	}
}

func (s *stupidDescrambler) Read(p []byte) (int, error) {
	n, err := s.ReadCloser.Read(p)
	if supportsUnaligned {
		fastXOR(p, s.c64)
	} else {
		safeXOR(p, byte(s.c64&0xff))
	}
	return n, err
}

// io.Reader wrapper to descramble various data.
type stupidDescramblerOff struct {
	r      io.ReadCloser
	off    int
	cookie byte
	addOff bool
}

func (d *stupidDescramblerOff) Read(p []byte) (int, error) {
	n, err := d.r.Read(p)
	if err == nil && d.addOff {
		if d.addOff {
			for i := 0; i < n; i++ {
				p[i] ^= d.cookie + byte(d.off+i)
			}
		} else {
			c := d.cookie
			for i := range p {
				p[i] ^= c
			}
		}
		d.off += n
	}
	return n, err
}

func (d *stupidDescramblerOff) Close() error {
	return d.r.Close()
}

type fs string

func (fn fs) Open() io.ReadCloser {
	f, err := os.Open(string(fn))
	if err != nil {
		log.Fatal(err)
	}
	return f
}

type readerWithAt interface {
	io.Reader
	io.ReaderAt
}

type nopclose struct {
	readerWithAt
}

func (nc nopclose) Close() error {
	return nil
}

type cd struct {
	f      io.ReaderAt
	off, n int64
}

func (c cd) Open() io.ReadCloser {
	return nopclose{io.NewSectionReader(c.f, c.off, c.n)}
}

type pck struct {
	xd Xdata
}

type pckReader struct {
	zr *gzip.Reader
	r  io.ReadCloser
}

func (p pck) Open() io.ReadCloser {
	r := p.xd.Open()
	ra := r.(io.ReaderAt)
	// 31, 139
	hdr := make([]byte, 4, 4)
	_, err := ra.ReadAt(hdr, 0)
	if err != nil {
		log.Fatal(err)
	}

	// Figure out the stupid scrambling.
	//
	// What we're looking for is the first two bytes of a gzip
	// header - 31, 139 (and 8 because we expect deflate).
	//
	// This code has the potential of giving false positives. It
	// is in fact trivial to generate a header that will break
	// this and no matter which comparison is done first, it will
	// be the wrong one. This seems to work for now.
	rs := &stupidDescramblerOff{r: r}
	if cookie := (hdr[0] ^ 31); hdr[1]^cookie == 139 && hdr[2]^cookie == 8 {
		// the first two bytes are a gzip header xor cookie.
		rs.cookie = cookie
	} else if cookie := (hdr[1] ^ 31); hdr[2]^cookie == 139 && hdr[3]^cookie == 8 {
		// apparently the cookie can be in the first byte and it's xor something.
		// it's easier to just ignore it and just figure it out from the header.

		// eat first byte
		tmp := make([]byte, 1, 1)
		_, _ = r.Read(tmp)
		rs.cookie = cookie
	} else if cookie := (hdr[0] ^ 31); hdr[1]^(cookie+1) == 138 && hdr[2]^(cookie+2) == 8 {
		rs.cookie = cookie
		rs.addOff = true
	} else if cookie := (hdr[1] ^ 31); hdr[2]^(cookie+2) == 138 && hdr[3]^(cookie+3) == 8 {
		tmp := make([]byte, 1, 1)
		_, _ = r.Read(tmp)

		rs.cookie = cookie
		rs.addOff = true
	} else {
		log.Printf("unknown scrambling method. fingers crossed.")
	}

	zr, err := gzip.NewReader(rs)
	if err != nil {
		log.Fatal("gzip.NewReader: ", err)
	}
	return &pckReader{zr, r}
}

func (pr *pckReader) Read(p []byte) (int, error) {
	return pr.zr.Read(p)
}

func (pr *pckReader) Close() error {
	pr.zr.Close()
	return pr.r.Close()
}

type bobOpener struct {
	xd Xdata
}

func (b bobOpener) Open() io.ReadCloser {
	return newSD(b.xd.Open(), 51)
}
