package xt

import (
	"bufio"
	"strings"
)

type SceneP struct {
	P  string
	B  string
	C  string
	N  string
	BB bool
}

type Scene struct {
	P []SceneP
}

func (x *X) Scene(f string) *Scene {
	r := x.Open(f)
	defer r.Close()

	sc := &Scene{}

	scan := bufio.NewScanner(r)
	for scan.Scan() {
		ln := scan.Text()
		if ln == "" || ln[0] != 'P' {
			continue
		}
		p := SceneP{}
		spl := strings.Split(ln, ";")
		for i := range spl {
			f := strings.TrimSpace(spl[i])
			ch := f[0]
			x := strings.TrimSpace(f[1:])
			switch ch {
			case 'P':
				p.P = x
			case 'B':
				p.B = x
			case 'C':
				p.C = x
			case 'N':
				p.N = x
			case 'b':
				p.BB = true
			}
		}
		sc.P = append(sc.P, p)
	}
	return sc
}
