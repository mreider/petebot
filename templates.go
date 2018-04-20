package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

// Template data.
type TemplateData struct {
	Session          map[string]interface{}
	SlackRedirectUrl string
}

// HTML templates registry.
// Based on https://hackernoon.com/golang-template-2-template-composition-and-how-to-organize-template-files-4cb40bcdf8f6
type TemplatesRegistry struct {
	templatesDir string
	templates    map[string]*template.Template
}

// Create new templates registry.
func NewTemplatesRegistry(templatesDir string) *TemplatesRegistry {
	return &TemplatesRegistry{
		templatesDir: templatesDir,
		templates:    make(map[string]*template.Template),
	}
}

// Load templates.
func (r *TemplatesRegistry) LoadTemplates() error {
	layoutFiles, err := filepath.Glob(filepath.Join(r.templatesDir, "layouts", "*.html"))
	if err != nil {
		return err
	}

	includeFiles, err := filepath.Glob(filepath.Join(r.templatesDir, "*.html"))
	if err != nil {
		return err
	}

	mainTemplate := template.New("main")
	mainTemplate, err = mainTemplate.Parse(`{{define "main" }} {{ template "base" . }} {{ end }}`)
	if err != nil {
		return err
	}
	for _, file := range includeFiles {
		fileName := filepath.Base(file)
		files := append(layoutFiles, file)
		r.templates[fileName], err = mainTemplate.Clone()
		if err != nil {
			return err
		}
		r.templates[fileName] = template.Must(r.templates[fileName].ParseFiles(files...))
	}
	return nil
}

// Render specified template with specified data.
func (r *TemplatesRegistry) RenderTemplate(w http.ResponseWriter, name string, data *TemplateData) {
	tmpl, ok := r.templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist.", name), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
