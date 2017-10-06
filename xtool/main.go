package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"
	"x3t/xt"
	"x3t/xt/bob"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: xtool <X3 directory> <ls|cat> [args]\n")
	fmt.Println(flag.NArg(), flag.NFlag())
	os.Exit(1)
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func main() {
	flag.Parse()
	if flag.NArg() < 2 {
		usage()
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
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
	case "bob":
		if flag.NArg() != 3 {
			usage()
		}

		for i := 0; i < 20; i++ {
			t := time.Now()
			f := x.Open(args[2])
			if f == nil {
				fmt.Fprintf(os.Stderr, "No such file: %s\n", args[2])
				os.Exit(1)
			}
			_ = bob.Read(f)
			f.Close()
			fmt.Printf("T: %v\n", time.Since(t))
		}
	default:
		usage()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		//		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		f.Close()
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
