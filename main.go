package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"x3t/xt"
)

var textFile = flag.String("-strings", "data/0001-L044.xml", "strings file")
var shipsFile = flag.String("-ships", "data/TShips.txt", "ships file")
var cockpitsFile = flag.String("-cockpits", "data/TCockpits.txt", "cockpits file")
var lasersFile = flag.String("-lasers", "data/TLaser.txt", "lasers file")

func main() {
	flag.Parse()

	text := xt.GetText(*textFile)
	ships := xt.GetShips(*shipsFile, text)
	cockpits := xt.GetCockpits(*cockpitsFile, text)
	lasers := xt.GetLasers(*lasersFile, text)

	ship := flag.Arg(0)

	s, _ := json.MarshalIndent(ships[ship], "", "\t")
	fmt.Printf("%s", s)
	s, _ = json.MarshalIndent(cockpits[ships[ship].TurretDescriptor[0].CIndex], "", "\t")
	fmt.Printf("%s", s)

	l := cockpits[ships[ship].TurretDescriptor[0].CIndex].LaserMask
	for i := uint(0); i < 64; i++ {
		if l&(1<<i) != 0 {
			fmt.Println(lasers[i].Description)
		}
	}
}
