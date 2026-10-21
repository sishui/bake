package codegen

import (
	"strings"

	"github.com/sishui/bake/internal/config"
)

func hasRelationship(tags ...*config.Tag) bool {
	for _, tag := range tags {
		if tag.Key != "bun" {
			continue
		}
		for _, opt := range tag.Options {
			name, value, ok := strings.Cut(opt, ":")
			if !ok {
				continue
			}
			switch strings.ToLower(name) {
			case "m2m":
				return true
			case "rel":
				switch strings.ToLower(value) {
				case "belongs-to", "has-one", "has-many":
					return true
				}
			default:
				continue
			}
		}
	}
	return false
}
