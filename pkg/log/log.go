package log

import (
	"context"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
)

type logger struct {
	logr.Logger
	level zapcore.Level
}

var log = &logger{Logger: ctrl.Log}

type Logger interface {
	Error(err error, msg string, keysAndValues ...interface{})
	Warning(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	DPanic(msg string, keysAndValues ...interface{})
	Panic(msg string, keysAndValues ...interface{})
	WithName(name string) Logger
	GetLogger() logr.Logger
}

func (l *logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.V(int(zapcore.ErrorLevel)).Error(err, msg, keysAndValues...)
}

func (l *logger) Warning(msg string, keysAndValues ...interface{}) {
	l.V(int(zapcore.WarnLevel)).Info(msg, keysAndValues...)
}

func (l *logger) Debug(msg string, keysAndValues ...interface{}) {
	l.V(int(zapcore.DebugLevel)).Info(msg, keysAndValues...)
}

func (l *logger) DPanic(msg string, keysAndValues ...interface{}) {
	l.V(int(zapcore.DPanicLevel)).Info(msg, keysAndValues...)
}

func (l *logger) Panic(msg string, keysAndValues ...interface{}) {
	l.V(int(zapcore.PanicLevel)).Info(msg, keysAndValues...)
}

func (l *logger) WithName(name string) Logger {
	return &logger{Logger: l.Logger.WithName(name)}
}

func (l *logger) GetLogger() logr.Logger {
	return l.Logger
}

func FromContext(ctx context.Context) Logger {
	l, _ := logr.FromContext(ctx)

	return &logger{Logger: l}
}

func NewLogger(ctx context.Context) Logger {
	newLogger, _ := logr.FromContext(ctx)

	return &logger{Logger: newLogger}
}
