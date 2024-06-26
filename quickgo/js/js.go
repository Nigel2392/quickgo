package js

import (
	"maps"

	"github.com/dop251/goja"
	"github.com/pkg/errors"
)

type (
	Command struct {
		// The name of the main function to run.
		Main string

		// An override of a VM to use.
		_VM *goja.Runtime

		_Globals map[string]any
		_Funcs   []VMFunc
	}

	CommandError struct {
		Importance int
		Message    string
		Wrapped    error
	}

	CommandResult struct {
		Importance int    `json:"importance"`
		Message    string `json:"message"`
	}

	OptFunc func(*Command)
	VMFunc  func(*goja.Runtime) error
)

var (
	ErrExitCode    = errors.New("script returned with non-zero exit code")
	ErrMainMissing = errors.New("main function not found")
	ErrMainInvalid = errors.New("main function is invalid")
)

func (e *CommandError) Error() string {
	return e.Message
}

func (e *CommandError) Unwrap() error {
	return e.Wrapped
}

func (e *CommandError) ExitCode() int {
	return e.Importance
}

func WithGlobal(key string, value any) OptFunc {
	return func(s *Command) {
		s._Globals[key] = value
	}
}

func WithGlobals(globals map[string]any) OptFunc {
	return func(s *Command) {
		maps.Copy(s._Globals, globals)
	}
}

func WithConsole(s *Command) {
	s._Globals["console"] = Console()
}

func WithVM(vm *goja.Runtime) OptFunc {
	return func(s *Command) {
		s._VM = vm
	}
}

func NewScript(mainFunc string, options ...OptFunc) (cmd *Command) {
	var s = &Command{
		Main:     mainFunc,
		_Globals: make(map[string]any),
		_Funcs:   make([]VMFunc, 0),
	}

	for _, opt := range options {
		opt(s)
	}

	return s
}

func (s *Command) AddFunc(f ...VMFunc) {
	s._Funcs = append(s._Funcs, f...)
}

func newResult(importance int, message string) *CommandResult {
	return &CommandResult{
		Importance: importance,
		Message:    message,
	}
}

func (s *Command) Run(scriptSource string) (result *CommandResult, err error) {

	var vm *goja.Runtime
	if s._VM != nil {
		vm = s._VM
	} else {
		vm = goja.New()
	}

	vm.SetFieldNameMapper(
		goja.TagFieldNameMapper("json", true),
	)

	vm.Set("base64", Base64())
	vm.Set("Result", newResult)
	vm.Set("Success", func(message string) *CommandResult {
		return newResult(0, message)
	})
	vm.Set("Fail", func(message string) *CommandResult {
		return newResult(1, message)
	})
	vm.Set("GoError", func(message string) error {
		return &CommandError{
			Importance: 1,
			Message:    message,
		}
	})

	for k, v := range s._Globals {
		err = vm.Set(k, v)
		if err != nil {
			return nil, errors.Wrap(err, "could not add global to VM")
		}
	}

	for _, f := range s._Funcs {
		if err = f(vm); err != nil {
			return nil, errors.Wrap(err, "could not add function to VM")
		}
	}

	_, err = vm.RunString(scriptSource)
	if err != nil {
		return nil, errors.Wrap(err, "error running script")
	}

	var mainFunc = vm.Get(s.Main)
	if mainFunc == nil {
		return nil, ErrMainMissing
	}

	var main func() *CommandResult
	if err = vm.ExportTo(mainFunc, &main); err != nil {
		return nil, ErrMainInvalid
	}

	result = main()

	if result.Importance != 0 {
		return result, &CommandError{
			Importance: result.Importance,
			Wrapped:    ErrExitCode,
			Message:    result.Message,
		}
	}

	return result, nil
}
