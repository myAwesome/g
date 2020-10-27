`{{range .}}
type {{.Name|snakeToCamel}} struct {
{{range .Fields}} {{ .Name|snakeToCamel }} {{ .GoType }}
{{end}}}
{{end}}`