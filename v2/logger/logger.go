package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type LogLevel int8

func (l LogLevel) String() string {
	return levelMap[l]
}

var levelMap = map[LogLevel]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
}

const (
	// DebugLevel is the lowest log level.
	DebugLevel LogLevel = iota

	// InfoLevel is the default log level.
	InfoLevel

	// WarnLevel is used for warnings.
	WarnLevel

	// ErrorLevel is used for errors.
	ErrorLevel
)

type LogWriter struct {
	Logger *Logger
	Level  LogLevel
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	if lw.Level >= lw.Logger.Level {
		lw.Logger.writePrefix(lw.Level, lw.Logger.Output)
		n, err = lw.Logger.Output.Write(p)
		lw.Logger.writeSuffix(lw.Logger.Output)
	}
	return
}

type Logger struct {
	// Level is the log level.
	Level LogLevel

	// Prefix is the prefix for each log message.
	Prefix string

	// Suffix is the suffix for each log message.
	Suffix string

	// Output is the output writer.
	Output io.Writer

	// WrapPrefix determines how the prefix should be wrapped
	// based on the LogLevel.
	WrapPrefix func(LogLevel, string) string
}

func (l *Logger) SetLevel(level LogLevel) {
	l.Level = level
}

func (l *Logger) Copy() *Logger {
	return &Logger{
		Level:      l.Level,
		Output:     l.Output,
		Prefix:     l.Prefix,
		WrapPrefix: l.WrapPrefix,
	}
}

func (l *Logger) Writer(level LogLevel) io.Writer {
	return &LogWriter{
		Logger: l.Copy(),
		Level:  level,
	}
}

func (l *Logger) PWriter(label string, level LogLevel) io.Writer {

	if l.Prefix != "" {
		label = fmt.Sprintf("%s / %s", l.Prefix, label)
	}

	var lw = &LogWriter{
		Logger: l.Copy(),
		Level:  level,
	}
	lw.Logger.Prefix = label
	return lw
}

func (l *Logger) Debug(args ...interface{}) {
	if l.Level <= DebugLevel {
		l.log(DebugLevel, args...)
	}
}

func (l *Logger) Info(args ...interface{}) {
	if l.Level <= InfoLevel {
		l.log(InfoLevel, args...)
	}
}

func (l *Logger) Warn(args ...interface{}) {
	if l.Level <= WarnLevel {
		l.log(WarnLevel, args...)
	}
}

func (l *Logger) Error(args ...interface{}) {
	if l.Level <= ErrorLevel {
		l.log(ErrorLevel, args...)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.Level <= DebugLevel {
		l.logf(DebugLevel, format, args...)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	if l.Level <= InfoLevel {
		l.logf(InfoLevel, format, args...)
	}
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.Level <= WarnLevel {
		l.logf(WarnLevel, format, args...)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.Level <= ErrorLevel {
		l.logf(ErrorLevel, format, args...)
	}
}

// Fatal is a convenience function for logging an error and exiting the program.
func (l *Logger) Fatal(errorcode int, args ...interface{}) {
	l.Error(args...)
	os.Exit(errorcode)
}

// Fatalf is a convenience function for logging an error and exiting the program.
func (l *Logger) Fatalf(errorcode int, format string, args ...interface{}) {
	l.Errorf(format, args...)
	os.Exit(errorcode)
}

func (l *Logger) Log(level LogLevel, args ...interface{}) {
	l.log(level, args...)
}

func (l *Logger) writePrefix(level LogLevel, w io.Writer) {
	var b = new(bytes.Buffer)

	_, _ = b.Write([]byte("["))
	if l.Prefix != "" {
		_, _ = b.Write([]byte(l.Prefix))
		_, _ = b.Write([]byte(" / "))
	}

	_, _ = b.Write([]byte(level.String()))
	_, _ = b.Write([]byte("]: "))

	var prefix = b.String()
	if l.WrapPrefix != nil {
		prefix = l.WrapPrefix(level, prefix)
	}

	_, _ = w.Write([]byte(prefix))
}

func (l *Logger) writeSuffix(w io.Writer) {
	if l.Suffix != "" {
		_, _ = w.Write([]byte(" "))
		_, _ = w.Write([]byte(l.Suffix))
	}
}

func (l *Logger) log(level LogLevel, args ...interface{}) {
	if l.Output == nil {
		return
	}

	var b = new(bytes.Buffer)
	l.writePrefix(level, b)
	fmt.Fprint(b, args...)
	l.writeSuffix(b)

	var message = b.String()
	if l.WrapPrefix != nil {
		message = l.WrapPrefix(level, message)
	}

	_, _ = l.Output.Write(
		[]byte(message),
	)

	_, _ = l.Output.Write([]byte("\n"))
}

func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	l.log(level, fmt.Sprintf(format, args...))
}
