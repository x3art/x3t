package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"x3t/xt"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: xtool <X3 directory> <ls|cat> [args]\n")
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() < 2 {
		usage()
	}
	args := flag.Args()
	x := xt.XFiles(args[0])

	switch args[1] {
	case "ls":
		x.Map(func(d, f string) {
			fmt.Printf("%s/%s\n", d, f)
		})
	case "cat":
		if flag.NArg() != 3 {
			usage()
		}
		f := x.Open(args[2])
		if f == nil {
			fmt.Fprintf(os.Stderr, "No such file: %s\n", args[2])
			os.Exit(1)
		}
		defer f.Close()
		io.Copy(os.Stdout, f)
	case "grep":
		if flag.NArg() != 3 {
			usage()
		}
		grep(x, args[2])
	default:
		usage()
	}
}

func grepone(x xt.Xfiles, fn string, needle string) {
	f := x.Open(fn)
	if f == nil {
		log.Fatalf("Eh? %s", fn)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	line := 0
	for s.Scan() {
		line++
		if strings.Index(s.Text(), needle) != -1 {
			fmt.Printf("%s:%d:%s\n", fn, line, s.Text())
		}
	}
}

func grep(x xt.Xfiles, needle string) {
	x.Map(func(dn, fn string) {
		grepone(x, dn+"/"+fn, needle)
	})
}
