package slogpretty

import (
	"context"
	"fmt"
	color "github.com/fatih/color"
	"io"
	stdLog "log"
	"log/slog"
	"strings"
	"time"
)

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type PrettyHandler struct {
	opts PrettyHandlerOptions
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func (opts PrettyHandlerOptions) NewPrettyHandler(
	out io.Writer,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
	}

	return h
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	var levelColor *color.Color

	switch r.Level {
	case slog.LevelDebug:
		levelColor = color.New(color.BgYellow)
	case slog.LevelError:
		levelColor = color.New(color.BgRed)
	case slog.LevelWarn:
		levelColor = color.New(color.BgGreen)
	default:
		levelColor = color.New(color.FgWhite)
	}

	b.WriteString(levelColor.Sprintf("[%s]", r.Level.String()))
	b.WriteString(fmt.Sprintf(" %s: ", r.Time.Format(time.RFC3339)))

	r.Attrs(func(attr slog.Attr) bool {
		b.WriteString(fmt.Sprintf("%s=%v ", attr.Key, attr.Value))
		return true
	})

	b.WriteString(r.Message)
	h.l.Println(b.String())
	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := append(h.attrs, attrs...)
	return &PrettyHandler{
		opts:    h.opts,
		Handler: h.Handler.WithAttrs(attrs),
		l:       h.l,
		attrs:   newAttrs,
	}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		opts:    h.opts,
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
		attrs:   h.attrs,
	}
}
