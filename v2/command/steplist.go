package command

import "fmt"

type StepList struct {
	Steps []Step `yaml:"steps"`
}

func (l *StepList) Execute(env map[string]string) error {
	if l == nil {
		return nil
	}
	for i, step := range l.Steps {
		fmt.Printf("%d: %s\n", i, step.Name)
		if err := step.Execute(env); err != nil {
			return &Error{
				Message:  "failed to execute step",
				Err:      err,
				Step:     step,
				Steplist: l,
			}
		}
	}

	return nil
}
