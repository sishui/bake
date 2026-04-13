package generate

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"text/template"

	"github.com/sishui/bake/internal/naming"
)

const templatesDir = "templates"

//go:embed templates/*.tmpl
var templatesFS embed.FS

var templateFuncs = template.FuncMap{
	"add":     add,
	"sub":     sub,
	"longest": longest,
	"concat":  naming.Concat,
	"align":   naming.Align,
}

type templates struct {
	templates map[string]*template.Template
}

func parseTemplates(dir string) (*templates, error) {
	files, err := templatesFS.ReadDir(templatesDir)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*template.Template, len(files))
	for _, file := range files {
		tmpl, err := parseTemplate(templatesDir+"/"+file.Name(), templatesFS.ReadFile)
		if err != nil {
			return nil, err
		}
		result[tmpl.Name()] = tmpl
	}
	result, err = parseCustomTemplates(dir, result)
	if err != nil {
		return nil, err
	}
	return &templates{
		templates: result,
	}, nil
}

func parseCustomTemplates(dir string, result map[string]*template.Template) (map[string]*template.Template, error) {
	if dir == "" {
		return result, nil
	}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".tmpl" {
			return nil
		}
		tmpl, err := parseTemplate(path, os.ReadFile)
		if err != nil {
			return err
		}
		result[tmpl.Name()] = tmpl
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *templates) render(tmplName string, data any) (*bytes.Buffer, error) {
	tmpl, ok := t.templates[tmplName]
	if !ok {
		return nil, fmt.Errorf("template %s not found", tmplName)
	}
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, data)
	if err != nil {
		return nil, err
	}
	return &buffer, nil
}

func parseTemplate(filename string, readBufferFunc func(filename string) ([]byte, error)) (*template.Template, error) {
	buffer, err := readBufferFunc(filename)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(filename)
	name = name[:len(name)-len(filepath.Ext(name))]
	tmpl, err := template.New(name).Funcs(templateFuncs).Parse(string(buffer))
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func add(a int, b int) int {
	return a + b
}

func sub(a int, b int) int {
	return a - b
}

func longest(fields []*Field, kinds ...string) int {
	var length int
	for _, field := range fields {
		if slices.Contains(kinds, field.Kind) {
			if len(field.Name) > length {
				length = len(field.Name)
			}
		}
	}
	return length
}
