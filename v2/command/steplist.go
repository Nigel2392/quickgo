package command

type StepList struct {
	Steps []*Step `yaml:"steps"`
}

func (l *StepList) Execute() error {
	if l == nil {
		return nil
	}
	for _, step := range l.Steps {
		if err := step.Execute(); err != nil {
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
