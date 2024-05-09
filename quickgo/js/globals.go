package js

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

var PanicOnError = false

func logConsole(args ...any) {
	fmt.Println(args...)
}

func Console() *JSConsole {
	return &JSConsole{
		Debug: logConsole,
		Log:   logConsole,
		Info:  logConsole,
		Warn:  logConsole,
		Error: logConsole,
		Fatal: logConsole,
	}
}

func JSON() *_json {
	return &_json{}
}

func Base64() *_base64 {
	return &_base64{}
}

type JSConsole struct {
	Debug func(...any) `json:"debug"`
	Log   func(...any) `json:"log"`
	Info  func(...any) `json:"info"`
	Warn  func(...any) `json:"warn"`
	Error func(...any) `json:"error"`
	Fatal func(...any) `json:"fatal"`
}

type _json struct{}

func (j *_json) Encode(v any) string {
	data, err := json.Marshal(v)
	must(err)
	return string(data)
}

func (j *_json) DecodeObject(s string) any {
	var v = make(map[string]any)
	json.Unmarshal([]byte(s), &v)
	return v
}

func (j *_json) DecodeArray(s string) []any {
	var v = make([]any, 0)
	json.Unmarshal([]byte(s), &v)
	return v
}

type _base64 struct{}

func (b *_base64) Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func (b *_base64) Decode(data string) any {
	var arr, err = base64.StdEncoding.DecodeString(data)
	if must(err) {
		return nil
	}
	return string(arr)
}

func must(err error) bool {
	if err != nil {
		if PanicOnError {
			panic(err)
		}
		return true
	}
	return false
}
