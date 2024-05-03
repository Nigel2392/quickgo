package main

import (
	"errors"
	"flag"
	"fmt"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	quickgo "github.com/Nigel2392/quickgo/v2"
	"github.com/Nigel2392/quickgo/v2/config"
	"github.com/Nigel2392/quickgo/v2/logger"
)

type Flagger struct {
	// Optional overrides for the project.
	Project config.Project

	// Optional overrides for the config.
	Config config.QuickGo

	// Files to exclude from the project.
	Exclude arrayFlags

	// The target directory to write the project to.
	TargetDir string

	// Log verbose output
	Verbose bool

	// Used to pass in the quickgo template
	Import string

	// Used to pass in the quickgo template
	Use string

	// List the projects available for use
	ListProjects bool

	// Write an example project configuration
	Example bool

	// Serve the project over HTTP
	Serve bool
}

func (f *Flagger) CopyProject(proj *config.Project) {
	if f.Project.Name != "" {
		proj.Name = f.Project.Name
	}
	if f.Project.DelimLeft != "" {
		proj.DelimLeft = f.Project.DelimLeft
	}
	if f.Project.DelimRight != "" {
		proj.DelimRight = f.Project.DelimRight
	}
	if f.Project.Exclude != nil {
		proj.Exclude = f.Project.Exclude
	}
}

func (f *Flagger) CopyConfig(conf *config.QuickGo) {
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

	flagSet.StringVar(&flagger.Project.Name, "name", "", "The name of the project.")
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
	flagSet.BoolVar(&flagger.Example, "example", false, "Print an example project configuration.")
	flagSet.BoolVar(&flagger.ListProjects, "list", false, "List the projects available for use.")
	flagSet.BoolVar(&flagger.Serve, "serve", false, "Serve the project over HTTP.")
	flagSet.BoolVar(&flagger.Verbose, "v", false, "Enable verbose logging.")

	flagSet.Usage = func() {
		fmt.Println(quickgo.Craft(quickgo.CMD_Cyan, "QuickGo: A simple project generator and server."))
		fmt.Println("Usage: quickgo [flags] [?command] [?args]")
		fmt.Println("Available application flags:")
		flagSet.VisitAll(func(f *flag.Flag) {

			var name = f.Name
			if f.DefValue != "" {
				name = fmt.Sprintf("%s=%s", name, f.DefValue)
			}

			fmt.Printf("  -%s\n", quickgo.Craft(quickgo.CMD_Cyan, name))
			fmt.Printf("    %s: %s\n", f.Name, f.Usage)
		})

		// Try to load the project configuration.
		// It might contain some more commands! :D
		if qg.ProjectConfig == nil {
			err = qg.LoadProjectConfig(".")
		}
		if err != nil && errors.Is(err, config.ErrProjectMissing) {
			// No project found in the current directory.
			fmt.Println(quickgo.Craft(quickgo.CMD_Red, "No project found in the current directory."))
			fmt.Println("Run 'quickgo -example' to create an example project configuration.")
		} else if err == nil {

			// Project found, commands is map -> sort to slice.
			var commands = make([]string, 0, len(qg.ProjectConfig.Commands))
			for k := range qg.ProjectConfig.Commands {
				commands = append(commands, k)
			}
			slices.Sort(commands)

			fmt.Println(
				quickgo.Craft(
					quickgo.CMD_Blue,
					"Available project commands:",
				),
			)
			for _, c := range commands {
				fmt.Printf("  - %s\n", quickgo.Craft(quickgo.CMD_Cyan, c))
			}
		}
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

	quickgo.PrintLogo()

	if len(os.Args) < 2 {
		flagSet.Usage()
		os.Exit(1)
	}

	err = flagSet.Parse(os.Args[1:])
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

		flagger.CopyProject(
			qg.ProjectConfig,
		)
		flagger.CopyConfig(
			qg.Config,
		)

		qg.ProjectConfig.Name = flagger.Import

		err = qg.WriteProjectConfig(qg.ProjectConfig)
		if err != nil {
			panic(fmt.Errorf("failed to write project config: %w", err))
		}
	case flagger.Use != "":

		// Parse optional extra context provided by CLI arguments.
		var (
			ctx              = parseCommandlineContext(flagSet.Args(), false)
			proj, close, err = qg.ReadProjectConfig(flagger.Use)
		)
		if err != nil {
			panic(fmt.Errorf("failed to read project config: %w", err))
		}

		// Copy over the CLI context to the project context.
		if proj.Context == nil {
			proj.Context = ctx
		} else {
			maps.Copy(proj.Context, ctx)
		}

		defer close()

		flagger.CopyProject(
			proj,
		)

		err = qg.WriteProject(proj, flagger.TargetDir, false)
		if err != nil {
			panic(fmt.Errorf("failed to write project: %w", err))
		}
	case flagger.Example:

		var example = config.ExampleProjectConfig()

		flagger.CopyProject(
			example,
		)

		if s, err := os.Stat(filepath.Join(flagger.TargetDir, config.PROJECT_CONFIG_NAME)); err == nil {
			var abs, _ = filepath.Abs(s.Name())
			fmt.Printf("Project configuration file already exists at '%s'\n", abs)
			var overwrite string
			for overwrite != "y" && overwrite != "n" {
				fmt.Print("Overwrite? [y/n]: ")
				fmt.Scanln(&overwrite)
			}
			if overwrite == "n" {
				os.Exit(1)
			}
		}

		err = config.WriteYaml(
			example,
			filepath.Join(
				flagger.TargetDir,
				config.PROJECT_CONFIG_NAME,
			),
		)

		if err != nil {
			panic(fmt.Errorf("failed to write example project config: %w", err))
		}
	case flagger.ListProjects:

		var projects, err = qg.ListProjects()
		if err != nil {
			panic(fmt.Errorf("failed to list projects: %w", err))
		}

		fmt.Println(quickgo.Craft(quickgo.CMD_Red, "Projects:"))
		for _, proj := range projects {
			fmt.Printf("  - %s\n", quickgo.Craft(
				quickgo.CMD_Blue, proj,
			))
		}

	case flagger.Serve:

		flagger.CopyConfig(
			qg.Config,
		)

		var addr = fmt.Sprintf(
			"%s:%s",
			qg.Config.Host,
			qg.Config.Port,
		)
		var server = &http.Server{
			Addr:    addr,
			Handler: qg,
		}

		if qg.Config.TLSKey != "" && qg.Config.TLSCert != "" {
			fmt.Printf("Serving on https://%s\n", addr)
			err = server.ListenAndServeTLS(
				qg.Config.TLSCert,
				qg.Config.TLSKey,
			)
		} else {
			fmt.Printf("Serving on http://%s\n", addr)
			err = server.ListenAndServe()
		}

		if err != nil {
			panic(fmt.Errorf("failed to start server: %w", err))
		}
	default:
		// Parse commands for the project itself.
		var args = flagSet.Args()
		if len(args) == 0 {
			flagSet.Usage()
			os.Exit(1)
		}

		var (
			cmd     *config.ProjectCommand
			command = args[0]
			ctx     = parseCommandlineContext(args[1:], true)
			err     = qg.LoadProjectConfig(".")
		)
		if err != nil && !errors.Is(err, config.ErrProjectMissing) {
			panic(fmt.Errorf("failed to read project config: %w", err))
		} else if errors.Is(err, config.ErrProjectMissing) {
			fmt.Println(quickgo.Craft(quickgo.CMD_Red, "Cannot execute project commands outside of a project."))
			os.Exit(1)
		}

		cmd, err = qg.ProjectConfig.Command(command, nil)
		if err != nil && errors.Is(err, config.ErrCommandMissing) {

			// Just a bunch of terminal printing.
			flagSet.Usage()
			fmt.Printf(quickgo.Craft(quickgo.CMD_Red, "Command '%s' not found\n"), command)
			os.Exit(1)

		} else if err != nil {
			panic(fmt.Errorf("failed to get command: %w", err))
		}

		err = cmd.Execute(ctx)
		if err != nil {
			panic(err)
		}
	}
}

func parseCommandlineContext(args []string, parseCtxImmediately bool) map[string]any {
	var ctx = make(map[string]any)
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if parseCtxImmediately {
			var args = strings.Split(arg, "=")
			if len(args) == 2 {
				ctx[args[0]] = args[1]
			} else {
				ctx[arg] = true
			}
			continue
		}
		if arg == "/" {
			parseCtxImmediately = true
			continue
		}
	}
	return ctx
}
