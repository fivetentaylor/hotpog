package config

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

var Logger *slog.Logger

func init() {
	if config.Env == "prod" {
		Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		// Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
		Logger = slog.New(NewPrettyHandler(os.Stdout))
	}

	slog.SetDefault(Logger)
}

// PrettyHandler is a custom slog.Handler for development
type PrettyHandler struct {
	slog.Handler
	w io.Writer
}

func (h *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Enable all log levels
	return true
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	// Format the log level with color
	level := r.Level.String()
	switch r.Level {
	case slog.LevelDebug:
		level = "\033[36mDEBUG\033[0m" // Cyan
	case slog.LevelInfo:
		level = "\033[32mINFO\033[0m" // Green
	case slog.LevelWarn:
		level = "\033[33mWARN\033[0m" // Yellow
	case slog.LevelError:
		level = "\033[31mERROR\033[0m" // Red
	}

	// Format the timestamp
	timeStr := r.Time.Format(time.DateTime)

	// Format the message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("%s %s %s", timeStr, level, r.Message))

	// Format attributes
	r.Attrs(func(attr slog.Attr) bool {
		msg.WriteString(fmt.Sprintf(" \033[34m%s\033[0m=", attr.Key)) // Blue for keys
		msg.WriteString(fmt.Sprintf("\033[37m%v\033[0m", attr.Value)) // White for values
		return true
	})

	// Format multi-line strings (e.g., stack traces)
	msgStr := msg.String()
	if strings.Contains(msgStr, "\n") {
		msgStr = strings.ReplaceAll(msgStr, "\n", "\n\t") // Indent new lines
	}

	// Write the formatted log
	_, err := fmt.Fprintln(h.w, msgStr)
	return err
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Return a new handler with additional attributes
	return &PrettyHandler{w: h.w}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	// Return a new handler with a group
	return &PrettyHandler{w: h.w}
}

// NewPrettyHandler creates a new PrettyHandler
func NewPrettyHandler(w io.Writer) *PrettyHandler {
	return &PrettyHandler{w: w}
}
