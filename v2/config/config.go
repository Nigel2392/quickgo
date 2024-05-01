package config

import "github.com/Nigel2392/quickgo/v2/command"

type (
	// Config represents the configuration for QuickGo.
	QuickGo struct {
		Host              string `yaml:"host"`              // The host to run the server on.
		Port              string `yaml:"port"`              // The port to run the server on.
		ProjectConfigName string `yaml:"projectConfigName"` // The name of the project config file.

		TLSKey  string `yaml:"privateKey"`  // The path to the TLS key.
		TLSCert string `yaml:"certificate"` // The path to the TLS certificate.
	}

	// Project represents the configuration for an individual project.
	Project struct {
		// The name of the project.
		Name string `yaml:"name"`

		// Optional context for project templates.
		Context string `yaml:"context"`

		// List of commands to run
		BeforeCopy command.List `yaml:"beforeCopy"`
		AfterCopy  command.List `yaml:"afterCopy"`
	}
)
