package templates

import (
	"embed"
	"html/template"
	"log"
)

//go:embed *.html
var tmplFS embed.FS

// Парсит все шаблоны из embed FS
func ParseTemplates() *template.Template {
	tmpl, err := template.ParseFS(tmplFS, "*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
	return tmpl
}
