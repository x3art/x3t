package xt

import (
	"encoding/xml"
	"log"
	"os"
)

type PageXML struct {
	Id    int    `xml:"id,attr"`
	Title string `xml:"title,attr"`
	Descr string `xml:"descr,attr"`
	T     []struct {
		Id    int    `xml:"id,attr"`
		Value string `xml:",chardata"`
	} `xml:"t"`
}

type TextFile struct {
	Language struct {
		Id int `xml:"id,attr"`
	} `xml:"language"`
	Pages []PageXML `xml:"page"`
}

type Text map[int]map[int]string

func GetText(n string) Text {
	f, err := os.Open(n)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	t := TextFile{}
	d.Decode(&t)

	ret := make(Text)
	merge := func(min, max int) {
		for pi := range t.Pages {
			px := &t.Pages[pi]
			pid := px.Id
			if pid >= max || pid < min {
				continue
			}
			pid -= min
			if _, ok := ret[pid]; !ok {
				ret[pid] = make(map[int]string, len(px.T))
			}
			for ti := range px.T {
				tx := &px.T[ti]
				ret[pid][tx.Id] = tx.Value
			}
		}

	}
	merge(0, 300000)
	merge(300000, 350000)
	merge(350000, 400000)
	return ret
}
