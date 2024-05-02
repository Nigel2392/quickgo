package config

import (
	"encoding/gob"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Nigel2392/quickgo/v2/command"
	"github.com/Nigel2392/quickgo/v2/logger"
	"github.com/Nigel2392/quickgo/v2/quickfs"
)

const (
	QUICKGO_DIR         = ".quickgo"     // The directory for QuickGo files, resides in the executable directory.
	QUICKGO_CONFIG_NAME = "quickgo.yaml" // Config file for QuickGo, resides in the executable directory.
	PROJECT_CONFIG_NAME = "quickgo.yaml" // Config file for the project, resides in the project (working) directory.
	PROJECT_ZIP_NAME    = "project.zip"  // The name of the project zip file.

	// Error messages.
	ErrProjectMissing = ErrorStr("project config not found")
)

type (
	ErrorStr string

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

		// Variable delimiters for the project templates.
		DelimLeft  string `yaml:"delimLeft" json:"delimLeft"`
		DelimRight string `yaml:"delimRight" json:"delimRight"`

		// A list of files to exclude from the project in glob format.
		Exclude []string `yaml:"exclude" json:"exclude"` // (NYI)

		// The root directory.
		Root *quickfs.FSDirectory `yaml:"-"`
	}
)

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
	}
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
