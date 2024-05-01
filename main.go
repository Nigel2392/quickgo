package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Nigel2392/quickgo/v2"
	"github.com/Nigel2392/quickgo/v2/logger"
)

func main() {

	logger.Setup(&logger.Logger{
		Level:  logger.DebugLevel,
		Output: os.Stdout,
		Prefix: "quickgo",
	})

	var qg, err = quickgo.LoadApp()
	if err != nil {
		panic(err)
	}

	err = qg.LoadProjectConfig()
	if err != nil {
		panic(err)
	}

	err = qg.ProjectConfig.Load(".")

	if errors.Is(err, quickgo.ErrProjectMissing) {
		err = qg.WriteExampleProjectConfig()
		if err != nil {
			panic(fmt.Errorf("failed to write example project config: %w", err))
		}
		return
	} else if err != nil {
		panic(err)
	}

	fmt.Println(qg)

	err = qg.WriteProjectConfig(qg.ProjectConfig)
	if err != nil {
		panic(fmt.Errorf("failed to write project config: %w", err))
	}

	fmt.Println("Project config written.")

	proj, close, err := qg.ReadProjectConfig(qg.ProjectConfig.Name)
	if err != nil {
		panic(fmt.Errorf("failed to read project config: %w", err))
	}

	defer close()

	err = qg.WriteProject(proj, "test")
	if err != nil {
		panic(fmt.Errorf("failed to write project: %w", err))
	}

}
