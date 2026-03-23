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
	"strings"
	"text/template"

	"github.com/sishui/bake/internal/config"
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

func parseTemplates(cfg *config.Template) (*templates, error) {
	files, err := templatesFS.ReadDir(templatesDir)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*template.Template, len(files))
	for _, file := range files {
		tmpl, err := parseBuiltinTemplate(file)
		if err != nil {
			return nil, err
		}
		result[tmpl.Name()] = tmpl
	}
	err = filepath.Walk(cfg.Dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}
		tmpl, err := parseCustomTemplate(path)
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
		if err := file.Close(); err != nil {
			slog.ErrorContext(ctx, "close file", "file", fullPath, "error", err)
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

func parseBuiltinTemplate(file fs.DirEntry) (*template.Template, error) {
	filename := templatesDir + "/" + file.Name()
	buffer, err := templatesFS.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseTemplate(filename, buffer)
}

func parseCustomTemplate(filename string) (*template.Template, error) {
	buffer, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parseTemplate(filename, buffer)
}

func parseTemplate(filename string, buffer []byte) (*template.Template, error) {
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
