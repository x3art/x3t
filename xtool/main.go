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

		f := x.Open(args[2])
		if f == nil {
			fmt.Fprintf(os.Stderr, "No such file: %s\n", args[2])
			os.Exit(1)
		}
		b, err := bob.Read(f)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("info: %s\n", b.Info)
		fmt.Printf("bodies: %d\n", len(b.Bodies))
		for i := range b.Bodies {
			bod := &b.Bodies[i]
			fmt.Printf("sz %d, fl %d\n", bod.Size, bod.Flags)
			fmt.Printf("bones: %v\n", bod.Bones)
			fmt.Printf("weights: %v\n", bod.Weights)
			fmt.Printf("parts: %v\n", len(bod.Parts))
			for j := range bod.Parts {
				fmt.Printf("pfl: 0x%x\n", bod.Parts[j].Flags)
				p := bod.Parts[j].P.(bob.PartX3)
				fmt.Printf("pfacelist: %d\n", len(p.FacesX3))
			}
			fmt.Printf("points: %v\n", len(bod.Points))
		}

		f.Close()
	case "bobBench":
		t := time.Now()
		x.Map(func(d, f string) {
			if strings.HasSuffix(f, ".bob") {
				n := fmt.Sprintf("%s/%s", d, f)
				f := x.Open(n)
				if f == nil {
					log.Fatalf("open: %s", n)
				}
				_, err := bob.Read(f)
				if err != nil {
					fmt.Printf("error: %s, %v\n", n, err)
				} else {
					fmt.Printf("ok: %s\n", n)
				}
				f.Close()
			}
		})
		fmt.Printf("T: %v\n", time.Since(t))
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
