package config

import (
	"encoding/gob"
	"path/filepath"

	"github.com/Nigel2392/quickgo/v2/command"
	"github.com/Nigel2392/quickgo/v2/quickfs"
)

type (
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
		Context map[string]string `yaml:"context" json:"context"`

		// List of commands to run
		BeforeCopy *command.StepList `yaml:"beforeCopy" json:"beforeCopy"`
		AfterCopy  *command.StepList `yaml:"afterCopy" json:"afterCopy"`

		// // Variable delimiters for the project templates.
		// DelimLeft  string `yaml:"delimLeft" json:"delimLeft"`
		// DelimRight string `yaml:"delimRight" json:"delimRight"`

		// A list of files to exclude from the project in glob format.
		Exclude []string `yaml:"exclude" json:"exclude"` // (NYI)

		Root *quickfs.FSDirectory `yaml:"-"` // The root directory.
	}
)

func init() {
	gob.Register(&QuickGo{})
	gob.Register(&Project{})
}

// ExampleProjectConfig returns an example project configuration.
func ExampleProjectConfig() *Project {
	return &Project{
		Name: "my-project",
		Context: map[string]string{
			"Name": "My Project",
		},
		Exclude: []string{
			"*node_modules*",
			"*dist*",
			"./.git",
		},
		BeforeCopy: &command.StepList{
			Steps: []command.Step{
				{
					Name:    "install-deps",
					Command: "npm",
					Args:    []string{"install"},
				},
				{
					Name:    "build",
					Command: "npm",
					Args:    []string{"run", "build"},
				},
			},
		},
		AfterCopy: &command.StepList{
			Steps: []command.Step{
				{
					Name:    "build",
					Command: "npm",
					Args:    []string{"run", "build"},
				},
			},
		},
	}
}

func (p *Project) IsExcluded(fl quickfs.FileLike) bool {
	for _, pattern := range p.Exclude {
		if m, err := filepath.Match(pattern, fl.GetPath()); err != nil {
			return false
		} else if m {
			return true
		}
	}
	return false
}
