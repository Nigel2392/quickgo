package logger_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
)

type LoggerTestStruct struct {
	b           bytes.Buffer
	LoggerLevel logger.LogLevel
	Level       logger.LogLevel
	Expected    string
	Input       string
}

var (
	outputBuffer    = new(bytes.Buffer)
	LoggerTestCases = []LoggerTestStruct{
		{LoggerLevel: logger.DebugLevel, Level: logger.DebugLevel, Expected: "[DEBUG]: test", Input: "test"},
		{LoggerLevel: logger.DebugLevel, Level: logger.InfoLevel, Expected: "[INFO]: test", Input: "test"},
		{LoggerLevel: logger.DebugLevel, Level: logger.WarnLevel, Expected: "[WARN]: test", Input: "test"},
		{LoggerLevel: logger.DebugLevel, Level: logger.ErrorLevel, Expected: "[ERROR]: test", Input: "test"},
		{LoggerLevel: logger.ErrorLevel, Level: logger.DebugLevel, Expected: "", Input: "test"},
		{LoggerLevel: logger.ErrorLevel, Level: logger.InfoLevel, Expected: "", Input: "test"},
		{LoggerLevel: logger.ErrorLevel, Level: logger.WarnLevel, Expected: "", Input: "test"},
		{LoggerLevel: logger.ErrorLevel, Level: logger.ErrorLevel, Expected: "[ERROR]: test", Input: "test"},
	}
)

func init() {
	logger.Setup(&logger.Logger{
		Level:       logger.WarnLevel,
		OutputDebug: outputBuffer,
		OutputInfo:  outputBuffer,
		OutputWarn:  outputBuffer,
		OutputError: outputBuffer,
	})
}

func Difference(a, b string) (string, string, string) {
	var diff string
	var i int
	for i = 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			diff += "^"
		} else {
			diff += " "
		}
	}
	for ; i < len(a); i++ {
		diff += "^"
		diff += " character (a): " + fmt.Sprint(rune(a[i]))
	}
	for ; i < len(b); i++ {
		diff += "^"
		diff += " character (b): " + fmt.Sprint(rune(b[i]))
	}
	return diff, a, b

}

func TestLogWriter(t *testing.T) {
	for _, test := range LoggerTestCases {
		var log = &logger.Logger{
			Level:       test.LoggerLevel,
			OutputDebug: &test.b,
			OutputInfo:  &test.b,
			OutputWarn:  &test.b,
			OutputError: &test.b,
		}
		var lw = log.Writer(test.Level)
		test.b.Reset()

		lw.Write([]byte(test.Input))

		if test.b.String() != test.Expected {
			t.Errorf("LogWriter.Write failed: %s != %s", test.b.String(), test.Expected)
		}
	}
}

func TestLogger(t *testing.T) {
	for _, test := range LoggerTestCases {
		var log = &logger.Logger{
			Level:       test.LoggerLevel,
			OutputDebug: &test.b,
			OutputInfo:  &test.b,
			OutputWarn:  &test.b,
			OutputError: &test.b,
		}
		test.b.Reset()

		switch test.Level {
		case logger.DebugLevel:
			log.Debug(test.Input)
		case logger.InfoLevel:
			log.Info(test.Input)
		case logger.WarnLevel:
			log.Warn(test.Input)
		case logger.ErrorLevel:
			log.Error(test.Input)
		}

		var expected = test.Expected
		if expected != "" {
			expected = fmt.Sprintf("%s\n", test.Expected)
		}
		if test.b.String() != expected {

			var diff, a, b string = Difference(
				test.b.String(), expected,
			)

			t.Errorf("Logger failed (%s <= %s): %s != %s\n%s\n%s\n%s", test.LoggerLevel, test.Level, test.b.String(), expected, a, b, diff)
		}
	}
}
