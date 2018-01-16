package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"x3t/xt"
)

func main() {
	flag.Parse()
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	err, scr := xt.DecodeScript(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "xml: %v\n", err)
	}
	for _, l := range scr.SourceText.Lines {
		fmt.Printf("%s%s\n", l.Indent, strings.Join(l.Str, ""))
	}

}
