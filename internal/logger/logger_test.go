package logger

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errMockHandler = errors.New("handler error")

// MockHandler is a mock implementation of slog.Handler for testing purposes.
type MockHandler struct {
	EnabledFunc   func(context.Context, slog.Level) bool
	HandleFunc    func(context.Context, slog.Record) error
	WithAttrsFunc func([]slog.Attr) slog.Handler
	WithGroupFunc func(string) slog.Handler
}

// Ensure MockHandler implements slog.Handler.
var _ slog.Handler = (*MockHandler)(nil)

func (m *MockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if m.EnabledFunc != nil {
		return m.EnabledFunc(ctx, level)
	}

	return true
}

func (m *MockHandler) Handle(ctx context.Context, r slog.Record) error {
	if m.HandleFunc != nil {
		return m.HandleFunc(ctx, r)
	}

	return nil
}

func (m *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if m.WithAttrsFunc != nil {
		return m.WithAttrsFunc(attrs)
	}

	return m
}

func (m *MockHandler) WithGroup(name string) slog.Handler {
	if m.WithGroupFunc != nil {
		return m.WithGroupFunc(name)
	}

	return m
}

func TestFanoutHandler_Enabled(t *testing.T) {
	t.Parallel()
	t.Run("returns true if at least one handler is enabled", func(t *testing.T) {
		t.Parallel()

		h1 := &MockHandler{
			EnabledFunc: func(ctx context.Context, level slog.Level) bool { return false },
		}
		h2 := &MockHandler{
			EnabledFunc: func(ctx context.Context, level slog.Level) bool { return true },
		}
		fanout := NewFanoutHandler(h1, h2)
		assert.True(t, fanout.Enabled(context.Background(), slog.LevelInfo))
	})

	t.Run("returns false if no handlers are enabled", func(t *testing.T) {
		t.Parallel()

		h1 := &MockHandler{
			EnabledFunc: func(ctx context.Context, level slog.Level) bool { return false },
		}
		h2 := &MockHandler{
			EnabledFunc: func(ctx context.Context, level slog.Level) bool { return false },
		}
		fanout := NewFanoutHandler(h1, h2)
		assert.False(t, fanout.Enabled(context.Background(), slog.LevelInfo))
	})
}

func createMockHandler(enabled bool, called *bool) *MockHandler {
	return &MockHandler{
		EnabledFunc: func(ctx context.Context, level slog.Level) bool { return enabled },
		HandleFunc: func(ctx context.Context, r slog.Record) error {
			if called != nil {
				*called = true
			}

			return nil
		},
	}
}

func TestFanoutHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("calls handle on all enabled handlers", func(t *testing.T) {
		t.Parallel()

		h1Called := false
		h1 := createMockHandler(true, &h1Called)

		h2Called := false
		h2 := createMockHandler(true, &h2Called)

		fanout := NewFanoutHandler(h1, h2)
		err := fanout.Handle(context.Background(), slog.Record{Level: slog.LevelInfo})
		require.NoError(t, err)
		assert.True(t, h1Called)
		assert.True(t, h2Called)
	})

	t.Run("skips disabled handlers", func(t *testing.T) {
		t.Parallel()

		h1Called := false
		h1 := createMockHandler(false, &h1Called)

		h2Called := false
		h2 := createMockHandler(true, &h2Called)

		fanout := NewFanoutHandler(h1, h2)
		err := fanout.Handle(context.Background(), slog.Record{Level: slog.LevelInfo})
		require.NoError(t, err)
		assert.False(t, h1Called)
		assert.True(t, h2Called)
	})

	t.Run("returns first error encountered", func(t *testing.T) {
		t.Parallel()

		h1 := &MockHandler{
			EnabledFunc: func(ctx context.Context, level slog.Level) bool { return true },
			HandleFunc: func(ctx context.Context, r slog.Record) error {
				return errMockHandler
			},
		}
		h2 := &MockHandler{
			EnabledFunc: func(ctx context.Context, level slog.Level) bool { return true },
			HandleFunc: func(ctx context.Context, r slog.Record) error {
				return nil
			},
		}

		fanout := NewFanoutHandler(h1, h2)
		err := fanout.Handle(context.Background(), slog.Record{Level: slog.LevelInfo})
		assert.Equal(t, errMockHandler, err)
	})
}

func TestFanoutHandler_WithGroup(t *testing.T) {
	t.Parallel()
	t.Run("calls WithGroup on all handlers", func(t *testing.T) {
		t.Parallel()

		h1Called := false
		h1 := &MockHandler{
			WithGroupFunc: func(name string) slog.Handler {
				h1Called = true
				return &MockHandler{}
			},
		}
		h2Called := false
		h2 := &MockHandler{
			WithGroupFunc: func(name string) slog.Handler {
				h2Called = true
				return &MockHandler{}
			},
		}

		fanout := NewFanoutHandler(h1, h2)
		newHandler := fanout.WithGroup("test-group")
		assert.IsType(t, &FanoutHandler{}, newHandler)
		assert.True(t, h1Called)
		assert.True(t, h2Called)
	})
}

func TestFanoutHandler_WithAttrs(t *testing.T) {
	t.Parallel()
	t.Run("calls WithAttrs on all handlers", func(t *testing.T) {
		t.Parallel()

		h1Called := false
		h1 := &MockHandler{
			WithAttrsFunc: func(attrs []slog.Attr) slog.Handler {
				h1Called = true
				return &MockHandler{}
			},
		}
		h2Called := false
		h2 := &MockHandler{
			WithAttrsFunc: func(attrs []slog.Attr) slog.Handler {
				h2Called = true
				return &MockHandler{}
			},
		}

		fanout := NewFanoutHandler(h1, h2)
		newHandler := fanout.WithAttrs([]slog.Attr{{Key: "k", Value: slog.StringValue("v")}})
		assert.IsType(t, &FanoutHandler{}, newHandler)
		assert.True(t, h1Called)
		assert.True(t, h2Called)
	})
}
