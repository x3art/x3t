package xt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// Access files according to the rules as I understand them.
// cat/dat files in number order, then actual directories.
// later overriding the earlier.

type xdata interface {
	Open() io.ReadCloser
}

type Xfiles struct {
	f map[string]map[string]xdata // [directory][file]
}

func XFiles(dir string) Xfiles {
	ret := Xfiles{f: make(map[string]map[string]xdata)}
	for i := 1; ret.parseCD(filepath.Join(dir, fmt.Sprintf("%.2d", i))); i++ {
	}
	fmt.Print(ret)
	return ret
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

	_ = fd

	// We deliberately leak the dat file descriptor.  There won't
	// be that many of them, so it's not worth the effort to
	// figure out the logic of when they should be opened and
	// closed.

	s := bufio.NewScanner(&catDescrambler{r: fc})
	s.Scan() // throw away the first line
	off := int64(0)
	for s.Scan() {
		split := bytes.Split(s.Bytes(), []byte{' '})
		if len(split) != 2 {
			log.Fatal("I guess the format is more complex than this")
		}
		i, err := strconv.ParseInt(string(split[1]), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		// we probably need to do some path mangling here
		// Also, how to deal with certain files being .pck without knowing their actual suffix?
		// Do we need a function that maps the directory to the actual file name?
		p := filepath.FromSlash(string(split[0]))
		d, b := filepath.Dir(p), filepath.Base(p)
		if xf.f[d] == nil {
			xf.f[d] = make(map[string]xdata)
		}
		xf.f[d][b] = cd{fd, off, off + i}
		off += i
	}
	return true
}

// io.Reader wrapper to descramble cat files.
type catDescrambler struct {
	r   io.Reader
	off int
}

func (d *catDescrambler) Read(p []byte) (int, error) {
	n, err := d.r.Read(p)
	if err == nil {
		for i := 0; i < n; i++ {
			p[i] ^= byte(219 + d.off + i)
		}
		d.off += n
	}
	return n, nil
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
