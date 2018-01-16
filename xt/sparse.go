package xt

import (
	"encoding/xml"
	"io"
)

type Script struct {
	SourceText struct {
		Lines []struct {
			LineNr string   `xml:"linenr,attr"`
			Indent string   `xml:"indent,attr"`
			Str    []string `xml:",any"`
		} `xml:"line"`
	} `xml:"sourcetext"`
}

func DecodeScript(r io.Reader) (error, *Script) {
	d := xml.NewDecoder(r)
	s := Script{}
	err := d.Decode(&s)
	if err != nil {
		return err, nil
	}
	return nil, &s
}
