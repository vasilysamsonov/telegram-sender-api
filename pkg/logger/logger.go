package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

type Logger struct {
	logger *zerolog.Logger
}

var _ Interface = (*Logger)(nil)

func New(level string) *Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		l = zerolog.ErrorLevel
	case "warn":
		l = zerolog.WarnLevel
	case "info":
		l = zerolog.InfoLevel
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(l)

	skipFrameCount := 3
	logger := zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	return &Logger{
		logger: &logger,
	}
}

func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg(zerolog.DebugLevel, message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.log(zerolog.InfoLevel, message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(zerolog.WarnLevel, message, args...)
}

func (l *Logger) Error(message interface{}, args ...interface{}) {
	l.msg(zerolog.ErrorLevel, message, args...)
}

func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg(zerolog.FatalLevel, message, args...)
	os.Exit(1)
}

func (l *Logger) log(level zerolog.Level, message string, args ...interface{}) {
	event := l.logger.WithLevel(level)
	if len(args) == 0 {
		event.Msg(message)
		return
	}

	event.Msgf(message, args...)
}

func (l *Logger) msg(level zerolog.Level, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(level, msg.Error(), args...)
	case string:
		l.log(level, msg, args...)
	default:
		l.log(level, fmt.Sprintf("%v", message), args...)
	}
}
