// Package logger provides a simple logger.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/sishui/bake/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Init(c *config.Log) *slog.Logger {
	level := parseLevel(c.Level)

	writers := make([]io.Writer, 0, 2)
	writers = append(writers, os.Stdout)

	if c.File != "" {
		writers = append(writers, &lumberjack.Logger{
			Filename:   c.File,
			MaxSize:    100,
			MaxBackups: 30,
			MaxAge:     10,
			Compress:   true,
			LocalTime:  true,
		})
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.SourceKey {
				if src, ok := attr.Value.Any().(*slog.Source); ok {
					attr.Value = slog.StringValue(shortCaller(src.File, src.Line))
				}
			}
			return attr
		},
	}

	handler := slog.NewTextHandler(io.MultiWriter(writers...), opts)

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func shortCaller(file string, line int) string {
	short := file
	idx := strings.LastIndexByte(short, '/')
	if idx == -1 {
		return short
	}
	idx = strings.LastIndexByte(short[:idx], '/')
	if idx == -1 {
		return short
	}
	return short[idx+1:] + ":" + strconv.Itoa(line)
}
