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
	Models    map[string]map[string]string
	Vo        map[string]map[string]string
	Relations map[string]map[string]string
	Imports   map[string]bool

	Env         Env
	ModelsGo    []Model
	VoGo        []Model
	RelationsGo []Model
}

type Env struct {
	Db_Port     int
	Db_User     string
	Db_Pass     string
	Db_Name     string
	Server_Port int
	Project     string
}

type Model struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name      string
	Type      string
	GoType    string
	DbType    string
	ReactType string

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
	return len(arr) - 1
}

func main() {
	fileName := os.Getenv("YML")
	if fileName == "" {
		fileName = "single.yml"
	}
	data, _ := ioutil.ReadFile(fileName)
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

	// todo Back
	fmt.Println(" ")
	fmt.Println("BACK ...")
	fmt.Println(" ")

	err = os.Mkdir("./app", 0750)
	if err != nil {
		panic(err)
	}

	fmt.Println("docker-compose generating...")
	dcTemplt, err := template.New("docker-compose.txt").Funcs(funcMap).ParseFiles("tpl/docker-compose.txt")
	if err != nil {
		panic(err)
	}

	dcFile, err := os.Create("./app/docker-compose.yml")
	if err != nil {
		panic(err)
	}
	err = dcTemplt.ExecuteTemplate(dcFile, "docker-compose.txt", nil)
	if err != nil {
		panic(err)
	}

	backFolderName := "./app/back"
	err = os.Mkdir(backFolderName, 0750)
	if err != nil {
		panic(err)
	}

	// server file
	fmt.Println("server generating...")
	tmplt, err := template.New("server.txt").Funcs(funcMap).ParseFiles("tpl/server.txt")

	if err != nil {
		panic(err)
	}
	file, err := os.Create(backFolderName + "/server.go")
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
	envFile, err := os.Create(backFolderName + "/.env")
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
	sqlFile, err := os.Create(backFolderName + "/sql.sql")
	if err != nil {
		panic(err)
	}
	err = tmpltSql.ExecuteTemplate(sqlFile, "sql.txt", config)
	if err != nil {
		panic(err)
	}

	// ROUTES
	fmt.Println("routes generating...")
	tmpltRoutes, err := template.New("routes.txt").Funcs(funcMap).ParseFiles("tpl/routes.txt")

	if err != nil {
		panic(err)
	}

	routesFile, err := os.Create(backFolderName + "/routes.txt")
	if err != nil {
		panic(err)
	}
	err = tmpltRoutes.ExecuteTemplate(routesFile, "routes.txt", config)
	if err != nil {
		panic(err)
	}

	// Dockerfile
	fmt.Println("Dockerfile generating...")
	_, err = os.Create(backFolderName + "/Dockerfile")
	if err != nil {
		panic(err)
	}

	// todo FRONT
	fmt.Println(" ")
	fmt.Println("FRONT ...")
	fmt.Println(" ")

	frontFolderName := "./app/front"
	err = os.Mkdir(frontFolderName, 0750)
	if err != nil {
		panic(err)
	}

	err = os.Mkdir(frontFolderName+"/public", 0750)
	if err != nil {
		panic(err)
	}

	// INDEX
	indexFront, err := template.New("index.txt").Funcs(funcMap).ParseFiles("tpl/front/index.txt")
	if err != nil {
		panic(err)
	}
	indexFrontFile, err := os.Create(frontFolderName + "/public/index.html")
	if err != nil {
		panic(err)
	}
	err = indexFront.ExecuteTemplate(indexFrontFile, "index.txt", config.Env.Project)
	if err != nil {
		panic(err)
	}

	// package
	packageFrontTmplt, err := template.New("package.txt").Funcs(funcMap).ParseFiles("tpl/front/package.txt")
	if err != nil {
		panic(err)
	}
	packageFrontFile, err := os.Create(frontFolderName + "/package.json")
	if err != nil {
		panic(err)
	}
	err = packageFrontTmplt.ExecuteTemplate(packageFrontFile, "package.txt", config.Env.Project)
	if err != nil {
		panic(err)
	}

	// index.js
	err = os.Mkdir(frontFolderName+"/src", 0750)
	if err != nil {
		panic(err)
	}

	indexjsFrontTmplt, err := template.New("indexjs.txt").Funcs(funcMap).ParseFiles("tpl/front/indexjs.txt")
	if err != nil {
		panic(err)
	}
	indexjsFrontFile, err := os.Create(frontFolderName + "/src/index.js")
	if err != nil {
		panic(err)
	}
	err = indexjsFrontTmplt.ExecuteTemplate(indexjsFrontFile, "indexjs.txt", config)
	if err != nil {
		panic(err)
	}

	// app.js
	appjsFrontTmplt, err := template.New("appjs.txt").Funcs(funcMap).ParseFiles("tpl/front/appjs.txt")
	if err != nil {
		panic(err)
	}
	appjsFrontFile, err := os.Create(frontFolderName + "/src/App.js")
	if err != nil {
		panic(err)
	}
	err = appjsFrontTmplt.ExecuteTemplate(appjsFrontFile, "appjs.txt", config)
	if err != nil {
		panic(err)
	}

	componentFrontTmplt, err := template.New("frc.txt").Funcs(funcMap).ParseFiles("tpl/front/frc.txt")
	if err != nil {
		panic(err)
	}

	componentFrontEditTmplt, err := template.New("fr-edit.txt").Funcs(funcMap).ParseFiles("tpl/front/fr-edit.txt")
	if err != nil {
		panic(err)
	}

	for _, v := range config.ModelsGo {
		err = os.Mkdir(frontFolderName+"/src/"+v.Name, 0750)
		if err != nil {
			panic(err)
		}

		frcFile, err := os.Create(frontFolderName + "/src/" + v.Name + "/create.js")
		if err != nil {
			panic(err)
		}
		err = componentFrontTmplt.ExecuteTemplate(frcFile, "frc.txt", v)
		if err != nil {
			panic(err)
		}

		freditFile, err := os.Create(frontFolderName + "/src/" + v.Name + "/edit.js")
		if err != nil {
			panic(err)
		}
		err = componentFrontEditTmplt.ExecuteTemplate(freditFile, "fr-edit.txt", v)
		if err != nil {
			panic(err)
		}
	}

	// Dockerfile
	fmt.Println("Dockerfile generating...")
	_, err = os.Create(frontFolderName + "/Dockerfile")
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
				f.ReactType = "DateTimeInput"
				if false == config.Imports["time"] {
					config.Imports["time"] = true
				}
				break
			case "text":
				f.GoType = "string"
				f.DbType = "LONGTEXT"
				f.ReactType = "TextInput"
				break
			case "float":
				f.GoType = "float64"
				f.DbType = "DECIMAL"
				f.ReactType = "NumberInput"
				break
			case "int":
				f.GoType = "int"
				f.DbType = "INT(11)"
				f.ReactType = "NumberInput"
				break
			case "string":
				f.GoType = "string"
				f.DbType = "VARCHAR(255)"
				f.ReactType = "TextInput"
				break
			case "bool":
				f.GoType = "bool"
				f.DbType = "BOOLEAN"
				f.ReactType = "BooleanInput"
				break
			// todo: Many To One Relation
			case "rel":
				f.GoType = "int"
				f.IsRelation = true
				f.DbType = "INT(11)"
				break
			default:
				f.IsRelation = true
				f.GoType = tp
			}
			m.Fields = append(m.Fields, f)
		}
		config.ModelsGo = append(config.ModelsGo, m)
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
