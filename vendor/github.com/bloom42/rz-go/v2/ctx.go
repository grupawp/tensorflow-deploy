package rz

import (
	"context"
)

type ctxKey struct{}

// ToCtx returns a copy of ctx with l associated. If an instance of Logger
// is already in the context, the context is not updated.
func (l *Logger) ToCtx(ctx context.Context) context.Context {
	if lp, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		if lp == l {
			// Do not store same logger.
			return ctx
		}
	} else if l.level == Disabled {
		// Do not store disabled logger.
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromCtx returns the Logger associated with the ctx. If no logger
// is associated, a New() logger is returned with a addedfield "rz.FromCtx": "error".
//
// For example, to add a field to an existing logger in the context, use this
// notation:
//
//     ctx := r.Context()
//     l := rz.FromCtx(ctx)
//     l.With(...)
func FromCtx(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return l
	}
	logger := New().With(Fields(String("rz.FromCtx", "error")))
	return &logger
}
