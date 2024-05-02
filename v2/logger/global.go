package logger

import "io"

var _globalLogger *Logger = &Logger{
	Level:  InfoLevel,
	Output: io.Discard,
}

func Global() *Logger {
	return _globalLogger
}

func Setup(l *Logger) {
	_globalLogger = l
}

func SetLevel(level LogLevel) {
	_globalLogger.Level = level
}

func Debug(args ...interface{}) {
	_globalLogger.Debug(args...)
}

func Info(args ...interface{}) {
	_globalLogger.Info(args...)
}

func Warn(args ...interface{}) {
	_globalLogger.Warn(args...)
}

func Error(args ...interface{}) {
	_globalLogger.Error(args...)
}

func Debugf(format string, args ...interface{}) {
	_globalLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	_globalLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	_globalLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	_globalLogger.Errorf(format, args...)
}
