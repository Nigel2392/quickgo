package command

import "os/exec"

// CommandStep represents a command to run.
type Step struct {
	Name    string   `yaml:"name"`    // The name of the step.
	Command string   `yaml:"command"` // The command to run.
	Args    []string `yaml:"args"`    // The arguments to pass to the command.
}

// Execute runs the command.
func (s *Step) Execute() error {
	var cmd = exec.Command(s.Command, s.Args...)
	return cmd.Run()
}
