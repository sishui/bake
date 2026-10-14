package codegen

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

const templatesDir = "templates"

//go:embed templates/*.tmpl
var templatesFS embed.FS

var templateFuncs = template.FuncMap{
	"sub":  sub,
	"dict": dict,
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
	if err := tmpl.Execute(&buffer, data); err != nil {
		return nil, err
	}
	src, err := format.Source(buffer.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format %s output: %w", tmplName, err)
	}
	return bytes.NewBuffer(src), nil
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

func sub(a int, b int) int {
	return a - b
}

// dict creates a map from alternating key-value pairs for use in templates.
func dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("dict: odd number of arguments")
	}
	m := make(map[string]any, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: key %d is not a string", i)
		}
		m[key] = pairs[i+1]
	}
	return m, nil
}
