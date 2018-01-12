package xt

import "encoding/xml"

type Script struct {
	SourceText struct {
		Lines []struct {
			LineNr string   `xml:"linenr,attr"`
			Indent string   `xml:"indent,attr"`
			Str    []string `xml:",any"`
		} `xml:"line"`
	} `xml:"sourcetext"`
}

func (x *X) DecodeScript(name string) (error, *Script) {
	f := x.xf.Open(name)
	defer f.Close()
	d := xml.NewDecoder(f)
	s := Script{}
	err := d.Decode(&s)
	if err != nil {
		return err, nil
	}
	return nil, &s
}
