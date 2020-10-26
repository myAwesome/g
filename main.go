package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

type Config struct {
	Models map[string]map[string]string
	Vo     map[string]map[string]string
}

type Model struct {
	Name   string
	Fields []Fields
}

type Fields struct {
	Name   string
	Type   string
	GoType string
	DbType string
}

var link = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")

func toCamelCase(str string) string {
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

func main() {
	data, _ := ioutil.ReadFile("models.yml")
	config := Config{}
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(config.Vo)
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
	for voName, voFields := range config.Vo {
		vo := Model{Name: voName}

		for key, tp := range voFields {
			f := Fields{Name: key, Type: tp}
			vo.Fields = append(vo.Fields, f)
		}
		vos = append(vos, vo)
	}

	funcMap := template.FuncMap{
		"snakeToCamel": toCamelCase,
	}

	tpmlt := `
{{range .}}
type {{.Name|snakeToCamel}} struct {
{{range .Fields}} {{ .Name|snakeToCamel }} {{ .Type }}
{{end}}
}
{{end}}
`

	tmpl, _ := template.New("models").Funcs(funcMap).Parse(tpmlt)

	var result bytes.Buffer
	fmt.Println("package main")

	tmpl.Execute(&result, models)
	fmt.Println(result.String())
	result.Truncate(0)

	tmpl.Execute(&result, vos)
	fmt.Println(result.String())

	//TODO:

}
