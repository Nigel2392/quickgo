package command

import "github.com/Nigel2392/quickgo/v2/logger"

type StepList struct {
	Steps []Step `yaml:"steps"`
}

func (l *StepList) Execute(env map[string]string) error {
	if l == nil {
		return nil
	}
	for i, step := range l.Steps {
		logger.Infof("%d: %s", i, step.Name)
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
