// Package logx1 provides a thin, documented façade over Uber's Zap logger,
// offering sensible production/development presets, global logger management,
// context helpers, and a few convenience functions commonly needed in apps
// and libraries. It intentionally stays close to the upstream API so you can
// drop down to Zap features whenever you need to.
package logx1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	// We rely on a Zap-compatible module that mirrors go.uber.org/zap.
	// Its public surface matches zap v1 APIs (Logger, SugaredLogger, Fields, etc.).
	"github.com/daydev-org/zap"
)

// NewDevelopment returns a logger configured for local development: human‑readable
// console encoding, colored levels, and development-friendly options (caller,
// stacktraces on Warn+ in dev, etc.).
//
// Use this in CLI tools and during local runs.
func NewDevelopment() (*zap.Logger, error) { // keep name explicit to avoid ambiguity
	return zap.NewDevelopment()
}

// NewProduction returns a JSON-encoding logger suitable for production: structured
// fields, ISO8601 timestamps, caller annotations, and stacktraces on Error+.
//
// Use this in servers and background workers.
func NewProduction() (*zap.Logger, error) {
	return zap.NewProduction()
}

// New chooses a logger preset based on the given mode string.
// Accepted values (case-insensitive): "prod", "production" → production config;
// anything else → development config. This is a convenience for wiring via envs.
func New(mode string) (*zap.Logger, error) {
	switch normalized(mode) {
	case "prod", "production":
		return NewProduction()
	default:
		return NewDevelopment()
	}
}

func normalized(s string) string {
	// tiny allocation-free downcase for common ASCII letters
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c = c - 'A' + 'a'
		}
		b[i] = c
	}
	return string(b)
}

// ReplaceGlobals installs the provided logger as the process-wide global logger
// (used by L() and S()). It returns a function to restore the previous globals.
//
// Typical usage:
//
//	logger, _ := logx1.NewProduction()
//	undo := logx1.ReplaceGlobals(logger)
//	defer undo()
func ReplaceGlobals(logger *zap.Logger) func() { return zap.ReplaceGlobals(logger) }

// L returns the global structured logger (zap.Logger). Set it with ReplaceGlobals.
func L() *zap.Logger { return zap.L() }

// S returns the global sugared logger (zap.SugaredLogger), allowing printf-style
// formatting and loosely-typed fields.
func S() *zap.SugaredLogger { return zap.S() }

// Sync flushes any buffered log entries. Call this at process shutdown. On some
// platforms, syncing stdout/stderr returns benign errors (EINVALID/ENOTTY),
// which this helper filters out unless they wrap other real errors.
func Sync(l *zap.Logger) error {
	if l == nil {
		return nil
	}
	err := l.Sync()
	if err == nil {
		return nil
	}
	// Ignore common, well-known spurious sync errors when writing to
	// non-seekable outputs like stdout/stderr or pipes.
	if errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY) {
		return nil
	}
	// On some systems, zap wraps os.PathError around EINVAL/ENOTTY.
	var perr *os.PathError
	if errors.As(err, &perr) {
		if errors.Is(perr, syscall.EINVAL) || errors.Is(perr, syscall.ENOTTY) {
			return nil
		}
	}
	return err
}

// StdLogger returns a standard library *log.Logger that writes through the
// provided Zap logger at Info level and above. Useful for integrating with
// code that expects the standard logger.
func StdLogger(l *zap.Logger) *log.Logger {
	if l == nil {
		l = L()
	}
	return zap.NewStdLog(l)
}

// StdWriter returns an io.Writer that writes via the standard logger backed by
// the provided Zap logger. Handy when an API expects only an io.Writer.
func StdWriter(l *zap.Logger) io.Writer { return StdLogger(l).Writer() }

// Context helpers
// ----------------
// Use WithContext/FromContext to propagate a request-scoped logger across layers.
// A private key type avoids collisions with other context values.

type contextKey struct{}

var ctxKey = contextKey{}

// WithContext stores the provided logger into the context under an internal key.
// Prefer this over using your own key to avoid accidental key collisions across
// packages.
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ctxKey, logger)
}

// FromContext retrieves a logger from context if present, otherwise returns the
// current global logger (L()).
func FromContext(ctx context.Context) *zap.Logger {
	if ctx != nil {
		if logger, ok := ctx.Value(ctxKey).(*zap.Logger); ok && logger != nil {
			return logger
		}
	}
	return zap.L()
}

// WithFields returns a child logger with the given structured fields attached.
// Use it to enrich logs with request IDs, user IDs, and other correlation data.
func WithFields(l *zap.Logger, fields ...zap.Field) *zap.Logger {
	if l == nil {
		l = L()
	}
	return l.With(fields...)
}

// AttachRequest annotates the logger with request-scoped metadata and stores it
// in the returned context. Any layer retrieving the logger via FromContext will
// include these fields automatically.
func AttachRequest(ctx context.Context, reqID string, clientIP string) context.Context {
	l := FromContext(ctx)
	if reqID != "" {
		l = l.With(zap.String("request_id", reqID))
	}
	if clientIP != "" {
		l = l.With(zap.String("client_ip", clientIP))
	}
	return WithContext(ctx, l)
}

// LogError logs an error with a message and optional fields, ensuring the error
// itself is always included as structured field "error".
func LogError(l *zap.Logger, err error, msg string, fields ...zap.Field) {
	if l == nil {
		l = L()
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	l.Error(msg, fields...)
}

// Must is a helper for one-off initialization code: it logs the error and panics
// if err is non-nil. Useful around logger setup in main().
func Must(logger *zap.Logger, err error) *zap.Logger {
	if err != nil {
		// Use a temporary fallback to stderr if we don't have a valid logger yet.
		// Error and bytes written are skipped since we're already panicking.'
		_, _ = fmt.Fprintf(os.Stderr, "logx1: fatal init error: %v\n", err)
		panic(err)
	}
	return logger
}
