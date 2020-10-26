package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"log"
	"strings"
)

type Config struct {
	Models       map[string]map[string]string
	ValueObjects map[string]map[string]string
}

type Model struct {
	Name   string
	Fields []Fields
}

type Fields struct {
	Name string
	Type string
}

func main() {
	data, _ := ioutil.ReadFile("short.yml")
	config := Config{}
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var models []Model
	for modelName, modelFields := range config.Models {
		m := Model{Name: modelName}
		for key, tp := range modelFields {
			f := Fields{Name: key, Type: tp}
			m.Fields = append(m.Fields, f)
		}
		models = append(models, m)
	}
	var vos []Model
	for modelName, modelFields := range config.ValueObjects {
		vo := Model{Name: modelName}

		for key, tp := range modelFields {
			f := Fields{Name: key, Type: tp}
			vo.Fields = append(vo.Fields, f)
		}
		vos = append(vos, vo)
	}
	fmt.Println(models[0])

	funcMap := template.FuncMap{
		"ToUpper":      strings.ToUpper,
		"snakeToCamel": strings.ToUpper,
	}

	tpmlt := `
{{range .}}
type {{.Name}} struct {
{{range .Fields}} {{ .Name }} {{ .Type }} 
{{end}}
{{end}}
}
`

	tmpl, _ := template.New("models").Funcs(funcMap).Parse(tpmlt)

	var result bytes.Buffer

	tmpl.Execute(&result, models)
	fmt.Println(result.String())

}
