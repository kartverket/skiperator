/*
 * Thin wrapper for the controller-runtime logger.
 * Just to make it easier to log different levels
 */

package log

import (
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Logger interface {
	Error(err error, msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	WithName(name string) Logger
	GetLogger() logr.Logger
}

type logger struct {
	logr.Logger
}

func (l *logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.Logger.Error(err, msg, keysAndValues...)
}

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Info(msg, keysAndValues...)
}

func (l *logger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.V(1).Info(msg, keysAndValues...)
}

func (l *logger) WithName(name string) Logger {
	return &logger{Logger: l.Logger.WithName(name)}
}

func (l *logger) GetLogger() logr.Logger {
	return l.Logger
}

func NewLogger() Logger {
	return &logger{Logger: ctrl.Log}
}
