package reporter

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger = slog.New(handler)
}

func ReportError(msg string, err error) {
	if err != nil {
		logger.Error(msg, "error", err)
	} else {
		logger.Error(msg)
	}
}

func ReportWarning(msg string, err error) {
	if err != nil {
		logger.Warn(msg, "error", err)
	} else {
		logger.Warn(msg)
	}
}

func ReportInfo(msg string) {
	logger.Info(msg)
}
