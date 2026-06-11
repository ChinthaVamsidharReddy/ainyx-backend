package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a production-ready Zap logger that writes JSON to stdout.
// Call logger.Sync() before the process exits.
func New() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	// ISO-8601 timestamps are easier to read in logs and grep.
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg.Build()
}
