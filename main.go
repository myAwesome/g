package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type Config struct {
	Models    map[string]map[string]string
	Vo        map[string]map[string]string
	Relations map[string]map[string]string

	ModelsGo    []Model
	VoGo        []Model
	RelationsGo []Model
	Imports     map[string]bool
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
	data, _ := ioutil.ReadFile("single.yml")
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

	tmplt, err := template.New("server.txt").Funcs(funcMap).ParseFiles("tpl/server.txt")

	if err != nil {
		panic(err)
	}
	err = tmplt.ExecuteTemplate(os.Stdout, "server.txt", config)
	if err != nil {
		panic(err)
	}
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

	for relName, relFields := range config.Relations {
		vo := Model{Name: relName}
		for key, tp := range relFields {
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
		config.RelationsGo = append(config.RelationsGo, vo)
	}
}
