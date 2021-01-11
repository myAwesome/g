package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
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
	Name        string
	Fields      []Field
	ReactInputs map[string]bool
}

type Field struct {
	Name      string
	Type      string
	GoType    string
	DbType    string
	ReactType string

	IsId       bool
	IsRelation bool
	Relation   string

	IsEnum     bool
	EnumValues []string
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

	ymlToGoConvert(&config)
	ymlValidate(&config)
	codeGenerate(&config)

}

func ymlValidate(config *Config) {
	fmt.Println("yml Validation ...")

	modelNames := make(map[string]bool)
	for _, model := range config.ModelsGo {
		modelNames[model.Name] = true
	}

	for _, model := range config.ModelsGo {
		for _, field := range model.Fields {
			if field.IsRelation && true != modelNames[field.Relation] {
				fmt.Println("Relation not found: ", field.Relation)
				fmt.Println("model: ", model.Name)
			}
		}
	}
}

func ymlToGoConvert(config *Config) {
	config.Env.Db_Name = config.Env.Db_Name + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	for modelName, modelFields := range config.Models {
		m := Model{Name: modelName}
		m.ReactInputs = make(map[string]bool)
		for key, tp := range modelFields {
			f := Field{Name: key, Type: tp}
			f.IsId = key == "id"
			f.IsRelation = false
			switch true {
			case strings.HasPrefix(tp, "enum_"):
				f.GoType = "string"
				f.IsEnum = true
				f.DbType = "ENUM('" + strings.Replace(tp[5:], "_", "','", -1) + "')"
				enumConfig := strings.Split(tp, "_")
				f.EnumValues = enumConfig[1:]
				if false == m.ReactInputs["SelectInput"] {
					m.ReactInputs["SelectInput"] = true
				}
				break
			case strings.HasPrefix(tp, "rel_"):
				f.GoType = "int"
				f.IsRelation = true
				f.DbType = "INT(11)"
				f.Relation = tp[4:]
				if false == m.ReactInputs["ReferenceInput"] {
					m.ReactInputs["ReferenceInput"] = true
				}
				if false == m.ReactInputs["SelectInput"] {
					m.ReactInputs["SelectInput"] = true
				}
				break
			case tp == "date":
				f.GoType = "string"
				f.DbType = "DATETIME"
				f.ReactType = "DateInput"
				if false == m.ReactInputs["DateInput"] {
					m.ReactInputs["DateInput"] = true
				}
				break
			case tp == "datetime":
				f.GoType = "time.Time"
				f.DbType = "DATETIME"
				f.ReactType = "DateTimeInput"
				if false == config.Imports["time"] {
					config.Imports["time"] = true
				}
				if false == m.ReactInputs["DateTimeInput"] {
					m.ReactInputs["DateTimeInput"] = true
				}
				break
			case tp == "text":
				f.GoType = "string"
				f.DbType = "LONGTEXT"
				f.ReactType = "TextInput"
				if false == m.ReactInputs["TextInput"] {
					m.ReactInputs["TextInput"] = true
				}
				break
			case tp == "float":
				f.GoType = "float64"
				f.DbType = "DECIMAL"
				f.ReactType = "NumberInput"
				if false == m.ReactInputs["NumberInput"] {
					m.ReactInputs["NumberInput"] = true
				}
				break
			case tp == "int":
				f.GoType = "int"
				f.DbType = "INT(11)"
				f.ReactType = "NumberInput"
				if false == m.ReactInputs["NumberInput"] && key != "id" {
					m.ReactInputs["NumberInput"] = true
				}
				break
			case tp == "string":
				f.GoType = "string"
				f.DbType = "VARCHAR(255)"
				f.ReactType = "TextInput"
				if false == m.ReactInputs["TextInput"] {
					m.ReactInputs["TextInput"] = true
				}
				break
			case tp == "bool":
				f.GoType = "bool"
				f.DbType = "BOOLEAN"
				f.ReactType = "BooleanInput"
				if false == m.ReactInputs["BooleanInput"] {
					m.ReactInputs["BooleanInput"] = true
				}
				break
			// todo: Many To One Relation
			case tp == "rel":
				f.GoType = "int"
				f.IsRelation = true
				f.DbType = "INT(11)"
				f.Relation = key
				f.ReactType = "ReferenceInput,SelectInput"
				if false == m.ReactInputs["ReferenceInput"] {
					m.ReactInputs["ReferenceInput"] = true
				}
				if false == m.ReactInputs["SelectInput"] {
					m.ReactInputs["SelectInput"] = true
				}
				break
			default:
				panic("Error unsupported type: " + tp + " model: " + modelName + " field: " + key)
			}
			m.Fields = append(m.Fields, f)
		}
		config.ModelsGo = append(config.ModelsGo, m)
	}

	for relName, relFields := range config.Relations {
		relation := Model{Name: relName}
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
			relation.Fields = append(relation.Fields, f)
		}
		config.RelationsGo = append(config.RelationsGo, relation)
	}
}

func codeGenerate(config *Config) {
	funcMap := template.FuncMap{
		"snakeToCamel": toCamelCase,
		"toUrl":        toUrl,
		"fieldVarName": fieldVarName,
		"count":        count,
	}

	// todo root
	fmt.Println(" ")
	fmt.Println("Root ...")
	fmt.Println(" ")

	err := os.Mkdir("./app", 0750)
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

	// todo back
	fmt.Println(" ")
	fmt.Println("back ...")
	fmt.Println(" ")

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

	envGeneralFile, err := os.Create("./app/.env")
	if err != nil {
		panic(err)
	}

	err = tmpltEnv.ExecuteTemplate(envFile, "env.txt", config.Env)
	if err != nil {
		panic(err)
	}
	err = tmpltEnv.ExecuteTemplate(envGeneralFile, "env.txt", config.Env)
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
	fmt.Println("Back Dockerfile generating...")
	backDockerTmplt, err := template.New("dockerfileServer.txt").Funcs(funcMap).ParseFiles("tpl/dockerfileServer.txt")
	if err != nil {
		panic(err)
	}
	backDocker, err := os.Create(backFolderName + "/Dockerfile")
	if err != nil {
		panic(err)
	}
	err = backDockerTmplt.ExecuteTemplate(backDocker, "dockerfileServer.txt", nil)
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
	frontDockerTmplt, err := template.New("dockerfile.txt").Funcs(funcMap).ParseFiles("tpl/front/dockerfile.txt")
	if err != nil {
		panic(err)
	}

	frontDocker, err := os.Create(frontFolderName + "/Dockerfile")
	if err != nil {
		panic(err)
	}

	err = frontDockerTmplt.ExecuteTemplate(frontDocker, "dockerfile.txt", nil)
	if err != nil {
		panic(err)
	}
}
