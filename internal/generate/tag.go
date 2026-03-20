package generate

import (
	"sort"
	"strings"

	"github.com/sishui/bake/internal/config"
	"github.com/sishui/bake/internal/naming"
	"github.com/sishui/bake/internal/schema"
)

type Tag config.Tag

func NewTag(key string, name string, options ...string) *Tag {
	return &Tag{
		Key:     key,
		Name:    name,
		Options: options,
	}
}

func (t *Tag) String() string {
	if t.Key == "" {
		return ""
	}
	var s strings.Builder
	s.WriteString(t.Key)
	s.WriteString(`:"`)
	s.WriteString(t.Name)
	if len(t.Options) > 0 {
		if t.Name != "" {
			s.WriteString(",")
		}
		for i, option := range t.Options {
			if i > 0 {
				s.WriteString(",")
			}
			s.WriteString(option)
		}
	}
	s.WriteString(`"`)
	return s.String()
}

type Tags struct {
	tags []*Tag
}

func NewTags(tags ...*Tag) *Tags {
	t := &Tags{}
	t.Add(tags...)
	return t
}

func (t *Tags) Add(tags ...*Tag) *Tags {
	if len(tags) == 0 {
		return t
	}
	for _, nt := range tags {
		if nt.Key == "" {
			continue
		}
		index := t.find(nt.Key)
		if index < 0 {
			t.tags = append(t.tags, nt)
			continue
		}
		tag := t.tags[index]
		tag.Name = nt.Name
		if nt.Name != "-" {
			tag.Options = mergeOptions(tag.Options, nt.Options)
		} else {
			tag.Options = nil
		}
	}
	sort.Slice(t.tags, func(i, j int) bool {
		return t.tags[i].Key < t.tags[j].Key
	})
	return t
}

func (t *Tags) String() string {
	if len(t.tags) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("`")
	for i, tag := range t.tags {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(tag.String())
	}
	builder.WriteString("`")
	return builder.String()
}

func (t *Tags) find(key string) int {
	for i, tag := range t.tags {
		if tag.Key == key {
			return i
		}
	}
	return -1
}

func newBunTag(c *schema.Column) *Tag {
	options := make([]string, 0, 8)
	if strings.Contains(c.Key, "PRI") {
		options = append(options, "pk")
	}
	if strings.Contains(c.Extra, "auto_increment") {
		options = append(options, "autoincrement")
	}
	if strings.Contains(c.Extra, "unique") {
		options = append(options, "unique")
	}
	if c.DataType == "decimal" {
		options = append(options, "type:"+c.ColumnType)
	}

	if c.Nullable == "YES" {
		options = append(options, "nullzero")
	} else {
		options = append(options, "notnull")
	}
	if c.Default != "" {
		columnDefault := strings.ToLower(c.Default)
		if strings.Contains(columnDefault, `current_timestamp`) {
			options = append(options, `default:`+columnDefault)
		} else {
			options = append(options, `default:'`+c.Default+`'`)
		}
	}
	if c.Name == "deleted_at" {
		options = append(options, "soft_delete")
	}
	return NewTag("bun", c.Name, options...)
}

func newJSONTag(name string, options ...string) *Tag {
	if name == "deleted_at" || name == "-" {
		return NewTag("json", "-")
	}
	return NewTag("json", name, options...)
}

func newCustomTags(fieldName string, cfg []*config.Tag) []*Tag {
	count := len(cfg)
	if count == 0 {
		return nil
	}
	tags := make([]*Tag, 0, count)
	for _, v := range cfg {
		if v.Key == "" {
			continue
		}
		tags = append(tags, NewTag(v.Key, customTagName(fieldName, v.Name), v.Options...))
	}
	return tags
}

func mergeOptions(opts ...[]string) []string {
	var count int
	for _, o := range opts {
		count += len(o)
	}

	if count == 0 {
		return nil
	}

	uniq := make(map[string]struct{}, count)

	for _, o := range opts {
		for _, opt := range o {
			uniq[opt] = struct{}{}
		}
	}

	merged := make([]string, 0, len(uniq))
	for o := range uniq {
		merged = append(merged, o)
	}

	sort.Strings(merged)
	return merged
}

func customTagName(fieldName string, tagName string) string {
	switch tagName {
	case "$SnakeCase":
		return naming.ToSnakeCase(fieldName)
	case "$CamelCase":
		return naming.ToCamelCase(fieldName)
	default:
		return tagName
	}
}
