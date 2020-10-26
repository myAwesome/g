package main

import "bytes"
import (
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

	ModelsGo []Model
	VoGo     []Model
	Imports  map[string]bool
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
	config.Imports = make(map[string]bool)
	err := yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	ymlToGo(&config)

	funcMap := template.FuncMap{
		"snakeToCamel": toCamelCase,
	}

	tpmlt := `
{{range .}}
type {{.Name|snakeToCamel}} struct {
{{range .Fields}} {{ .Name|snakeToCamel }} {{ .GoType }}
{{end}}
}
{{end}}
`

	tmpl, _ := template.New("models").Funcs(funcMap).Parse(tpmlt)

	var result bytes.Buffer
	fmt.Println("package main \n")
	if len(config.Imports) > 0 {
		for i, _ := range config.Imports {
			fmt.Println(`import "` + i + `"`)
		}
	}

	tmpl.Execute(&result, config.ModelsGo)
	fmt.Println(result.String())
	result.Reset()

	tmpl.Execute(&result, config.VoGo)
	fmt.Println(result.String())

}

func ymlToGo(config *Config) {
	for modelName, modelFields := range config.Models {
		m := Model{Name: modelName}
		for key, tp := range modelFields {
			f := Fields{Name: key, Type: tp}
			switch tp {
			case "date":
				f.GoType = "time.Time"
				f.DbType = "DATETIME"
				if false == config.Imports["time"] {
					config.Imports["time"] = true
				}
			case "text":
				f.GoType = "string"
				f.DbType = "LONGTEXT"
			case "float":
				f.GoType = "float64"
				f.DbType = "DECIMAL"
			default:
				f.GoType = tp
			}
			m.Fields = append(m.Fields, f)
		}
		config.ModelsGo = append(config.ModelsGo, m)
	}
	for voName, voFields := range config.Vo {
		vo := Model{Name: voName}
		for key, tp := range voFields {
			f := Fields{Name: key, Type: tp}
			switch tp {
			case "date":
				f.GoType = "time.Time"
				f.DbType = "DATETIME"
				if false == config.Imports["time"] {
					config.Imports["time"] = true
				}
			case "text":
				f.GoType = "string"
				f.DbType = "DECIMAL"
			case "float":
				f.GoType = "float64"
				f.DbType = "DECIMAL"
			default:
				f.GoType = tp
			}
			vo.Fields = append(vo.Fields, f)
		}
		config.VoGo = append(config.VoGo, vo)
	}
}
