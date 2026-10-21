package codegen

import (
	"sort"
	"strings"

	"github.com/sishui/bake/internal/config"
)

type Import struct {
	Path  string
	Alias string
}

func groupImports(module string, imports ...Import) [][]Import {
	if len(imports) == 0 {
		return nil
	}

	groups := [3]map[string]Import{
		make(map[string]Import), // std
		make(map[string]Import), // third-party
		make(map[string]Import), // local
	}
	for _, imp := range imports {
		if imp.Path == "" {
			continue
		}
		index := 1
		switch {
		case module != "" && strings.HasPrefix(imp.Path, module):
			index = 2
		case !strings.Contains(imp.Path, "."):
			index = 0
		}
		groups[index][imp.Path] = imp
	}

	result := make([][]Import, 0, len(groups))
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		result = append(result, sortedImports(group))
	}
	return result
}

func sortedImports(in map[string]Import) []Import {
	imports := make([]Import, 0, len(in))
	for _, imp := range in {
		imports = append(imports, imp)
	}
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Path < imports[j].Path
	})
	return imports
}

func importsForType(typeExpr string, configured []*config.Import) []Import {
	if typeExpr == "" || len(configured) == 0 {
		return nil
	}
	var result []Import
	for _, imp := range configured {
		for _, t := range imp.Types {
			if isQualifiedType(typeExpr, t) {
				result = append(result, Import{Path: imp.Package, Alias: imp.Alias})
				break
			}
		}
	}
	return result
}

func isQualifiedType(typeExpr string, qualified string) bool {
	for i := 0; ; {
		j := strings.Index(typeExpr[i:], qualified)
		if j < 0 {
			return false
		}
		start, end := i+j, i+j+len(qualified)
		if (start == 0 || !isIdentChar(typeExpr[start-1])) && (end == len(typeExpr) || !isIdentChar(typeExpr[end])) {
			return true
		}
		i = end
	}
}

func isIdentChar(c byte) bool {
	return c == '_' ||
		'0' <= c && c <= '9' ||
		'a' <= c && c <= 'z' ||
		'A' <= c && c <= 'Z'
}
