package generate

import (
	"sort"
	"strings"
)

func groupImports(module string, packages ...string) [][]string {
	if len(packages) == 0 {
		return nil
	}
	std := make(map[string]struct{})
	third := make(map[string]struct{})
	local := make(map[string]struct{})
	for _, pkg := range packages {
		switch {
		case !strings.Contains(pkg, "."):
			std[pkg] = struct{}{}
		case strings.HasPrefix(pkg, module):
			local[pkg] = struct{}{}
		default:
			third[pkg] = struct{}{}
		}
	}
	result := make([][]string, 0, 3)
	if len(std) > 0 {
		result = append(result, sortedKeys(std))
	}
	if len(third) > 0 {
		result = append(result, sortedKeys(third))
	}
	if len(local) > 0 {
		result = append(result, sortedKeys(local))
	}
	return result
}

func sortedKeys(in map[string]struct{}) []string {
	if len(in) == 0 {
		return nil
	}
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
