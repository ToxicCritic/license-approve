package templates

import (
	"embed"
	"html/template"
	"log"
	"time"
)

//go:embed *.html
var tmplFS embed.FS

func dateFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func ParseTemplates() *template.Template {
	tmpl := template.New("").Funcs(template.FuncMap{
		"dateFormat": dateFormat,
	})

	tmpl, err := tmpl.ParseFS(tmplFS, "*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
	return tmpl
}
