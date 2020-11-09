package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type Config struct {
	Models      map[string]map[string]string
	Vo          map[string]map[string]string
	Relations   map[string]map[string]string
	Env         Env
	ModelsGo    []Model
	VoGo        []Model
	RelationsGo []Model
	Imports     map[string]bool
}

type Env struct {
	Db_Port     int
	Db_User     string
	Db_Pass     string
	Db_Name     string
	Server_Port int
}

type Model struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name   string
	Type   string
	GoType string
	DbType string

	IsId       bool
	IsRelation bool
}

var link = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")

func toCamelCase(str string) string {
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

func toUrl(str string) string {
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.Replace(s, "_", "-", -1)
	})
}

func fieldVarName(str string) string {
	strCC := toCamelCase(str)

	if len(strCC) < 2 {
		return strings.ToLower(strCC)
	}

	bts := []byte(strCC)

	lc := bytes.ToLower([]byte{bts[0]})
	rest := bts[1:]

	return string(bytes.Join([][]byte{lc, rest}, nil))

}

func count(arr []Field) int {
	return len(arr)
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
		"toUrl":        toUrl,
		"fieldVarName": fieldVarName,
		"count":        count,
	}

	// server file
	fmt.Println("server generating...")
	tmplt, err := template.New("server.txt").Funcs(funcMap).ParseFiles("tpl/server.txt")

	if err != nil {
		panic(err)
	}
	file, err := os.Create("./app/server.go")
	if err != nil {
		panic(err)
	}
	err = tmplt.ExecuteTemplate(file, "server.txt", config)
	if err != nil {
		panic(err)
	}

	// ENV file
	fmt.Println("env generating...")
	tmpltEnv, err := template.New("env.txt").Funcs(funcMap).ParseFiles("tpl/env.txt")

	if err != nil {
		panic(err)
	}
	envFile, err := os.Create("./app/.env")
	if err != nil {
		panic(err)
	}
	err = tmpltEnv.ExecuteTemplate(envFile, "env.txt", config.Env)
	if err != nil {
		panic(err)
	}

	// SQL
	fmt.Println("sql generating...")
	tmpltSql, err := template.New("sql.txt").Funcs(funcMap).ParseFiles("tpl/sql.txt")

	if err != nil {
		panic(err)
	}
	sqlFile, err := os.Create("./app/sql.sql")
	if err != nil {
		panic(err)
	}
	err = tmpltSql.ExecuteTemplate(sqlFile, "sql.txt", config)
	if err != nil {
		panic(err)
	}

}
func ymlToGo(config *Config) {
	for modelName, modelFields := range config.Models {
		m := Model{Name: modelName}
		for key, tp := range modelFields {
			f := Field{Name: key, Type: tp}
			f.IsId = key == "id"
			f.IsRelation = false
			switch tp {
			case "date":
				f.GoType = "time.Time"
				f.DbType = "DATETIME"
				if false == config.Imports["time"] {
					config.Imports["time"] = true
				}
				break
			case "text":
				f.GoType = "string"
				f.DbType = "LONGTEXT"
				break
			case "float":
				f.GoType = "float64"
				f.DbType = "DECIMAL"
				break
			case "int":
				f.GoType = "int"
				f.DbType = "INT(11)"
				break
			case "string":
				f.GoType = "string"
				f.DbType = "VARCHAR(255)"
				break
			default:
				f.IsRelation = true
				f.IsId = true
				f.GoType = tp
			}
			m.Fields = append(m.Fields, f)
		}
		config.ModelsGo = append(config.ModelsGo, m)
	}
	for voName, voFields := range config.Vo {
		vo := Model{Name: voName}
		for key, tp := range voFields {
			f := Field{Name: key, Type: tp}
			switch tp {
			case "date":
				f.GoType = "time.Time"
				f.DbType = "DATETIME"
				if false == config.Imports["time"] {
					config.Imports["time"] = true
				}
				break
			case "text":
				f.GoType = "string"
				f.DbType = "LONGTEXT"
				break
			case "float":
				f.GoType = "float64"
				f.DbType = "DECIMAL"
				break
			case "int":
				f.GoType = "int"
				f.DbType = "INT(11)"
				break
			case "string":
				f.GoType = "string"
				f.DbType = "VARCHAR(255)"
				break
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
			f := Field{Name: key, Type: tp}
			for _, modelType := range config.ModelsGo {
				if key == modelType.Name {
					for _, vvalue := range modelType.Fields {
						if vvalue.Name == tp {
							f = Field{Name: key, Type: tp, GoType: vvalue.GoType, DbType: vvalue.DbType}
						}
					}
				}
			}
			vo.Fields = append(vo.Fields, f)
		}
		config.RelationsGo = append(config.RelationsGo, vo)
	}
}
