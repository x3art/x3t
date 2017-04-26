package main

import (
	"log"
	"net/http"
)

var _ = tmpls.Add("map", `
{{template "header"}}
<div style="width: 95%; height: 95%">
	<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="100%" height="100%" viewBox="0 0 24 20" id="themap">
		<style>
			.s {
				width: 0.8;
				height: 0.8;
				stroke: black;
				stroke-width: 0.02;
			}
			{{/*Argon*/}}
			.r1 {
				fill: #a0a0ff;
			}
			{{/*Boron*/}}
			.r2 {
				fill: #a0ffa0;
			}
			{{/*Split*/}}
			.r3 {
				fill: #ffa0ff;
			}
			{{/*Paranid*/}}
			.r4 {
				fill: #ffa0a0;
			}
			{{/*Teladi*/}}
			.r5 {
				fill: #ffffa0;
			}
			{{/*Xenon*/}}
			.r6 {
				fill: #b06666;
			}
			{{/*Kha'ak*/}}
			.r7 {
				fill: #caa5a5;
			}
			{{/*Pirates*/}}
			.r8 {
				fill: #727272;
			}
			{{/*Goner*/}}
			.r9 {
				fill: #6060ff;
			}
			{{/*ufo?*/}}
			.r10 {
			}
			{{/*hostile?*/}}
			.r11 {
			}
			{{/*neutral*/}}
			.r12 {
				fill: #a0a0a0;
			}
			{{/*friendly?*/}}
			.r13 {
			}
			{{/*unknown*/}}
			.r14 {
				fill: #a0a0a0;
			}
			{{/*unused?*/}}
			.r15 {
			}
			{{/*unused?*/}}
			.r16 {
			}
			{{/*ATF*/}}
			.r17 {
				fill: #80ff80;
			}
			{{/*Terran*/}}
			.r18 {
				fill: #b0ffb0;
			}
			{{/*Yaki*/}}
			.r19 {
				fill: #ffff80;
			}
			.sectorname {
				font-size: 1;
			}
			.zoomedsector {
				transform: scale(3);
				fill-opacity: 1.0;
			}
		</style>
		<g>
{{- range .Sectors}}
{{template "map-sector" .}}
{{- end}}
		</g>
	</svg>
</div>
<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.2.1/jquery.min.js"></script>
<script src='/js/svg-pan-zoom.min.js'></script>
<script>
svgPanZoom("#themap")
$(document).ready(function() {
	$("g.sector").hover(
	  function() {
	    this.parentElement.appendChild(this);
	    $(this).find("g:first").addClass("zoomedsector");
	  }, function() {
	    $(this).find("g:first").removeClass("zoomedsector");
  	});
});
</script>
{{template "footer"}}
`)

var _ = tmpls.Add("map-sector", `
<g transform="translate({{.X}} {{.Y}})" class="sector">
 <g>
  <rect class="s r{{.R}}" />
  <g transform="scale(0.12) translate(0.5 1.3)">
{{- range $i, $row := (sectorName .)}}
    <text transform="translate(0 {{$i}})" class="sectorname">{{$row}}</text>
{{- end}}
  </g>
 </g>
</g>
`)

func (st *state) showMap(w http.ResponseWriter, req *http.Request) {
	err := st.tmpl.ExecuteTemplate(w, "map", st.u)
	if err != nil {
		log.Fatal(err)
	}
}
