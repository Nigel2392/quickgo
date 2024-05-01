package quickgo

import (
	"github.com/Nigel2392/quickgo/v2/config"
	"github.com/Nigel2392/quickgo/v2/quickfs"
)

type (
	App struct {
		Config        config.QuickGo    `yaml:"config"`        // The configuration for QuickGo.
		ProjectConfig config.Project    `yaml:"projectConfig"` // The configuration for the project.
		Root          quickfs.Directory `yaml:"-"`             // The root directory.
	}
)
