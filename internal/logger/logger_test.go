package logger

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"info", slog.LevelInfo},
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
		{"DEBUG", slog.LevelDebug},
		{"Warn", slog.LevelWarn},
		{"ERROR", slog.LevelError},
		{"Info", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestShortCaller(t *testing.T) {
	tests := []struct {
		name string
		file string
		line int
		want string
	}{
		{
			name: "no slash returns file as-is",
			file: "main.go",
			line: 42,
			want: "main.go",
		},
		{
			name: "one slash returns full path",
			file: "cmd/main.go",
			line: 10,
			want: "cmd/main.go",
		},
		{
			name: "two slashes returns last two segments",
			file: "cmd/bake/main.go",
			line: 15,
			want: "bake/main.go:15",
		},
		{
			name: "multiple slashes returns last two segments",
			file: "github.com/sishui/bake/cmd/bake/main.go",
			line: 100,
			want: "bake/main.go:100",
		},
		{
			name: "windows backslashes treated as no slash",
			file: "cmd\\bake\\main.go",
			line: 1,
			want: "cmd\\bake\\main.go",
		},
		{
			name: "trailing slash",
			file: "a/b/",
			line: 5,
			want: "b/:5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortCaller(tt.file, tt.line)
			if got != tt.want {
				t.Errorf("shortCaller(%q, %d) = %q, want %q", tt.file, tt.line, got, tt.want)
			}
		})
	}
}
