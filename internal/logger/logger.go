package logger

import (
	"context"
	"log/slog"
)

// FanoutHandler broadcasts records to multiple handlers.
type FanoutHandler struct {
	handlers []slog.Handler
}

// NewFanoutHandler creates a new FanoutHandler with the given handlers.
func NewFanoutHandler(handlers ...slog.Handler) *FanoutHandler {
	return &FanoutHandler{handlers: handlers}
}

// Enabled returns true if any of the handlers are enabled.
func (h *FanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle calls Handle on all underlying handlers.
func (h *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// WithGroup returns a new FanoutHandler with the group applied to all handlers.
func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewFanoutHandler(handlers...)
}

// WithAttrs returns a new FanoutHandler with the attributes applied to all handlers.
func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewFanoutHandler(handlers...)
}
