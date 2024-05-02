package logger

import (
	"fmt"
	"io"
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
	if lw.Level <= lw.Logger.Level {
		lw.Logger.writePrefix(lw.Level, lw.Logger.Output)
		n, err = lw.Logger.Output.Write(p)
		lw.Logger.writeSuffix()
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
}

func (l *Logger) SetLevel(level LogLevel) {
	l.Level = level
}

func (l *Logger) Write(p []byte) (n int, err error) {
	l.writePrefix(InfoLevel, l.Output)
	n, err = l.Output.Write(p)
	l.writeSuffix()
	return
}

func (l *Logger) Writer(level LogLevel) io.Writer {
	return &LogWriter{
		Logger: l,
		Level:  level,
	}
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

func (l *Logger) Log(level LogLevel, args ...interface{}) {
	l.log(level, args...)
}

func (l *Logger) writePrefix(level LogLevel, w io.Writer) {
	_, _ = w.Write([]byte("["))
	if l.Prefix != "" {
		_, _ = w.Write([]byte(l.Prefix))
		_, _ = w.Write([]byte(" / "))
	}

	_, _ = l.Output.Write([]byte(level.String()))
	_, _ = l.Output.Write([]byte("]: "))
}

func (l *Logger) writeSuffix() {
	if l.Suffix != "" {
		_, _ = l.Output.Write([]byte(" "))
		_, _ = l.Output.Write([]byte(l.Suffix))
	}
}

func (l *Logger) log(level LogLevel, args ...interface{}) {
	if l.Output == nil {
		return
	}

	l.writePrefix(level, l.Output)

	fmt.Fprint(l.Output, args...)

	l.writeSuffix()

	_, _ = l.Output.Write([]byte("\n"))
}

func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	l.log(level, fmt.Sprintf(format, args...))
}
