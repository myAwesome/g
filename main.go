package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
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
	Relations map[string]map[string]string
	Imports   map[string]bool

	Env         Env
	ModelsGo    []Model
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
	Metadata    interface{} `yaml:"_meta"`
	ReactInputs map[string]bool
}

type Field struct {
	Name       string
	Type       string
	GoType     string
	DbType     string
	ReactType  string
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

func addReactInput(m *Model, inputType string) {
	if !m.ReactInputs[inputType] {
		m.ReactInputs[inputType] = true
	}
}

func getYAMLFileName() string {
	if name := os.Getenv("YML"); name != "" {
		return name
	}
	return "single.yml"
}

func main() {
	fileName := getYAMLFileName()

	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	config := Config{Imports: make(map[string]bool)}

	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("error parsing YAML: %v", err)
	}

	setDefaultEnv(&config)
	logParsedModels(&config)
	ymlToGoConvert(&config)
	ymlValidate(&config)
	codeGenerate(&config)
}

func setDefaultEnv(config *Config) {
	if config.Env.Project == "" {
		config.Env.Project = "project_default"
	}
	if config.Env.Db_Pass == "" {
		config.Env.Db_Pass = "pass"
	}
	if config.Env.Db_User == "" {
		config.Env.Db_User = "user"
	}
	if config.Env.Db_Name == "" {
		config.Env.Db_Name = "db_name"
	}
	config.Env.Db_Name = config.Env.Db_Name + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	if config.Env.Db_Port == 0 {
		config.Env.Db_Port = 3316
	}
	if config.Env.Server_Port == 0 {
		config.Env.Server_Port = 8833
	}
}

func logParsedModels(parsed *Config) {
	for modelName, fields := range parsed.Models {
		fmt.Printf("Model: %s\n", modelName)
		for field, value := range fields {
			fmt.Printf("  Field: %s, Type: %v\n", field, value)
		}
	}
}

func ymlValidate(config *Config) {
	fmt.Println("YML Validation ...")

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

func addDefaultId(modelFields map[string]string) {
	if _, exists := modelFields["id"]; !exists {
		modelFields["id"] = "int"
	}
}

func ymlToGoConvert(config *Config) {

	for modelName, modelFields := range config.Models {
		m := Model{Name: modelName}
		m.ReactInputs = make(map[string]bool)

		addDefaultId(modelFields)

		for key, tp := range modelFields {
			f := Field{Name: key, Type: tp}
			f.IsId = key == "id"
			f.IsRelation = false
			if strings.HasPrefix(tp, "relation_") || strings.HasPrefix(tp, "enum") {
				if strings.HasPrefix(tp, "relation_") {
					f.GoType = "int"
					f.IsRelation = true
					f.DbType = "INT(11)"
					f.Relation = tp[9:]
					addReactInput(&m, "ReferenceInput")
					addReactInput(&m, "SelectInput")

				} else {
					f.GoType = "string"
					f.IsEnum = true
					enumValues := strings.Split(tp[5:len(tp)-1], `,`)
					for i, value := range enumValues {

						fmt.Println("value", value)
						fmt.Println("trim", value[1:len(value)-1])

						enumValues[i] = value[1 : len(value)-1]
					}
					fmt.Println("ENUM VALUES: ", enumValues)
					f.EnumValues = enumValues
					enumValuesFormatted := "'" + strings.Join(enumValues, "','") + "'"
					f.DbType = "ENUM(" + enumValuesFormatted + ")"
					addReactInput(&m, "SelectInput")
				}

			} else {
				switch tp {
				case "date":
					f.GoType = "string"
					f.DbType = "DATETIME"
					f.ReactType = "DateInput"
					addReactInput(&m, "DateInput")
					break
				case "datetime":
					f.GoType = "time.Time"
					f.DbType = "DATETIME"
					f.ReactType = "DateTimeInput"
					if false == config.Imports["time"] {
						config.Imports["time"] = true
					}
					addReactInput(&m, "DateTimeInput")

					break
				case "text":
					f.GoType = "string"
					f.DbType = "LONGTEXT"
					f.ReactType = "TextInput"
					addReactInput(&m, "TextInput")
					break
				case "float":
					f.GoType = "float64"
					f.DbType = "DECIMAL"
					f.ReactType = "NumberInput"
					addReactInput(&m, "NumberInput")
					break
				case "int":
					f.GoType = "int"
					f.DbType = "INT(11)"
					f.ReactType = "NumberInput"
					if false == m.ReactInputs["NumberInput"] && key != "id" {
						m.ReactInputs["NumberInput"] = true
					}
					break
				case "string":
					f.GoType = "string"
					f.DbType = "VARCHAR(255)"
					f.ReactType = "TextInput"
					addReactInput(&m, "TextInput")
					break
				case "boolean":
					f.GoType = "bool"
					f.DbType = "BOOLEAN"
					f.ReactType = "BooleanInput"
					addReactInput(&m, "BooleanInput")

					break
				// todo: Many To One Relation
				case "rel":
					f.GoType = "int"
					f.IsRelation = true
					f.DbType = "INT(11)"
					f.Relation = key
					f.ReactType = "ReferenceInput,SelectInput"
					addReactInput(&m, "ReferenceInput")
					addReactInput(&m, "SelectInput")
					break
				default:
					panic("Error unsupported type: " + tp + " model: " + modelName + " field: " + key)
				}
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

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func mustMkdir(path string) {
	err := os.Mkdir(path, 0750)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
}

func generateFromTemplate(outputPath, templatePath, templateName string, data interface{}, funcMap template.FuncMap) {
	tmpl, err := template.New(templateName).Funcs(funcMap).ParseFiles(templatePath)
	checkErr(err)
	file, err := os.Create(outputPath)
	checkErr(err)
	err = tmpl.ExecuteTemplate(file, templateName, data)
	checkErr(err)
}

func codeGenerate(config *Config) {
	funcMap := template.FuncMap{
		"snakeToCamel": toCamelCase,
		"toUrl":        toUrl,
		"fieldVarName": fieldVarName,
		"count":        count,
	}

	projectRoot := "./app/" + config.Env.Project
	fmt.Println("\nGenerating Root Structure...")
	mustMkdir(projectRoot)

	// Root-level files
	generateFromTemplate(projectRoot+"/docker-compose.yml", "tpl/docker-compose.txt", "docker-compose.txt", nil, funcMap)
	generateFromTemplate(projectRoot+"/sql.sql", "tpl/sql.txt", "sql.txt", config, funcMap)

	// Backend
	fmt.Println("\nGenerating Backend...")
	backFolder := projectRoot + "/back"
	mustMkdir(backFolder)

	generateFromTemplate(backFolder+"/server.go", "tpl/server.txt", "server.txt", config, funcMap)
	generateFromTemplate(backFolder+"/.env", "tpl/env.txt", "env.txt", config.Env, funcMap)
	generateFromTemplate(projectRoot+"/.env", "tpl/env.txt", "env.txt", config.Env, funcMap)
	generateFromTemplate(backFolder+"/routes.txt", "tpl/routes.txt", "routes.txt", config, funcMap)
	generateFromTemplate(backFolder+"/Dockerfile", "tpl/dockerfileServer.txt", "dockerfileServer.txt", nil, funcMap)

	// Frontend
	fmt.Println("\nGenerating Frontend...")
	frontFolder := projectRoot + "/front"
	mustMkdir(frontFolder)
	mustMkdir(frontFolder + "/public")
	mustMkdir(frontFolder + "/src")

	generateFromTemplate(frontFolder+"/public/index.html", "tpl/front/index.txt", "index.txt", config.Env.Project, funcMap)
	generateFromTemplate(frontFolder+"/package.json", "tpl/front/package.txt", "package.txt", config.Env.Project, funcMap)
	generateFromTemplate(frontFolder+"/src/index.js", "tpl/front/indexjs.txt", "indexjs.txt", config, funcMap)
	generateFromTemplate(frontFolder+"/src/App.js", "tpl/front/appjs.txt", "appjs.txt", config, funcMap)
	generateFromTemplate(frontFolder+"/Dockerfile", "tpl/front/dockerfile.txt", "dockerfile.txt", nil, funcMap)

	// Components per model
	componentTpl := "tpl/front/frc.txt"
	componentEditTpl := "tpl/front/fr-edit.txt"

	for _, model := range config.ModelsGo {
		componentDir := frontFolder + "/src/" + model.Name
		mustMkdir(componentDir)

		generateFromTemplate(componentDir+"/create.js", componentTpl, "frc.txt", model, funcMap)
		generateFromTemplate(componentDir+"/edit.js", componentEditTpl, "fr-edit.txt", model, funcMap)
	}
}
