package command

import "os/exec"

// CommandStep represents a command to run.
type CommandStep struct {
	StepName string   `yaml:"name"`    // The name of the step.
	Command  string   `yaml:"command"` // The command to run.
	Args     []string `yaml:"args"`    // The arguments to pass to the command.
}

// Name returns the name of the step.
func (s CommandStep) Name() string {
	return s.StepName
}

// Execute runs the command.
func (s CommandStep) Execute() error {
	var cmd = exec.Command(s.Command, s.Args...)
	return cmd.Run()
}
