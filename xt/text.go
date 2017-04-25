package xt

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	merge(350000, 380000)
	merge(380000, 600000)
	return ret
}

var reCurly = regexp.MustCompile("\\{([[:digit:]]+),([[:digit:]]+)\\}")
var reParen = regexp.MustCompile("\\(.*\\)")

func (t Text) Get(pid, tid int) (string, error) {
	if t[pid] == nil {
		return "", fmt.Errorf("Bad page: %d", pid)
	}
	if t[pid][tid] == "" {
		// This can't be fatal (yet?).
		log.Printf("bad string ID: %d/%d", pid, tid)
		return fmt.Sprintf("bad string %d,%d", pid, tid), nil
	}
	s := t[pid][tid]
	if m := reCurly.FindStringSubmatch(s); len(m) > 0 {
		pid, err := strconv.Atoi(m[1])
		if err != nil {
			return "", fmt.Errorf("Bad page id: %v", m[1])
		}
		tid, err := strconv.Atoi(m[2])
		if err != nil {
			return "", fmt.Errorf("Bad text id: %v", m[2])
		}
		repl, err := t.Get(pid, tid)
		if err != nil {
			return "", err
		}
		s = reCurly.ReplaceAllString(s, repl)
	}
	return strings.TrimSpace(reParen.ReplaceAllString(s, "")), nil
}
