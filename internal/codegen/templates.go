package generate

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
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
	if dir == "" {
		return &templates{
			templates: result,
		}, nil
	}
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
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

	return &templates{
		templates: result,
	}, nil
}

func (t *templates) writeTo(ctx context.Context, tmplName string, outputDir string, filename string, data any) (string, error) {
	tmpl, ok := t.templates[tmplName]
	if !ok {
		return "", fmt.Errorf("template %s not found", tmplName)
	}
	fullPath := filepath.Join(outputDir, filename+".gen.go")
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "close file", "file", fullPath, "error", closeErr)
		}
		if err == nil {
			return
		}
		if removeErr := os.Remove(fullPath); removeErr != nil {
			slog.ErrorContext(ctx, "remove file", "file", fullPath, "error", removeErr)
		}
	}()
	err = tmpl.Execute(file, data)
	if err != nil {
		return "", err
	}
	err = file.Sync()
	if err != nil {
		return "", err
	}
	return fullPath, nil
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
