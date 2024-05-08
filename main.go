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
	"strconv"
	"strings"

	quickgo "github.com/Nigel2392/quickgo/v2/quickgo"
	"github.com/Nigel2392/quickgo/v2/quickgo/config"
	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
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

	// Used to pass in the quickgo template
	Save bool

	// Used to pass in the quickgo template
	Use string

	// List the projects available for use
	ListProjects bool

	// List the commands available for all projects
	ListCommands bool

	// Save a global command for this user.
	SaveCommand string

	// Write an example project configuration
	Example bool

	// Serve the project over HTTP
	Serve bool

	// Lock the project configuration.
	// 1: Lock the project configuration.
	// 0: Unlock the project configuration.
	Lock int
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

	logger.Setup(&logger.Logger{
		Level:      logger.InfoLevel,
		Prefix:     "quickgo",
		OutputTime: true,
		WrapPrefix: quickgo.ColoredLogWrapper,
	})

	logger.SetOutput(
		logger.OutputAll,
		quickgo.Logfile(os.Stdout),
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
	flagSet.BoolVar(&flagger.Save, "save", false, "Import the project from the current directory.")
	flagSet.StringVar(&flagger.Use, "use", "", "Use the specified project configuration.")
	flagSet.BoolVar(&flagger.Example, "example", false, "Print an example project configuration.")
	flagSet.BoolVar(&flagger.ListProjects, "list", false, "List the projects available for use.")
	flagSet.BoolVar(&flagger.ListCommands, "list-commands", false, "List the commands available for all projects.")
	flagSet.StringVar(&flagger.SaveCommand, "save-command", "", "Save a global command for this user by providing a path to a JS file.")
	flagSet.BoolVar(&flagger.Serve, "serve", false, "Serve the project over HTTP.")
	flagSet.IntVar(&flagger.Lock, "lock", -1, "Lock the project configuration. 1=Lock, 0=Unlock.")
	flagSet.BoolFunc("v", "Enable verbose logging.", enableVerboseLogging)

	flagSet.Usage = func() {
		fmt.Println(quickgo.Craft(quickgo.CMD_Cyan, "QuickGo: A simple project generator and server."))
		fmt.Println("Usage: quickgo [-flags | exec <command> | <project-command>] [?args]")
		fmt.Println("Available application flags:")
		flagSet.VisitAll(func(f *flag.Flag) {

			var name = f.Name
			if f.DefValue != "" {
				name = fmt.Sprintf("%s=%s", name, f.DefValue)
			}

			fmt.Printf(
				"  -%s: %s\n",
				quickgo.BuildColorString(
					quickgo.CMD_Cyan,
					quickgo.CMD_Bold,
					name,
				),
				f.Usage,
			)
		})

		var commands, err = qg.ListJSFiles()
		if err != nil {
			logger.Warn(1, fmt.Errorf("failed to list commands: %w", err))
		}

		if len(commands) > 0 {
			fmt.Println(
				quickgo.Craft(quickgo.CMD_Blue, "Available commands:"),
			)
			for _, cmd := range commands {
				fmt.Printf("  - %s\n", quickgo.Craft(
					quickgo.CMD_Cyan, cmd,
				))
			}
		}

		// Try to load the project configuration.
		// It might contain some more commands! :D
		if qg.ProjectConfig == nil {
			err = qg.LoadCurrentProject(".")
		}
		if err != nil && errors.Is(err, config.ErrProjectMissing) {
			// No project found in the current directory.
			fmt.Println(quickgo.Craft(quickgo.CMD_Red, "No project found in the current directory."))
			fmt.Println("Run 'quickgo -example' to create an example project configuration.")
		} else if err == nil {

			// Project found, commands is map -> sort to slice.
			var commands = make([]*config.ProjectCommand, 0, len(qg.ProjectConfig.Commands))
			for _, v := range qg.ProjectConfig.Commands {
				commands = append(commands, v)
			}

			slices.SortFunc(commands, func(a, b *config.ProjectCommand) int {
				return strings.Compare(a.Name, b.Name)
			})

			if len(commands) == 0 {
				fmt.Println(quickgo.Craft(
					quickgo.CMD_Yellow,
					"No commands found in the project.",
				))
				return
			}

			fmt.Println(
				quickgo.Craft(
					quickgo.CMD_Blue,
					"Available project commands:",
				),
			)
			for _, c := range commands {
				if c.Description == "" {
					fmt.Printf("  - %s\n", quickgo.Craft(quickgo.CMD_Cyan, c.Name))
					continue
				}
				fmt.Printf("  - %s: %s\n", quickgo.Craft(quickgo.CMD_Cyan, c.Name), c.Description)
			}
		}
	}

	quickgo.PrintLogo()

	if len(os.Args) < 2 {
		logger.Fatal(1, "no command provided, run 'quickgo -h' for more information.")
	}

	// Initially load the application.
	qg, err = quickgo.LoadApp()
	if err != nil {
		logger.Fatal(1, err)
	}

	err = flagSet.Parse(os.Args[1:])
	if err != nil {
		logger.Fatal(1, err)
	}

	switch {
	case flagger.Save:

		if flagger.TargetDir == "" {
			flagger.TargetDir = "."
		}

		err = qg.LoadCurrentProject(flagger.TargetDir)
		if err != nil {
			fmt.Println(err)
			if !errors.Is(err, config.ErrProjectMissing) {
				logger.Fatal(1, fmt.Errorf("failed to read project config: %w", err))
			}

			var ctx = parseCommandlineContext(flagSet.Args(), false)
			var abs, _ = filepath.Abs(flagger.TargetDir)

			_, err = qg.NewProject(quickgo.SimpleProject{
				Name:       filepath.Base(abs),
				DelimLeft:  flagger.Project.DelimLeft,
				DelimRight: flagger.Project.DelimRight,
				Context:    ctx,
				Exclude:    flagger.Exclude,
			})
			if err != nil {
				logger.Fatal(1, fmt.Errorf("failed to create project: %w", err))
			}
		}

		err = qg.ProjectConfig.Load(flagger.TargetDir)
		if err != nil {
			logger.Fatal(1, err)
		}

		flagger.CopyProject(
			qg.ProjectConfig,
		)
		flagger.CopyConfig(
			qg.Config,
		)

		if err = qg.ProjectConfig.Validate(); err != nil {
			logger.Fatal(1, fmt.Errorf("failed to validate project config: %w", err))
		}

		err = qg.WriteProjectConfig(qg.ProjectConfig)
		if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to write project config: %w", err))
		}
	case flagger.Use != "":

		// Parse optional extra context provided by CLI arguments.
		var (
			ctx              = parseCommandlineContext(flagSet.Args(), false)
			proj, close, err = qg.ReadProjectConfig(flagger.Use)
		)
		if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to read project config: %w", err))
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
			logger.Fatal(1, fmt.Errorf("failed to write project: %w", err))
		}
	case flagger.Example:

		var example = config.ExampleProjectConfig()

		flagger.CopyProject(
			example,
		)

		if err = config.IsLocked(flagger.TargetDir); err != nil {
			logger.Fatal(1, err)
		}

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
			logger.Fatal(1, fmt.Errorf("failed to write example project config: %w", err))
		}
	case flagger.ListProjects:

		var projects, err = qg.ListProjects()
		if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to list projects: %w", err))
		}

		fmt.Println(quickgo.Craft(quickgo.CMD_Red, "Projects:"))
		for _, proj := range projects {
			fmt.Printf("  - %s\n", quickgo.Craft(
				quickgo.CMD_Blue, proj,
			))
		}

	case flagger.Lock == 1 || flagger.Lock == 0:

		if err = qg.LoadCurrentProject(flagger.TargetDir); err != nil && errors.Is(err, config.ErrProjectMissing) {
			logger.Fatal(1, "Cannot lock/unlock project outside of a project.")
		} else if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to read project config: %w", err))
		}

		var action string

		if flagger.Lock == 1 {
			err = config.LockProject(flagger.TargetDir)
			action = "lock"
		} else {
			err = config.UnlockProject(flagger.TargetDir)
			action = "unlock"
		}

		if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to %s project: %w", action, err))
		}

		logger.Infof("Project was %sed.", action)

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
			Handler: qg.HttpHandler(),
		}

		if qg.Config.TLSKey != "" && qg.Config.TLSCert != "" {
			logger.Infof("Serving on https://%s", addr)
			err = server.ListenAndServeTLS(
				qg.Config.TLSCert,
				qg.Config.TLSKey,
			)
		} else {
			logger.Infof("Serving on http://%s", addr)
			err = server.ListenAndServe()
		}

		if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to start server: %w", err))
		}

	case flagger.ListCommands:

		var commands, err = qg.ListJSFiles()
		if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to list commands: %w", err))
		}

		if len(commands) == 0 {
			fmt.Println(quickgo.Craft(quickgo.CMD_Yellow, "No commands found."))
			return
		}

		fmt.Println(quickgo.Craft(quickgo.CMD_Red, "Commands:"))
		for _, cmd := range commands {
			fmt.Printf("  - %s\n", quickgo.Craft(
				quickgo.CMD_Blue, cmd,
			))
		}

	case flagger.SaveCommand != "":

		var err = qg.SaveJS(flagger.SaveCommand)
		if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to save command: %w", err))
		}

	default:
		// Parse commands for the project itself.
		var args = flagSet.Args()
		if len(args) == 0 {
			flagSet.Usage()
			os.Exit(1)
		}

		// Execute app JS files if the command is 'exec'.
		if args[0] == "exec" && len(args) > 1 {
			var (
				fn  = args[1]
				ctx = parseCommandlineContext(args[2:], true)
			)

			if err = qg.LoadCurrentProject(flagger.TargetDir); err != nil {
				logger.Warnf("failed to read project config: %s", err)
			}

			err = qg.ExecJS(
				flagger.TargetDir, fn, ctx,
			)

			if err != nil {
				logger.Fatal(1, err)
			}

			return
		} else if args[0] == "exec" {
			logger.Fatal(1, "no function provided to execute")
		}

		// Look for commands for this project specifically
		var (
			cmd     *config.ProjectCommand
			command = args[0]
			ctx     = parseCommandlineContext(args[1:], true)
			err     = qg.LoadCurrentProject(flagger.TargetDir)
		)
		if err != nil && !errors.Is(err, config.ErrProjectMissing) {
			logger.Fatal(1, fmt.Errorf("failed to read project config: %w", err))
		} else if err != nil {
			logger.Fatal(1, "Cannot execute project commands outside of a project.")
		}

		cmd, err = qg.ProjectConfig.Command(command, nil)
		if err != nil && errors.Is(err, config.ErrCommandMissing) {
			logger.Fatal(1, fmt.Errorf("command '%s' not found", command))
		} else if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to get command: %w", err))
		}

		err = cmd.Execute(ctx)
		if err != nil {
			logger.Fatal(1, fmt.Errorf("failed to execute command: %w", err))
		}
	}
}

func enableVerboseLogging(b string) error {
	var boolVal, err = strconv.ParseBool(b)
	if err != nil {
		return err
	}

	if boolVal {
		logger.Info("Enabling verbose logging.")
		logger.SetLevel(logger.DebugLevel)
	} else {
		logger.SetLevel(logger.InfoLevel)
	}

	return nil
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
