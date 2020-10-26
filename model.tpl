{{range .}}
type {{.Name}} struct {
{{range .Fields}} {{ .Name }} {{ .Type }}
{{end}}
{{end}}
}