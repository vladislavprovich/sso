package slogpretty

import (
	"context"
	"io"
	stdLog "log"
	"log/slog"
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
	//todo
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	//todo
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	//todo
}
