package config

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nigel2392/quickgo/v2/quickgo/command"
	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
	"github.com/Nigel2392/quickgo/v2/quickgo/quickfs"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	QUICKGO_DIR         = ".quickgo"     // The directory for QuickGo files, resides in the executable directory.
	QUICKGO_LOG_NAME    = "quickgo.log"  // The log file name.
	QUICKGO_CONFIG_NAME = "quickgo.yaml" // Config file for QuickGo, resides in the executable directory.
	PROJECT_CONFIG_NAME = "quickgo.yaml" // Config file for the project, resides in the project (working) directory.
	PROJECT_ZIP_NAME    = "project.zip"  // The name of the project zip file.
	LOCKFILE_NAME       = "quickgo.lock" // The lock file name.

	// Error messages.
	ErrCommandMissing = ErrorStr("command not found")
	ErrProjectMissing = ErrorStr("project config not found")
	ErrProjectExists  = ErrorStr("project config already exists")
	ErrProjectInvalid = ErrorStr("project config is invalid")
)

var (
	ErrProjectName = errors.Wrap(
		ErrProjectInvalid,
		"project name cannot be a blank string, period or filepath",
	)
)

func IsLocked(path string) error {
	if path == "" {
		path = "."
	}
	var lockfile = filepath.Join(path, LOCKFILE_NAME)
	var _, err = os.Stat(lockfile)
	if err == nil {
		return errors.Errorf("project is locked: '%s'", lockfile)
	}
	return nil
}

// Lock locks the project.
func LockProject(path string) error {
	if path == "" {
		path = "."
	}
	var lockfile = filepath.Join(path, LOCKFILE_NAME)
	var file, err = os.OpenFile(lockfile, os.O_CREATE|os.O_EXCL, 0666)
	if err != nil && os.IsExist(err) {
		return errors.Errorf("project is already locked: '%s'", lockfile)
	} else if err != nil {
		return errors.Wrapf(err, "error creating lock file '%s'", lockfile)
	}
	return file.Close()
}

// Unlock unlocks the project.
func UnlockProject(path string) (err error) {
	if path == "" {
		path = "."
	}
	var lockfile = filepath.Join(path, LOCKFILE_NAME)
	err = os.Remove(lockfile)
	if err != nil && os.IsNotExist(err) {
		return errors.Errorf("project is not locked: '%s'", lockfile)
	} else if err != nil {
		return errors.Wrapf(err, "error removing lock file '%s'", lockfile)
	}
	return nil
}

type (
	ErrorStr string

	Validator interface {
		Validate() error
	}

	// Config represents the configuration for QuickGo.
	QuickGo struct {
		Host    string `yaml:"host"`        // The host to run the server on.
		Port    string `yaml:"port"`        // The port to run the server on.
		TLSKey  string `yaml:"privateKey"`  // The path to the TLS key.
		TLSCert string `yaml:"certificate"` // The path to the TLS certificate.
	}

	// Project represents the configuration for an individual project.
	Project struct {
		// The name of the project.
		Name string `yaml:"name" json:"name"`

		// Optional context for project templates.
		Context map[string]any `yaml:"context" json:"context"`

		// List of commands to run
		BeforeCopy *command.StepList `yaml:"beforeCopy" json:"beforeCopy"`
		AfterCopy  *command.StepList `yaml:"afterCopy" json:"afterCopy"`
		Commands   ProjectCommandMap `yaml:"commands" json:"commands"` // [name] => [steps]

		// Variable delimiters for the project templates.
		DelimLeft  string `yaml:"delimLeft" json:"delimLeft"`
		DelimRight string `yaml:"delimRight" json:"delimRight"`

		// A list of files to exclude from the project in glob format.
		Exclude []string `yaml:"exclude" json:"exclude"` // (NYI)

		// The root directory.
		Root *quickfs.FSDirectory `yaml:"-"`
	}

	// ProjectCommand represents a command for a project.
	ProjectCommand struct {
		// The name of the command.
		// Only used internally for logging purposes.
		Name        string `yaml:"-" json:"-"`
		Description string `yaml:"description" json:"description"`
		// Args are the arguments to pass to the command.
		// These will be asked via stdin if not provided.
		Args map[string]any `yaml:"args" json:"args"`
		// The steps to run for the command.
		Steps *command.StepList `yaml:"steps" json:"steps"`
	}
)

type ProjectCommandMap map[string]*ProjectCommand

func (p *ProjectCommandMap) UnmarshalYAML(value *yaml.Node) error {
	var commands = make(map[string]*ProjectCommand)
	if err := value.Decode(&commands); err != nil {
		return err
	}
	for k, v := range commands {
		v.Name = k
	}
	*p = commands
	return nil
}

func (e ErrorStr) Error() string {
	return string(e)
}

func init() {
	gob.Register(&QuickGo{})
	gob.Register(&Project{})
}

// ExampleProjectConfig returns an example project configuration.
func ExampleProjectConfig() *Project {
	return &Project{
		Name: "my-project",
		Context: map[string]any{
			"Name": "My Project",
		},
		Exclude: []string{
			"*node_modules*",
			"*dist*",
			"./.git",
		},
		DelimLeft:  "${{",
		DelimRight: "}}",
		BeforeCopy: &command.StepList{
			Steps: []command.Step{
				{
					Name:    "Echo Project Name Before",
					Command: "echo",
					Args:    []string{"$projectName"},
				},
				{
					Name:    "Echo Project Path Before",
					Command: "echo",
					Args:    []string{"$projectPath"},
				},
			},
		},
		AfterCopy: &command.StepList{
			Steps: []command.Step{
				{
					Name:    "Echo Project Name After",
					Command: "echo",
					Args:    []string{"$projectName"},
				},
				{
					Name:    "Echo Project Path After",
					Command: "echo",
					Args:    []string{"$projectPath"},
				},
			},
		},
		Commands: ProjectCommandMap{
			"echoName": {
				Description: "Echo the project name, supply one with customProjectName=myValue",
				Args: map[string]any{
					"customProjectName": "$projectName",
				},
				Steps: &command.StepList{
					Steps: []command.Step{
						{
							Name:    "Echo Project Name",
							Command: "echo",
							Args:    []string{"$customProjectName"},
						},
					},
				},
			},
		},
	}
}

// Validate validates the project configuration.
func (p *Project) Validate() error {
	var name = strings.TrimPrefix(p.Name, ".")
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || name == "" {
		return ErrProjectName
	}
	return nil
}

func (p *Project) Command(name string, context map[string]any) (*ProjectCommand, error) {
	var cmd, ok = p.Commands[name]
	if !ok {
		return nil, ErrCommandMissing
	}

	if cmd.Args == nil {
		cmd.Args = make(map[string]any)
	}

	cmd.Name = name
	cmd.Args["projectName"] = p.Name
	cmd.Args["projectPath"], _ = os.Getwd()

	for k, v := range p.Context {
		cmd.Args[k] = v
	}

	for k, v := range context {
		cmd.Args[k] = v
	}

	return cmd, nil
}

// Execute executes the project commands.
func (p *Project) ExecCommand(commandName string, env map[string]any) error {
	var cmd, err = p.Command(commandName, env)
	if err != nil {
		return err
	}
	return cmd.Execute(env)
}

// Load loads the project configuration.
func (p *Project) Load(projectDirectory string) error {
	p.Root = quickfs.NewFSDirectory(
		fmt.Sprintf("%s .Name %s", p.DelimLeft, p.DelimRight),
		projectDirectory,
		nil,
	)

	p.Root.IsExcluded = p.IsExcluded

	return p.Root.Load()
}

func (p *Project) IsExcluded(fl quickfs.FileLike) bool {
	var path = fl.GetPath()

	if strings.HasSuffix(path, PROJECT_CONFIG_NAME) {
		return true
	}

	for _, pattern := range p.Exclude {
		var match, err = filepath.Match(pattern, path)
		if err != nil {
			logger.Errorf("Error matching pattern %s: %v", pattern, err)
			continue
		}

		if match {
			logger.Debugf("Excluding %s (%s)", path, pattern)
			return true
		}
	}
	return false
}

// Execute executes the project command.
func (c *ProjectCommand) Execute(env map[string]any) error {
	if c.Steps == nil {
		return nil
	}

	var newEnv = make(map[string]any)
	maps.Copy(newEnv, c.Args)
	maps.Copy(newEnv, env)

	for k, v := range newEnv {
		if s, ok := v.(string); ok {
			newEnv[k] = command.ExpandArg(s, newEnv)
		}
	}

	var jsonData, err = json.MarshalIndent(newEnv, "", "  ")
	if err == nil {
		logger.Debugf("Running command '%s' with environment: %s", c.Name, jsonData)
	} else {
		logger.Warnf("Error marshalling environment map for logging: %v", err)
	}

	return c.Steps.Execute(newEnv)
}
