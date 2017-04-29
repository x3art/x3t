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

type xdata interface {
	Open() io.ReadCloser
}

type Xfiles struct {
	f map[string]map[string]xdata // [directory][file]
}

func XFiles(dir string) Xfiles {
	ret := Xfiles{f: make(map[string]map[string]xdata)}
	// 01, 02, 03, etc. stop at the first that doesn't exist.
	for i := 1; ret.parseCD(filepath.Join(dir, fmt.Sprintf("%.2d", i))); i++ {
	}
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
	"types": "txt",
}

// The regex capture groups are:
// 0 - all of it
// 1 - directory path
// 2 - path without last directory (unused)
// 3 - last directory in the path
// 4 - file name without suffix
// 5 - suffix
// 6 - size
var pathRe = regexp.MustCompile(`((.+/)*(.+))/(.+)\.(.+) ([0-9]+)`)

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
		if len(split) != 7 {
			log.Printf("can't parse .cat line: '%s'", s.Text())
			continue
		}
		i, err := strconv.ParseInt(string(split[6]), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		d := split[1]
		if xf.f[d] == nil {
			xf.f[d] = make(map[string]xdata)
		}
		suffix := split[5]
		packed := false
		if suffix == "pck" {
			packed = true
			if pm := pckMap[split[3]]; pm == "" {
				log.Printf("Path '%s' has a .pck file without mapping", split[0])
			} else {
				suffix = pm
			}
		}
		xf.f[split[1]][split[4]+"."+suffix] = cd{fd, off, off + i, packed}
		off += i
	}
	return true
}

// io.Reader wrapper to descramble various data.
type stupidDescrambler struct {
	r      io.Reader
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
	pck    bool
}

func (c cd) Open() io.ReadCloser {
	var r io.Reader
	r = io.NewSectionReader(c.f, c.off, c.n)
	if c.pck {
		zr, err := gzip.NewReader(&stupidDescrambler{r: r, cookie: 51})
		if err != nil {
			log.Fatal(err)
		}
		r = zr
	}
	return ioutil.NopCloser(r)
}
