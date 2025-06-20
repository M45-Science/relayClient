package main

import (
	"embed"
	"os"
	"text/template"
)

const (
	publicTemplateName  = "public.tmpl"
	privateTemplateName = "private.tmpl"
	logoFile            = "logo.ans"

	publicIndexFilename  = "connect-links.html"
	privateIndexFilename = "index.html"
	statsFilename        = "serverStats.json"
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
		doLog("failed to read embedded template %s: %v", publicTemplateName, err)
		os.Exit(1)
	}

	publicServerTemplate, err = template.New("page").Parse(string(content))
	if err != nil {
		doLog("failed to parse template %s: %v", publicTemplateName, err)
		os.Exit(1)
	}

	//Private template
	content, err = templateFiles.ReadFile(tmplDir + privateTemplateName)
	if err != nil {
		doLog("failed to read embedded template %s: %v", privateTemplateName, err)
		os.Exit(1)
	}

	privateServerTemplate, err = template.New("page").Parse(string(content))
	if err != nil {
		doLog("failed to parse template %s: %v", privateTemplateName, err)
		os.Exit(1)
	}

	//Login logo ANSI
	logoANSI, err = templateFiles.ReadFile(tmplDir + logoFile)
	if err != nil {
		doLog("failed to read %s: %v", logoFile, err)
		os.Exit(1)
	}

}
