package command

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
)

// CommandStep represents a command to run.
type Step struct {
	Name    string   `yaml:"name"`    // The name of the step.
	Command string   `yaml:"command"` // The command to run.
	Args    []string `yaml:"args"`    // The arguments to pass to the command.
}

// Execute runs the command.
func (s Step) Execute(env map[string]string) error {
	var (
		envSlice = make([]string, 0, len(env))
		args     = slices.Clone(s.Args)
	)

	for k, v := range env {
		var envVar = fmt.Sprintf("%s=%s", k, v)
		envSlice = append(envSlice, envVar)
	}

	for i, arg := range args {
		args[i] = os.Expand(arg, func(key string) string {
			return env[key]
		})
	}

	var cmd = exec.Command(s.Command, args...)
	cmd.Env = envSlice
	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//	HideWindow: true,
	//}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
