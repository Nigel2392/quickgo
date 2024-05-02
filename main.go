package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Nigel2392/quickgo/v2"
	"github.com/Nigel2392/quickgo/v2/config"
	"github.com/Nigel2392/quickgo/v2/logger"
)

type Flagger struct {
	Project   config.Project
	Config    config.QuickGo
	Exclude   arrayFlags
	TargetDir string
	Verbose   bool

	// Used to pass in the quickgo template
	Import string
	// Used to pass in the quickgo template
	Use string
}

func (f *Flagger) Copy(proj *config.Project, conf *config.QuickGo) {
	if f.Project.DelimLeft != "" {
		proj.DelimLeft = f.Project.DelimLeft
	}
	if f.Project.DelimRight != "" {
		proj.DelimRight = f.Project.DelimRight
	}
	if f.Project.Exclude != nil {
		proj.Exclude = f.Project.Exclude
	}
	if f.Config.Host != "" {
		conf.Host = f.Config.Host
	}
	if f.Config.Port != "" {
		conf.Port = f.Config.Port
	}
	if f.Config.TLSKey != "" {
		conf.TLSKey = f.Config.TLSKey
	}
	if f.Config.TLSCert != "" {
		conf.TLSCert = f.Config.TLSCert
	}
}

type arrayFlags []string

func (a *arrayFlags) String() string {
	var b strings.Builder
	for i, v := range *a {
		b.WriteString(v)
		if i < len(*a)-1 {
			b.WriteString(", ")
		}
	}
	return b.String()
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {

	var (
		err     error
		flagger Flagger
		flagSet = flag.NewFlagSet("quickgo", flag.ExitOnError)
		qg      *quickgo.App
	)

	flagSet.StringVar(&flagger.Project.DelimLeft, "delim-left", "", "The left delimiter for the project templates.")
	flagSet.StringVar(&flagger.Project.DelimRight, "delim-right", "", "The right delimiter for the project templates.")

	flagSet.StringVar(&flagger.Config.Host, "host", "localhost", "The host to run the server on.")
	flagSet.StringVar(&flagger.Config.Port, "port", "8080", "The port to run the server on.")
	flagSet.StringVar(&flagger.Config.TLSKey, "tls-key", "", "The path to the TLS key.")
	flagSet.StringVar(&flagger.Config.TLSCert, "tls-cert", "", "The path to the TLS certificate.")

	flagSet.Var(&flagger.Exclude, "e", "A list of files to exclude from the project in glob format.")
	flagSet.StringVar(&flagger.TargetDir, "d", "", "The target directory to write the project to.")
	flagSet.StringVar(&flagger.Import, "get", "", "Import the project from the current directory.")
	flagSet.StringVar(&flagger.Use, "use", "", "Use the specified project configuration.")
	flagSet.BoolVar(&flagger.Verbose, "v", false, "Enable verbose logging.")

	quickgo.PrintLogo()

	if len(os.Args) < 2 {
		flagSet.Usage()
		os.Exit(1)
	}

	err = flagSet.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	logger.Setup(&logger.Logger{
		Level:  logger.InfoLevel,
		Output: os.Stdout,
		Prefix: "quickgo",
	})

	if flagger.Verbose {
		logger.SetLevel(logger.DebugLevel)
	}

	// Initially load the application.
	qg, err = quickgo.LoadApp()
	if err != nil {
		panic(err)
	}

	switch {
	case flagger.Import != "":

		err = qg.LoadProjectConfig(".")
		if err != nil {
			panic(err)
		}

		err = qg.ProjectConfig.Load(".")
		if err != nil {
			panic(err)
		}

		flagger.Copy(
			qg.ProjectConfig,
			qg.Config,
		)

		qg.ProjectConfig.Name = flagger.Import

		err = qg.WriteProjectConfig(qg.ProjectConfig)
		if err != nil {
			panic(fmt.Errorf("failed to write project config: %w", err))
		}

		return

	case flagger.Use != "":

		var proj, close, err = qg.ReadProjectConfig(flagger.Use)
		if err != nil {
			panic(fmt.Errorf("failed to read project config: %w", err))
		}

		defer close()

		err = qg.WriteProject(proj, flagger.TargetDir, false)
		if err != nil {
			panic(fmt.Errorf("failed to write project: %w", err))
		}
	}

}
