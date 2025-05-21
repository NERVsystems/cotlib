package cotlib

import (
	"io"
	"log/slog"
	"sync/atomic"
)

// logger is the package-level logger instance
var logger atomic.Pointer[slog.Logger]

func init() {
	// Guarantee there is at least a no-op logger
	if l := logger.Load(); l == nil {
		logger.Store(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}

	// Set default XML parsing limits
	SetMaxXMLSize(2 << 20)
	SetMaxElementDepth(32)
	SetMaxElementCount(10000)
	SetMaxTokenLen(1024)
	SetMaxValueLen(512 * 1024)
}

// SetLogger sets the package-level logger
func SetLogger(l *slog.Logger) {
	if l == nil {
		l = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	logger.Store(l)
}
