package main

import (
	"html/template"
)

type templates map[string]string

var tmpls = templates{}

func (t templates) Add(name, template string) bool {
	t[name] = template
	return true
}

func (t templates) Compile(fm template.FuncMap) *template.Template {
	tmpl := template.New("")
	tmpl.Funcs(fm)
	for k, v := range t {
		template.Must(tmpl.New(k).Parse(v))
	}
	return tmpl
}
