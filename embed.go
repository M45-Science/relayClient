package main

import (
	"embed"
	"log"
	"text/template"
)

const (
	publicTemplateName  = "public.tmpl"
	privateTemplateName = "private.tmpl"

	publicIndexFilename  = "connect-links.html"
	privateIndexFilename = "index.html"
	tmplDir              = "templates/"
)

var (
	publicServerTemplate  *template.Template
	privateServerTemplate *template.Template
)

//go:embed templates/*.tmpl
var templateFiles embed.FS

func init() {
	//Public template
	content, err := templateFiles.ReadFile(tmplDir + publicTemplateName)
	if err != nil {
		log.Fatalf("failed to read embedded template %s: %v", publicTemplateName, err)
	}

	publicServerTemplate, err = template.New("page").Parse(string(content))
	if err != nil {
		log.Fatalf("failed to parse template %s: %v", publicTemplateName, err)
	}

	//Private template
	content, err = templateFiles.ReadFile(tmplDir + privateTemplateName)
	if err != nil {
		log.Fatalf("failed to read embedded template %s: %v", privateTemplateName, err)
	}

	privateServerTemplate, err = template.New("page").Parse(string(content))
	if err != nil {
		log.Fatalf("failed to parse template %s: %v", privateTemplateName, err)
	}

}
