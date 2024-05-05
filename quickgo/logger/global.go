package logger

import "io"

var _globalLogger *Logger = &Logger{
	Level:       InfoLevel,
	OutputDebug: io.Discard,
	OutputInfo:  io.Discard,
	OutputWarn:  io.Discard,
	OutputError: io.Discard,
}

func Global() *Logger {
	return _globalLogger
}

func Writer(level LogLevel) io.Writer {
	return _globalLogger.Writer(level)
}

func PWriter(label string, level LogLevel) io.Writer {
	return _globalLogger.PWriter(label, level)
}

func Setup(l *Logger) {
	_globalLogger = l
}

func SetOutput(level LogLevel, w io.Writer) {
	_globalLogger.SetOutput(level, w)
}

func Output(level LogLevel) io.Writer {
	return _globalLogger.Output(level)
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

func Fatal(errorCode int, args ...interface{}) {
	_globalLogger.Fatal(errorCode, args...)
}

func Fatalf(errorCode int, format string, args ...interface{}) {
	_globalLogger.Fatalf(errorCode, format, args...)
}
