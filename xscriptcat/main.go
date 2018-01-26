package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/x3art/x3t/xt"
)

func main() {
	flag.Parse()
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %s %v\n", flag.Arg(0), err)
		os.Exit(1)
	}
	defer f.Close()
	err, scr := xt.DecodeScript(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "xml: %v\n", err)
	}
	for _, l := range scr.SourceText.Lines {
		a := make([]string, 0)
		for _, s := range l.Str {
			a = append(a, strings.TrimSpace(s))
		}
		fmt.Printf("%s\n", strings.Join(l.Str, ""))
	}

}
