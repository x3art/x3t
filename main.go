package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"x3t/xt"
)

func main() {
	flag.Parse()

	text := xt.GetText(flag.Arg(1))
	ships := xt.GetShips(flag.Arg(0), text)

	s, _ := json.MarshalIndent(ships["Mammoth"], "", "\t")
	fmt.Printf("%s", s)
}
