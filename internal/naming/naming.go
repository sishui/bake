// Package naming provides naming texts
package naming

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
	"github.com/jinzhu/inflection"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const aliasSuffix = "_alias"

var variantSanitizeRegex = regexp.MustCompile(`[^a-zA-Z\d\-_]`)

func IsASCIIUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func IsASCIILower(c byte) bool {
	return c >= 'a' && c <= 'z'
}

func IsASCIILetter(c byte) bool {
	return IsASCIIUpper(c) || IsASCIILower(c)
}

func ToASCIIUpper(c byte) byte {
	return c - 32
}

func ToASCIILower(c byte) byte {
	return c + 32
}

func Singular(s string) string {
	return inflection.Singular(s)
}

func ToCamelCase(s string) string {
	r := make([]byte, 0, len(s))
	upperNext := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			upperNext = true
			continue
		}
		if upperNext {
			if IsASCIILower(c) {
				c = ToASCIIUpper(c)
			}
			upperNext = false
		}
		r = append(r, c)
	}
	return string(r)
}

func ToSnakeCase(s string) string {
	r := make([]byte, 0, len(s)+5)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if IsASCIIUpper(c) {
			if i > 0 && i+1 < len(s) && (IsASCIILower(s[i-1]) || IsASCIILower(s[i+1])) {
				r = append(r, '_', ToASCIILower(c))
			} else {
				r = append(r, ToASCIILower(c))
			}
		} else {
			r = append(r, c)
		}
	}
	return string(r)
}

func NormalizeIdentifier(s string) string {
	sanitized := strings.ReplaceAll(variantSanitizeRegex.ReplaceAllString(s, ""), "-", "_")
	if len(sanitized) != 0 && ((sanitized[0] >= '0' && sanitized[0] <= '9') || sanitized[0] == '_') {
		sanitized = "T" + sanitized
	}
	return sanitized
}

func TableToStruct(s string) string {
	if s == "" {
		panic("empty table name")
	}
	parts := camelcase.Split(ToCamelCase(NormalizeIdentifier(s)))
	index := len(parts) - 1
	last := parts[index]
	singular := Singular(last)
	if !strings.EqualFold(last, singular) {
		parts[index] = cases.Title(language.Und).String(singular)
	}
	return strings.Join(parts, "")
}

func TableToAlias(s string) string {
	if s == "" {
		panic("empty table name")
	}
	return s + aliasSuffix
}

func StructToReceiver(s string) string {
	if s == "" {
		panic("empty struct name")
	}
	return strings.ToLower(s[:1])
}

func ColumnToField(s string, initialisms map[string]string) string {
	if s == "" {
		panic("empty column name")
	}
	parts := camelcase.Split(ToCamelCase(NormalizeIdentifier(s)))

	var b strings.Builder
	for _, w := range parts {
		b.WriteString(normalize(w, initialisms))
	}
	return b.String()
}

func SplitCommentLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return lines
}

func Align(s string, maxN int) string {
	if maxN <= 0 {
		return s
	}
	if len(s) >= maxN {
		return s
	}
	return s + strings.Repeat(" ", maxN-len(s))
}

func Concat(a, b string, maxN int) string {
	s := a + b
	if maxN <= 0 {
		return s
	}
	if maxN-len(a) <= 0 {
		return s
	}
	return s + strings.Repeat(" ", maxN-len(a))
}

func applySuffix(s string, initialisms map[string]string) string {
	words := SplitWords(s)

	var b strings.Builder
	for _, w := range words {
		b.WriteString(normalize(w, initialisms))
	}
	return b.String()
}

func SplitWords(s string) []string {
	var words []string
	runes := []rune(s)

	start := 0
	for i := 1; i < len(runes); i++ {
		if unicode.IsLower(runes[i-1]) && unicode.IsUpper(runes[i]) {
			words = append(words, string(runes[start:i]))
			start = i
			continue
		}

		if i+1 < len(runes) &&
			unicode.IsUpper(runes[i-1]) &&
			unicode.IsUpper(runes[i]) &&
			unicode.IsLower(runes[i+1]) {
			words = append(words, string(runes[start:i]))
			start = i
		}
	}

	words = append(words, string(runes[start:]))

	return words
}

func normalize(word string, initialisms map[string]string) string {
	lower := strings.ToLower(word)

	if v, ok := initialisms[lower]; ok {
		return v
	}

	return strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
}
