package command

type StepList []Step

func (l StepList) Steps() []Step {
	return l
}

func (l StepList) Execute() error {
	for _, step := range l {
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
