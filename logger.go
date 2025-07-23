package main

import (
	"fmt"
	"log/slog"
	"os"
)

type internalLogger struct {
	*slog.Logger
	level slog.Level
}

var logger *internalLogger

func setupLogger() {
	// Set log level based on environment
	level := slog.LevelInfo
	if os.Getenv("SPACELIFT_DEBUG") != "" {
		level = slog.LevelDebug
	}

	// Create logger with desired level
	slogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	logger = &internalLogger{
		Logger: slogger,
		level:  level,
	}
}

func (l *internalLogger) Fatalf(msg string, args ...any) {
	// Custom Fatalf to log error and exit
	l.Error(fmt.Sprintf(msg, args...))
	os.Exit(1)
}

func (l *internalLogger) Debugf(format string, args ...any) {
	// Custom Debugf to log debug messages
	if l.level >= slog.LevelDebug {
		l.Debug(fmt.Sprintf(format, args...))
	}
}
