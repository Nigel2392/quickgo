package command

import "fmt"

type Error struct {
	// The error message.
	Message string

	// An optional error that caused this error.
	Err error

	// The step or steplist that caused the error.
	Step     *Step
	Steplist *StepList
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}

	return e.Message
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Err
}
