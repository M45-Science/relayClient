package main

import (
	"embed"
	"log"
	"os"
	"text/template"
)

const (
	publicTemplateName  = "public.tmpl"
	privateTemplateName = "private.tmpl"
	logoFile            = "logo.ans"

	publicIndexFilename  = "connect-links.html"
	privateIndexFilename = "index.html"
	tmplDir              = "templates/"
)

var (
	publicServerTemplate  *template.Template
	privateServerTemplate *template.Template
	logoANSI              []byte
)

//go:embed templates/*

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

	//Login logo ANSI
	logoANSI, err = os.ReadFile(tmplDir + logoFile)
	if err != nil {
		log.Fatalf("failed to read %s: %v", logoFile, err)
	}

}
