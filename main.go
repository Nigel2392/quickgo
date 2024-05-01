package main

import (
	"errors"
	"fmt"

	"github.com/Nigel2392/quickgo/v2"
)

func main() {
	var qg, err = quickgo.LoadApp()
	if err != nil {
		panic(err)
	}

	err = qg.LoadProjectConfig("test")
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

	err = qg.WriteProject(qg.ProjectConfig, "test")
	if err != nil {
		panic(fmt.Errorf("failed to write project: %w", err))
	}
}
