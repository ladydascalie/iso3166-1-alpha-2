// this file is used to retrieve the version of the ISO country codes from
// the datasource named id rawFile.
// the iso.go file is then generated.
package main

import (
	"bytes"
	"encoding/json"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

const rawFile = "https://datahub.io/core/country-codes/r/country-codes.json"
const textFile = "latest_iso.txt"

type ISOSource []struct {
	Name              string `json:"CLDR display name"`
	CodeISO3166Alpha2 string `json:"ISO3166-1-Alpha-2"`
	CodeISO3166Alpha3 string `json:"ISO3166-1-Alpha-3"`
}

var client = &http.Client{
	Timeout: 5 * time.Second,
}
var tpl = `
// Code generated by "make build"; DO NOT EDIT.

package iso3166

type ISO31661Alpha2 string

func (i ISO31661Alpha2) String() string {
	return string(i)
}

const (
	{{range $k, $v := . -}}
		// {{$v.Name }}
		{{ $v.CodeISO3166Alpha2 }} ISO31661Alpha2 = "{{$v.CodeISO3166Alpha2 }}"
	{{end -}}
)

var ValidAlpha2Codes = []ISO31661Alpha2{
	{{range $k, $v := .}}
		{{- $v.CodeISO3166Alpha2 -}},
	{{end -}}
}

type ISO31661Alpha3 string

func (i ISO31661Alpha3) String() string {
	return string(i)
}

const (
	{{range $k, $v := . -}}
		// {{$v.Name }}
		{{ $v.CodeISO3166Alpha3 }} ISO31661Alpha3 = "{{$v.CodeISO3166Alpha3 }}"
	{{end -}}
)

var ValidAlpha3Codes = []ISO31661Alpha3{
	{{range $k, $v := .}}
		{{- $v.CodeISO3166Alpha3 -}},
	{{end -}}
}
`

func main() {
	// request the raw file
	resp, err := client.Get(rawFile)
	if err != nil {
		log.Printf("error connecting to datahub: %v", err)
	}
	// read it
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading response body: %v", err)
	}
	defer resp.Body.Close()

	var source ISOSource
	if err := json.Unmarshal(b, &source); err != nil {
		log.Fatalf("cannot unmarshal json: %v", err)
	}
	for i, s := range source {
		if s.Name == "" {
			source = append(source[:i], source[i+1:]...)
		}
	}

	// execute the template into a buffer
	var buf bytes.Buffer
	t := template.Must(template.New("iso").Parse(tpl))
	if err := t.Execute(&buf, source); err != nil {
		log.Fatal(err)
	}

	// format the new source code
	byt, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("error formatting source: %v", err)
	}

	// create the outfile
	outfile, err := os.Create("../iso.go")
	if err != nil {
		log.Printf("error opening outfile: %v", err)
	}

	// write it to disk
	if _, err := outfile.Write(byt); err != nil {
		log.Printf("error writing to disk: %v", err)
	}
}
