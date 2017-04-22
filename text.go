package main

import (
	"encoding/xml"
	"log"
	"os"
)

/*
<page id="7" title="Boardcomp. Sectornames" descr="Names of all sectors (spoken by Boardcomputer)" voice="yes">
 <t id="1020000">Unknown Sector</t>
 <t id="1020101">Kingdom End</t>
 <t id="1020117">{7,1020000}</t>
*/

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

func textDec(n string) Text {
	f, err := os.Open(n)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	d := xml.NewDecoder(f)
	t := TextFile{}
	d.Decode(&t)

	ret := make(Text)
	for pi := range t.Pages {
		px := &t.Pages[pi]
		if px.Id >= 300000 {
			continue
		}
		pid := px.Id
		if _, ok := ret[pid]; !ok {
			ret[pid] = make(map[int]string, len(px.T))
		}
		for ti := range px.T {
			tx := &px.T[ti]
			ret[pid][tx.Id] = tx.Value
		}
	}
	for pi := range t.Pages {
		px := &t.Pages[pi]
		if px.Id < 300000 {
			continue
		}
		pid := px.Id - 300000
		if _, ok := ret[pid]; !ok {
			ret[pid] = make(map[int]string, len(px.T))
		}
		for ti := range px.T {
			tx := &px.T[ti]
			ret[pid][tx.Id] = tx.Value
		}
	}
	return ret
}
