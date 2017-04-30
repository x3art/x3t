package xt

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	// 01, 02, 03, etc. stop at the first that doesn't exist.
	for i := 1; ret.parseCD(filepath.Join(dir, "addon", fmt.Sprintf("%.2d", i))); i++ {
	}
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
	// XXX - no idea about the directories below, just shut up the warning for now.
	"dds":       "wtf",
	"v":         "wtf",
	"cutscenes": "wtf",
}

var pathRe = regexp.MustCompile(`(.+) ([0-9]+)`)

// Must be called with native paths, we'll convert back to slashes.
func (xf *Xfiles) add(fn string, xd Xdata) {
	d, f := filepath.Split(fn)
	if filepath.Ext(f) == ".pck" {
		base := filepath.Base(d)
		if pm := pckMap[base]; pm == "" {
			log.Printf("Path '%s' has a .pck file without mapping", fn)
		} else {
			f = strings.TrimSuffix(f, "pck") + pm
		}
		xd = pck{xd}
	}
	d = strings.TrimSuffix(filepath.ToSlash(d), "/")
	if xf.f[d] == nil {
		xf.f[d] = make(map[string]Xdata)
	}
	xf.f[d][f] = xd

}

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

	s := bufio.NewScanner(&stupidDescrambler{r: fc, cookie: 219, addOff: true})
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
		xf.add(filepath.FromSlash(split[1]), cd{fd, off, off + i})
		off += i
	}
	return true
}

// io.Reader wrapper to descramble various data.
type stupidDescrambler struct {
	r      io.ReadCloser
	off    int
	cookie byte
	addOff bool
}

func (d *stupidDescrambler) Read(p []byte) (int, error) {
	n, err := d.r.Read(p)
	if err == nil {
		if d.addOff {
			for i := 0; i < n; i++ {
				p[i] ^= d.cookie + byte(d.off+i)
			}
		} else {
			for i := 0; i < n; i++ {
				p[i] ^= d.cookie
			}
		}
		d.off += n
	}
	return n, err
}

func (d *stupidDescrambler) Close() error {
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

type cd struct {
	f      io.ReaderAt
	off, n int64
}

func (c cd) Open() io.ReadCloser {
	return ioutil.NopCloser(io.NewSectionReader(c.f, c.off, c.n))
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
	hdr := make([]byte, 3, 3)
	_, err := ra.ReadAt(hdr, 0)
	if err != nil {
		log.Fatal(err)
	}

	// Figure out the stupid scrambling.
	//
	// What we're looking for is the first two bytes of a gzip
	// header.  31 and 139. This has the potential of giving false
	// positives. It is in fact trivial to generate a header that
	// will break this and no matter which comparison is done
	// first, it will be the wrong one.
	rs := &stupidDescrambler{r: r}
	if cookie := (hdr[0] ^ 31); cookie^hdr[1] == 139 {
		// the first two bytes are a gzip header xor cookie.
		rs.cookie = cookie
	} else if cookie := (hdr[1] ^ 31); cookie^hdr[2] == 139 {
		// apparently the cookie can be in the first byte and it's xor something.
		// it's easier to just ignore it.
		tmp := make([]byte, 1, 1)
		_, _ = r.Read(tmp)
		rs.cookie = cookie
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
