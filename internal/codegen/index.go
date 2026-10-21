package codegen

import (
	"strings"

	"github.com/sishui/bake/internal/naming"
)

const indexPrefix = "idx_"

type Index struct {
	Key   string
	Value string
}

func newIndexes(values ...string) []Index {
	if len(values) == 0 {
		return nil
	}
	indexes := make([]Index, len(values))
	for i, v := range values {
		indexes[i] = Index{
			Key:   naming.ToCamelCase(trimIndexPrefix(v)),
			Value: v,
		}
	}
	return indexes
}

// trimIndexPrefix drops the conventional "idx_" prefix so the constant does not
// repeat "Index".
func trimIndexPrefix(name string) string {
	trimmed := strings.TrimPrefix(name, indexPrefix)
	if trimmed == "" {
		return name
	}
	return trimmed
}
