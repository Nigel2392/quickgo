package js

import "fmt"

func logConsole(args ...any) {
	fmt.Println(args...)
}

type JSConsole struct {
	Debug func(...any) `json:"debug"`
	Log   func(...any) `json:"log"`
	Info  func(...any) `json:"info"`
	Warn  func(...any) `json:"warn"`
	Error func(...any) `json:"error"`
}

func Console() *JSConsole {
	return &JSConsole{
		Debug: logConsole,
		Log:   logConsole,
		Info:  logConsole,
		Warn:  logConsole,
		Error: logConsole,
	}
}
