package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const helpText = `
Usage: go run gen_http_transport_config/main.go mapping [mapping [mapping [...]]]

mapping has the following format:
input_json_file_path:output_go_file_path

Example:

go run gen_http_transport_config/main.go my/api/transport.json:my/api_go/transport.go
`

func helpExit(errorMsg string) {
	if errorMsg != "" {
		fmt.Fprintln(os.Stderr, errorMsg)
	}
	fmt.Print(helpText)
	os.Exit(1)
}

func main() {
	mappings := os.Args[1:]
	if len(mappings) == 0 {
		helpExit("")
	}

	var pairs [][]string
	for _, m := range mappings {
		pair := strings.Split(m, ":")
		if len(pair) != 2 {
			helpExit("Invalid mapping: " + m)
		}
		pairs = append(pairs, pair)
	}

	var err error
	tpl, err = template.New("go").Parse(tplStr)
	if err != nil {
		helpExit("Error parsing template.")
	}

	errors := 0
	for _, pair := range pairs {
		if !generate(pair[0], pair[1]) {
			errors++
		}
	}

	if errors > 0 {
		fmt.Printf("ERRORS: %d\n", errors)
		os.Exit(1)
	}
}

func generate(inputJSONPath, outputGoPath string) bool {
	f, err := os.Open(inputJSONPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %q: %v\n", inputJSONPath, err)
		return false
	}
	defer f.Close()

	sc := new(ServiceConfig)
	err = json.NewDecoder(f).Decode(sc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding json file %q: %v\n", inputJSONPath, err)
		return false
	}

	f2, err := os.Create(outputGoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v", err)
		return false
	}
	defer f2.Close()

	err = tpl.Execute(f2, sc)
	if err != nil {
		os.Remove(outputGoPath)
		fmt.Fprintf(os.Stderr, "Template execution error: %v", err)
		return false
	}

	return true
}

type ServiceConfig struct {
	ServiceName string            `json:"service_name"`
	Endpoints   []*EndpointConfig `json:"endpoints"`
}

type EndpointConfig struct {
	Method        string `json:"method"`
	Path          string `json:"path"`
	HasReqContent bool   `json:"has_req_content"`
	ReqType       string `json:"req_type"`
	RespType      string `json:"resp_type"`
}

var tpl *template.Template

const tplStr = `/*
DO NOT EDIT!
This file has been generated from JSON by gen_http_transport_config.
*/
package {{ .ServiceName }}

import (
	"reflect"

	"github.com/pasztorpisti/nano/addons/transport/http/config"
)

var HTTPTransportConfig = &config.ServiceConfig{
	ServiceName: {{ printf "%q" .ServiceName }},
	Endpoints: []*config.EndpointConfig{
		{{- range $i, $ep := .Endpoints }}
		{
			Method:        {{ printf "%q" $ep.Method }},
			Path:          {{ printf "%q" $ep.Path }},
			HasReqContent: {{ $ep.HasReqContent }},
			ReqType:       reflect.TypeOf((*{{ $ep.ReqType }})(nil)).Elem(),
			RespType:      reflect.TypeOf((*{{ $ep.RespType }})(nil)).Elem(),
		},
		{{- end }}
	},
}
`
