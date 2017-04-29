package main

import (
	"flag"
	"fmt"
	"io"
	"os"
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
		io.Copy(os.Stdout, f)
	default:
		usage()
	}
}
