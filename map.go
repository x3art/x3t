package main

import (
	"log"
	"net/http"
)

var _ = tmpls.Add("map", `
{{template "header"}}
<div style="width: 95%; height: 95%">
	<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="100%" height="100%" viewBox="0 0 23 18" id="themap">
		<style>
			.s {
				width: 0.8;
				height: 0.8;
				stroke: black;
				stroke-width: 0.02;
				fill: red;
				fill-opacity: 1.0;
			}
			{{/*Argon*/}}
			.r1 {
				fill: blue;
				fill-opacity: 0.4;
			}
			{{/*Boron*/}}
			.r2 {
				fill: green;
				fill-opacity: 0.6;
			}
			{{/*Split*/}}
			.r3 {
				fill: purple;
				fill-opacity: 0.6;
			}
			{{/*Paranid*/}}
			.r4 {
				fill: red;
				fill-opacity: 0.6;
			}
			{{/*Teladi*/}}
			.r5 {
				fill: yellow;
				fill-opacity: 0.6;
			}
			{{/*Xenon*/}}
			.r6 {
				fill: brown;
				fill-opacity: 0.6;
			}
			{{/*Kha'ak*/}}
			.r7 {
				fill: brown;
				fill-opacity: 0.4;
			}
			{{/*Pirates*/}}
			.r8 {
				fill: black;
				fill-opacity: 0.2;
			}
			{{/*Goner*/}}
			.r9 {
				fill: blue;
				fill-opacity: 0.7;
			}
			{{/*ufo?*/}}
			.r10 {
			}
			{{/*hostile?*/}}
			.r11 {
			}
			{{/*neutral*/}}
			.r12 {
				fill: black;
				fill-opacity: 0.2;
			}
			{{/*friendly?*/}}
			.r13 {
			}
			{{/*unknown*/}}
			.r14 {
				fill: black;
				fill-opacity: 0.3;
			}
			{{/*unused?*/}}
			.r15 {
			}
			{{/*unused?*/}}
			.r16 {
			}
			{{/*ATF*/}}
			.r17 {
				fill: green;
				fill-opacity: 0.2;
			}
			{{/*Terran*/}}
			.r18 {
				fill: green;
				fill-opacity: 0.4;
			}
			{{/*Yaki*/}}
			.r19 {
				fill: yellow;
				fill-opacity: 0.2;
			}
			.sectorname {
				font-size: 1;
			}
		</style>
		<g>
{{- range .Sectors}}
{{template "map-sector" .}}
{{- end}}
		</g>
	</svg>
</div>
<script src='/js/svg-pan-zoom.min.js'></script>
<script>
var foo = svgPanZoom("#themap")
</script>
{{template "footer"}}
`)

var _ = tmpls.Add("map-sector", `
<g transform="translate({{.X}} {{.Y}})">
  <rect class="s r{{.R}}" />
  <g transform="scale(0.12) translate(0.5 1.3)">
{{- range $i, $row := (sectorName .)}}
    <text transform="translate(0 {{$i}})" class="sectorname">{{$row}}</text>
{{- end}}
  </g>
</g>
`)

func (st *state) showMap(w http.ResponseWriter, req *http.Request) {
	err := st.tmpl.ExecuteTemplate(w, "map", st.u)
	if err != nil {
		log.Fatal(err)
	}
}
