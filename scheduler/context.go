package scheduler

import (
	"context"

	"github.com/sirupsen/logrus"
)

type key int8

const (
	logKey key = iota
)

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, logKey, entry)
}

// ContextLogger retrieves a logger from the context or returns a default logger
// if not available
func ContextLogger(ctx context.Context) *logrus.Entry {
	v, ok := ctx.Value(logKey).(*logrus.Entry)
	if !ok {
		return logrus.WithContext(ctx)
	}
	return v
}
