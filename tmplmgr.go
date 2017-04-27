package main

import "html/template"

type templates map[string]string

var tmpls = templates{}

func (t templates) Add(name, template string) bool {
	t[name] = template
	return true
}

func (t templates) Compile(fm template.FuncMap) *template.Template {
	todo := templates{}
	for k, v := range t {
		todo[k] = v
	}
	if tmplDir, err := AssetDir("templates"); err == nil {
		for _, tn := range tmplDir {
			todo.Add(tn, string(MustAsset("templates/"+tn)))
		}
	}

	tmpl := template.New("")
	tmpl.Funcs(fm)
	for k, v := range todo {
		template.Must(tmpl.New(k).Parse(v))
	}
	return tmpl
}
