package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
)

// CommandStep represents a command to run.
type Step struct {
	Name    string   `yaml:"name"`    // The name of the step.
	Command string   `yaml:"command"` // The command to run.
	Args    []string `yaml:"args"`    // The arguments to pass to the command.
}

// ParseArgs parses the arguments.
func (s *Step) ParseArgs(env map[string]any) ([]string, []string) {
	var (
		envSlice = make([]string, 0, len(env))
		args     = slices.Clone(s.Args)
	)

	for k, v := range env {
		var envVar = fmt.Sprintf("%s=%s", k, v)
		envSlice = append(envSlice, envVar)
	}

	for i, arg := range args {
		args[i] = ExpandArg(arg, env)
	}

	return args, envSlice
}

// Execute runs the command.
func (s Step) Execute(env map[string]any) error {
	var args, envSlice = s.ParseArgs(env)
	var cmd = exec.Command(s.Command, args...)
	cmd.Env = envSlice
	//cmd.SysProcAttr = &syscall.SysProcAttr{
	//	HideWindow: true,
	//}

	var (
		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)
	)

	// Writer for newlines
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", err, stderr.String())
	} else {
		if stdout.Len() > 0 {
			logger.Infof("%s", stdout.String())
		}

		if stderr.Len() > 0 {
			logger.Errorf("%s", stderr.String())
		}
	}
	return nil
}

// ExpandArgs expands the arguments.
func ExpandArg(arg string, env map[string]any) string {

	return os.Expand(arg, func(key string) string {
		var val, ok = env[key]
		if !ok {
			return ""
		}
		if s, ok := val.(string); ok {
			return s
		}
		var v, err = json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(v)
	})
}
