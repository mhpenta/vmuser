package app

import (
	"log/slog"
	"time"
)

// Since logs the time since the start time, to be used ergonomically with defer.
func Since(start time.Time) {
	slog.Info("Finished", "time", time.Since(start))
}

// LogSince logs the time since the start time, to be used ergonomically with defer.
func LogSince(msg string, start time.Time) {
	slog.Info(msg, "time", time.Since(start))
}
