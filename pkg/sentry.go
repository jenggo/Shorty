package pkg

import (
	"bytes"
	"context"
	"io"
	"time"

	"shorty/config"
	"shorty/internal/version"

	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/attribute"
	sentryzerolog "github.com/getsentry/sentry-go/zerolog"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

// InitSentry configures the global Sentry client, which the log writer and the
// tracing middleware then share. It returns a zerolog writer that turns warn and
// above into grouped issues, with lower levels kept as breadcrumbs.
//
// A missing DSN is not an error: it returns (nil, nil) so the caller logs to the
// console only.
func InitSentry() (io.Writer, error) {
	if config.Use.App.SentryURL == "" {
		return nil, nil
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.Use.App.SentryURL,
		ServerName:       config.AppName,
		Release:          version.String(),
		EnableTracing:    true,
		TracesSampleRate: config.Use.App.SentryTracesSampleRate,
		BeforeSendLog:    WithoutLogTimestamp,
	}); err != nil {
		return nil, err
	}

	return sentryzerolog.NewWithHub(sentry.CurrentHub(), sentryzerolog.Options{
		Levels:          []zerolog.Level{zerolog.WarnLevel, zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel},
		WithBreadcrumbs: true,
	})
}

// sentryLogWriter mirrors zerolog entries into Sentry's structured log API, so
// every level stays searchable while the event writer keeps warn+ as grouped
// issues. Entries carry no request context, so they have no trace link.
type sentryLogWriter struct {
	logger sentry.Logger
}

// NewSentryLogWriter returns a zerolog writer that forwards log entries to
// Sentry as structured logs. Call it only after InitSentry, otherwise the
// logger falls back to a no-op.
func NewSentryLogWriter() *sentryLogWriter {
	return &sentryLogWriter{logger: sentry.NewLogger(context.Background())}
}

func (w *sentryLogWriter) Write(p []byte) (int, error) {
	fields, err := parseLogFields(p)
	if err != nil {
		return len(p), nil
	}

	message, _ := fields[zerolog.MessageFieldName].(string)
	entry := sentryLogEntry(w.logger, logLevel(fields))
	for key, value := range fields {
		if isLogMetadataField(key) {
			continue
		}
		entry = setLogAttribute(entry, key, value)
	}
	entry.Emit(message)

	return len(p), nil
}

// parseLogFields decodes a zerolog JSON entry. Numbers stay as json.Number so
// integers reach Sentry as int64 attributes instead of floats.
func parseLogFields(data []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var fields map[string]any
	if err := decoder.Decode(&fields); err != nil {
		return nil, err
	}

	return fields, nil
}

func isLogMetadataField(key string) bool {
	switch key {
	case zerolog.MessageFieldName, zerolog.LevelFieldName, zerolog.TimestampFieldName:
		return true
	default:
		return false
	}
}

func logLevel(fields map[string]any) zerolog.Level {
	raw, _ := fields[zerolog.LevelFieldName].(string)
	level, err := zerolog.ParseLevel(raw)
	if err != nil {
		return zerolog.InfoLevel
	}

	return level
}

// WithoutLogTimestamp clears the RFC3339 timestamp that sentry-go adds to log
// items, because RustRak only accepts epoch seconds and otherwise drops the log.
// The original timestamp is kept as an attribute.
func WithoutLogTimestamp(log *sentry.Log) *sentry.Log {
	if log == nil || log.Timestamp.IsZero() {
		return log
	}

	if log.Attributes == nil {
		log.Attributes = make(map[string]attribute.Value, 1)
	}
	log.Attributes["original_timestamp"] = attribute.StringValue(log.Timestamp.Format(time.RFC3339Nano))
	log.Timestamp = time.Time{}

	return log
}

// sentryLogEntry maps a zerolog level onto the Sentry log API. Fatal and panic
// use LFatal so the log is emitted without the writer taking over the exit.
func sentryLogEntry(logger sentry.Logger, level zerolog.Level) sentry.LogEntry {
	switch level {
	case zerolog.TraceLevel:
		return logger.Trace()
	case zerolog.DebugLevel:
		return logger.Debug()
	case zerolog.InfoLevel, zerolog.NoLevel, zerolog.Disabled:
		return logger.Info()
	case zerolog.WarnLevel:
		return logger.Warn()
	case zerolog.ErrorLevel:
		return logger.Error()
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return logger.LFatal()
	default:
		return logger.Info()
	}
}

func setLogAttribute(entry sentry.LogEntry, key string, value any) sentry.LogEntry {
	switch v := value.(type) {
	case string:
		return entry.String(key, v)
	case bool:
		return entry.Bool(key, v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return entry.Int64(key, i)
		}
		if f, err := v.Float64(); err == nil {
			return entry.Float64(key, f)
		}

		return entry.String(key, v.String())
	case nil:
		return entry
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return entry
		}

		return entry.String(key, string(encoded))
	}
}
