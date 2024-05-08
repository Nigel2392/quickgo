package quickgo

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Nigel2392/goldcrest"
	"github.com/Nigel2392/quickgo/v2/quickgo/config"
	"github.com/Nigel2392/quickgo/v2/quickgo/js"
	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
	"github.com/Nigel2392/quickgo/v2/quickgo/quickfs"
	"github.com/dop251/goja"
	"github.com/pkg/errors"
)

var (
	cliApplication *App
)

type ansiStrippedWriter struct {
	io.Writer
}

// https://github.com/acarl005/stripansi
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func (w *ansiStrippedWriter) Write(p []byte) (n int, err error) {
	return w.Writer.Write([]byte(re.ReplaceAllString(string(p), "")))
}

// Output the log file to the given writer.
// If the writer is nil, the log file will be written to the default location.
// This function may panic if the log file cannot be opened.
func Logfile(output io.Writer) io.Writer {
	var (
		logPath = GetQuickGoPath(config.QUICKGO_LOG_NAME)
		err     error
		file    *os.File
		writer  io.Writer
	)

	file, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	if output != nil {
		writer = io.MultiWriter(output, &ansiStrippedWriter{file})
	} else {
		writer = &ansiStrippedWriter{file}
	}

	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c

		logger.Infof("Closing log file %s", logPath)
		err = file.Close()
		if err != nil {
			fmt.Printf("Failed to close log file %s\n", logPath)
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()

	return writer
}

type (
	App struct {
		Config        *config.QuickGo `yaml:"config"`        // The configuration for QuickGo.
		ProjectConfig *config.Project `yaml:"projectConfig"` // The configuration for the project.
		Patterns      []string        `yaml:"patterns"`      // The patterns for the templates.
		AppFS         fs.FS           `yaml:"-"`             // The file system for the app, resides in the userprofile home directory.
		ProjectFS     fs.FS           `yaml:"-"`             // The file system for the project, resides in the project (working) directory.
		logfile       io.Writer       `yaml:"-"`             // The log file.
	}
)

func LoadApp() (*App, error) {
	if cliApplication != nil {
		return cliApplication, nil
	}

	var wd, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	// Check for the application config directory.
	var quickGoDir = GetQuickGoPath()
	var projectDir = GetQuickGoPath(config.PROJECTS_DIR)
	_, err = os.Stat(projectDir)

	// Create the application config directory if it does not exist.
	if err != nil && os.IsNotExist(err) {

		logger.Infof("Creating new config directory %s", projectDir)

		if err = os.MkdirAll(projectDir, os.ModePerm); err != nil {
			return nil, err
		}

	} else if err != nil {
		return nil, err
	}

	// Initialize the application with the proper file systems.
	var app = &App{
		AppFS:     os.DirFS(quickGoDir),
		ProjectFS: os.DirFS(wd),
		Patterns: []string{
			"_templates/base.tmpl",
			"_templates/parent_url.tmpl",
		},
	}

	// Setup the global application to prevent multiple instances.
	cliApplication = app

	// Load the QuickGo configuration.
	var configPath = GetQuickGoPath(config.QUICKGO_CONFIG_NAME)
	cfg, err := config.LoadYamlFS[config.QuickGo](
		app.AppFS,
		config.QUICKGO_CONFIG_NAME,
	)
	if err != nil {
		// Create a new config file if it does not exist.
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "failed to load config file %s", configPath)
		}

		cfg = &config.QuickGo{
			Host: "localhost",
			Port: "8080",
		}

		logger.Infof("Writing new config file %s", configPath)

		if err = config.WriteYaml(cfg, configPath); err != nil {
			logger.Errorf("Failed to write config file %s", configPath)
			return nil, errors.Wrapf(err, "failed to write config file %s", configPath)
		}
	}

	app.Config = cfg

	for _, hook := range goldcrest.Get[AppHook](HookQuickGoLoaded) {
		if err = hook(app); err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (a *App) Logger(level logger.LogLevel, output io.Writer) *logger.Logger {
	var lg = &logger.Logger{
		Level: level,
	}

	lg.SetOutput(
		logger.OutputAll,
		Logfile(output),
	)

	return lg
}

type SimpleProject struct {
	Name       string                 `yaml:"name"`
	DelimLeft  string                 `yaml:"delimLeft"`
	DelimRight string                 `yaml:"delimRight"`
	Exclude    []string               `yaml:"exclude"`
	Context    map[string]interface{} `yaml:"context"`
}

func (a *App) NewProject(conf SimpleProject) (proj *config.Project, err error) {
	proj = &config.Project{
		Name:       conf.Name,
		DelimLeft:  conf.DelimLeft,
		DelimRight: conf.DelimRight,
		Context:    conf.Context,
		Exclude:    conf.Exclude,
	}

	if err = proj.Validate(); err != nil {
		return nil, err
	}

	a.ProjectConfig = proj

	logger.Infof("Initialized new project config for '%s'", proj.Name)

	for _, hook := range goldcrest.Get[ProjectHook](HookProjectLoaded) {
		if err = hook(a, proj); err != nil {
			return nil, err
		}
	}

	var dirPath = GetProjectDirectoryPath(proj.Name, true)
	if _, err = os.Stat(dirPath); err == nil {
		return nil, config.ErrProjectExists
	}

	return proj, nil
}

// Load the project configuration from the current working directory.
func (a *App) LoadCurrentProject(directory string) (err error) {

	logger.Debugf("Loading project config: %s", config.PROJECT_CONFIG_NAME)

	if directory == "" {
		if directory, err = os.Getwd(); err != nil {
			return err
		}
	}

	directory = filepath.ToSlash(directory)
	directory = filepath.FromSlash(directory)

	proj, err := config.LoadYaml[config.Project](
		filepath.Join(directory, config.PROJECT_CONFIG_NAME),
	)

	if err != nil && os.IsNotExist(err) {
		return config.ErrProjectMissing
	} else if err != nil {
		return errors.Wrapf(err, "failed to load current project's config %s", directory)
	}

	if err = proj.Validate(); err != nil {
		return err
	}

	a.ProjectConfig = proj

	logger.Debugf("Loaded project config %s", a.ProjectConfig.Name)

	for _, hook := range goldcrest.Get[ProjectHook](HookProjectLoaded) {
		if err = hook(a, proj); err != nil {
			return err
		}
	}

	return nil
}

// Write an example configuration for the user.
func (a *App) WriteExampleProjectConfig(directory string) (err error) {
	var example = config.ExampleProjectConfig()

	if directory == "" {
		if directory, err = os.Getwd(); err != nil {
			return err
		}
	}

	logger.Debugf("Writing example project config to %s", directory)

	for _, hook := range goldcrest.Get[ProjectHook](HookProjectExample) {
		if err = hook(a, example); err != nil {
			return err
		}
	}

	return config.WriteYaml(example, filepath.Join(directory, config.PROJECT_CONFIG_NAME))
}

func (a *App) WriteProject(proj *config.Project, directory string, raw bool) error {

	var (
		err error
	)

	// The directory to copy the project files to.
	if directory == "" {
		directory, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	logger.Debugf("Checking if project is locked in '%s'", directory)

	if err = config.IsLocked(directory); err != nil {
		return err
	}

	// Setup context for project templates.
	// Also setup the directory paths.
	var (
		context    = maps.Clone(proj.Context)
		projectDir = filepath.Join(directory, proj.Name)
	)

	if !filepath.IsAbs(projectDir) {
		projectDir, err = filepath.Abs(projectDir)
		if err != nil {
			return errors.Wrapf(err, "failed to get absolute path for %s", projectDir)
		}
	}

	projectDir = strings.ReplaceAll(projectDir, "\\", "/")

	// Update the context - also found in config.go.*ProjectCommand.Command.
	context["projectName"] = proj.Name
	context["projectPath"] = projectDir

	// Run commands before copying the project files.
	if err = proj.BeforeCopy.Execute(context); err != nil {
		return errors.Wrap(err, "failed to execute before copy steps")
	}

	for _, hook := range goldcrest.Get[ProjectWithDirHook](HookProjectBeforeWrite) {
		if err = hook(a, proj, projectDir); err != nil {
			return err
		}
	}

	// Create the project directory.
	// This is located in <cwd>/<projectName>.
	if err = os.MkdirAll(projectDir, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create project directory %s", projectDir)
	}

	logger.Infof("Copying project files to %s", projectDir)

	// Loop over all files in the project.
	// This gets recursively called by subdirectories.
	_, err = proj.Root.Traverse(func(fl quickfs.FileLike) (cancel bool, err error) {
		var p = fl.GetPath()
		var b = new(bytes.Buffer)

		err = a.executeProjectTemplate(
			proj, b, p,
		)
		if err != nil {
			return true, errors.Wrapf(err, "failed to execute template for filename %s", p)
		}

		var path = filepath.Join(projectDir, b.String())

		if proj.IsExcluded(fl) {
			logger.Debugf("Excluded %s", fl.GetPath())
			return false, nil
		}

		switch f := fl.(type) {
		case *quickfs.FSFile:
			// Copy the file content to the new file.
			// Replace any template variables.
			var dir = filepath.Dir(path)
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return true, errors.Wrapf(err, "failed to create directory %s", dir)
			}

			osFile, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, os.ModePerm)
			if err != nil && os.IsExist(err) {
				logger.Infof("Skipping, file %s already exists", path)
				return false, nil
			} else if err != nil {
				return true, errors.Wrapf(err, "failed to create file %s", path)
			}
			defer osFile.Close()

			if err = a.CopyFileContent(proj, osFile, f, raw); err != nil {
				return true, errors.Wrapf(err, "failed to copy file content to %s", path)
			}

		case *quickfs.FSDirectory:
			// Create a new subdirectory inside of the project directory.
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return true, errors.Wrapf(err, "failed to create directory %s", path)
			}
		}

		rel, err := filepath.Rel(projectDir, path)

		logger.Debugf("Copied %s to %s",
			fl.GetPath(),
			rel,
		)

		return false, nil
	})

	if err != nil {
		logger.Errorf("Failed to copy project files to %s", projectDir)
		return err
	}

	// Write the project configuration to the project directory.
	var configPath = filepath.Join(projectDir, config.PROJECT_CONFIG_NAME)
	if err = config.WriteYaml(proj, configPath); err != nil {
		return errors.Wrapf(err, "failed to write project config to %s", configPath)
	}

	// Run commands after copying the project files.
	err = proj.AfterCopy.Execute(context)
	if err != nil {
		return errors.Wrap(err, "failed to execute after copy steps")
	}

	for _, hook := range goldcrest.Get[ProjectWithDirHook](HookProjectAfterWrite) {
		if err = hook(a, proj, projectDir); err != nil {
			return err
		}
	}

	logger.Infof("Finished copying project files to %s", projectDir)

	return nil
}

func (a *App) WriteProjectConfig(proj *config.Project) error {
	var (
		err     error
		file    *os.File
		dirPath = GetProjectDirectoryPath(proj.Name, true)
	)

	var path = filepath.Join(dirPath, config.PROJECT_CONFIG_NAME)
	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}

	logger.Infof("Writing project config to %s", path)

	for _, hook := range goldcrest.Get[ProjectHook](HookProjectBeforeSave) {
		if err = hook(a, proj); err != nil {
			return err
		}
	}

	err = config.WriteYaml(proj, path)
	if err != nil {
		return err
	}

	var zipPath = filepath.Join(dirPath, config.PROJECT_ZIP_NAME)

	if file, err = os.Create(zipPath); err != nil {
		return err
	}
	defer file.Close()

	var zf = zip.NewWriter(file)
	defer zf.Close()

	logger.Infof("Writing project files to %s", zipPath)

	_, err = proj.Root.Traverse(func(fl quickfs.FileLike) (cancel bool, err error) {
		var p = fl.GetPath()

		switch f := fl.(type) {
		case *quickfs.FSFile:
			var w, err = zf.Create(p)
			if err != nil {
				return true, err
			}
			if _, err = io.Copy(w, f); err != nil {
				return true, err
			}

		case *quickfs.FSDirectory:
			if !strings.HasSuffix(p, "/") {
				p += "/"
			}

			if _, err = zf.Create(p); err != nil {
				return true, err
			}
		}

		logger.Debugf("Wrote %s to %s", fl.GetPath(), p)

		return false, nil
	})

	if err != nil {
		logger.Errorf("Failed to write project to %s: %s", zipPath, err)
		return err
	}

	logger.Infof("Finished writing project to %s", zipPath)

	for _, hook := range goldcrest.Get[ProjectHook](HookProjectAfterSave) {
		if err = hook(a, proj); err != nil {
			return err
		}
	}

	return err
}

func (a *App) ReadProjectConfig(name string) (proj *config.Project, closeFiles func(), err error) {

	if name == "" || name == "." || strings.ContainsAny(name, `/\`) {
		return nil, nil, config.ErrProjectName
	}

	var (
		file *os.File
		// dirPath    = getProjectFilePath(name, false)
		// absDirPath = getProjectFilePath(name, true)
		absDirPath = GetProjectDirectoryPath(name, true)
	)
	proj, err = config.LoadYaml[config.Project](
		path.Join(absDirPath, config.PROJECT_CONFIG_NAME),
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to load YAML for project config %s", name)
	}

	for _, hook := range goldcrest.Get[ProjectWithDirHook](HookProjectBeforeLoad) {
		if err = hook(a, proj, absDirPath); err != nil {
			return nil, closeFiles, err
		}
	}

	proj.Root = quickfs.NewFSDirectory(
		proj.Name,
		".",
		nil,
	)
	proj.Root.IsExcluded = proj.IsExcluded

	file, err = os.Open(
		path.Join(absDirPath, config.PROJECT_ZIP_NAME),
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to open zip file for project %s", name)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get file info for %s", file.Name())
	}

	zf, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to read zip file %s", file.Name())
	}

	var zipFiles = make([]io.ReadCloser, 0, len(zf.File))
	closeFiles = func() {
		for _, f := range zipFiles {
			f.Close()
		}
	}

	for _, f := range zf.File {
		var (
			fInfo = f.FileInfo()
		)

		if f.Name == "." || f.Name == "./" {
			continue
		}

		if fInfo.IsDir() {
			proj.Root.AddDirectory(f.Name)

		} else {

			zipF, err := f.Open()
			if err != nil {
				return nil, closeFiles, err
			}

			zipFiles = append(zipFiles, zipF)

			var f = proj.Root.AddFile(f.Name, zipF)
			f.Size = fInfo.Size()
		}
	}

	for _, hook := range goldcrest.Get[ProjectHook](HookQuickGoLoaded) {
		if err = hook(a, proj); err != nil {
			return nil, closeFiles, err
		}
	}

	return proj, closeFiles, nil
}

func (a *App) CopyFileContent(proj *config.Project, file *os.File, f *quickfs.FSFile, raw bool) error {

	if raw {
		_, err := io.Copy(file, f)
		return err
	}

	var b = new(bytes.Buffer)
	if _, err := io.Copy(b, f); err != nil {
		return err
	}

	f.IsText = quickfs.IsText(b.Bytes()) && !strings.HasSuffix(
		f.Name, config.PROJECT_CONFIG_NAME,
	)

	if !f.IsText {
		_, err := io.Copy(file, b)
		return err
	}

	content, err := io.ReadAll(b)
	if err != nil {
		return err
	}

	return a.executeProjectTemplate(
		proj, file, string(content),
	)
}

func (a *App) ListProjectObjects() ([]*config.Project, error) {
	var (
		projects = make([]*config.Project, 0)
		dirPath  = GetProjectDirectoryPath("", true)
	)

	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", dirPath)
	}

	for _, d := range dir {
		if !d.IsDir() {
			continue
		}

		var path = filepath.Join(dirPath, d.Name())
		var configName = filepath.Join(path, config.PROJECT_CONFIG_NAME)
		proj, err := config.LoadYaml[config.Project](configName)
		if err != nil && !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "failed to load project config %s", configName)
		}

		projects = append(projects, proj)
	}

	for _, hook := range goldcrest.Get[AppListProjectsHook](HookQuickGoListProjects) {
		var p, err = hook(a, projects)
		if err != nil {
			return nil, err
		}

		projects = p

	}

	return projects, nil
}

func (a *App) ListJSFiles() ([]string, error) {
	var (
		dirPath = GetQuickGoPath(config.COMMANDS_DIR)
		files   = make([]string, 0)
	)

	dir, err := os.ReadDir(dirPath)
	if err != nil && os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %s", dirPath)
	}

	for _, d := range dir {
		if d.IsDir() {
			continue
		}

		if strings.HasSuffix(d.Name(), ".js") {
			files = append(files, d.Name()[0:len(d.Name())-3])
		}
	}

	return files, nil
}

func (a *App) SaveJS(path string) error {
	var (
		dirPath   = GetQuickGoPath(config.COMMANDS_DIR)
		filename  = filepath.Join(dirPath, filepath.Base(path))
		scriptSrc *os.File
		file      *os.File
		err       error
	)

	path = filepath.FromSlash(path)

	logger.Infof("Copying file %s to %s", path, filename)

	if s, err := os.Stat(path); err != nil || s.IsDir() {
		return errors.Wrapf(err, "file %s does not exist or is not a valid file", path)
	}

	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create directory %s", dirPath)
	}

	file, err = os.Create(
		filename,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", path)
	}
	defer file.Close()

	scriptSrc, err = os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", path)
	}

	_, err = io.Copy(file, scriptSrc)
	return errors.Wrapf(err, "failed to copy file %s to %s", path, file.Name())
}

func (a *App) ExecJS(scriptName string, args map[string]any) (err error) {
	var (
		scriptPath = GetQuickGoPath(
			config.COMMANDS_DIR,
			fmt.Sprintf("%s.js", scriptName),
		)

		script []byte
		proj   *config.Project
		cmd    *js.Command
	)

	if proj = a.ProjectConfig; proj != nil {
		args["project"] = proj
	}

	script, err = os.ReadFile(scriptPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read script %s", script)
	}

	cmd = js.NewScript(
		"main",
		js.WithGlobal("readFile", func(path string) []byte {
			var data, err = os.ReadFile(path)
			if err != nil {
				logger.Error(err)
			}
			return data
		}),
		js.WithGlobals(args),
	)

	cmd.AddFunc(func(vm *goja.Runtime) error {
		vm.Set("app", a)
		vm.Set("console", &js.JSConsole{
			Debug: logger.Debug,
			Log:   logger.Info,
			Info:  logger.Info,
			Warn:  logger.Warn,
			Error: logger.Error,
		})
		vm.Set("base64Encode", func(data goja.Value) string {
			if vm.InstanceOf(data, vm.Get("Uint8Array").ToObject(vm)) {
				var arr = data.Export().([]byte)
				return base64.StdEncoding.EncodeToString(arr)
			}

			if data, ok := data.Export().(goja.ArrayBuffer); ok {
				return base64.StdEncoding.EncodeToString(data.Bytes())
			}

			var s = data.String()
			return base64.StdEncoding.EncodeToString([]byte(s))
		})
		vm.Set("base64Decode", func(data string) goja.Value {
			var arr, err = base64.StdEncoding.DecodeString(data)
			if err != nil {
				logger.Error(err)
				return goja.Undefined()
			}

			var arrBuff = vm.NewArrayBuffer(arr)
			return vm.ToValue(arrBuff)
		})
		vm.Set("writeFile", func(path string, data goja.Value) error {
			var b []byte
			if vm.InstanceOf(data, vm.Get("Uint8Array").ToObject(vm)) {
				b = data.Export().([]byte)
			} else if d, ok := data.Export().(goja.ArrayBuffer); ok {
				b = d.Bytes()
			} else {
				b = []byte(data.String())
			}

			return os.WriteFile(path, b, os.ModePerm)
		})
		vm.Set("readFile", func(path string) goja.Value {
			var data []byte
			var err error
			if data, err = os.ReadFile(path); err != nil {
				logger.Error(err)
				return goja.Undefined()
			}

			var arr = vm.NewArrayBuffer(data)
			return vm.ToValue(arr)
		})
		vm.Set("readTextFile", func(path string) string {
			var data, err = os.ReadFile(path)
			if err != nil {
				logger.Error(err)
				return ""
			}
			return string(data)
		})

		return nil
	})

	return cmd.Run(string(script))
}

func (a *App) ListProjects() ([]string, error) {
	var p, err = a.ListProjectObjects()
	if err != nil {
		return nil, err
	}
	var names = make([]string, 0, len(p))
	for _, proj := range p {
		names = append(names, proj.Name)
	}
	return names, nil
}

func (a *App) WriteFile(data []byte, path string) error {
	return os.WriteFile(GetProjectDirectoryPath(path, true), data, os.ModePerm)
}

func (a *App) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(GetProjectDirectoryPath(path, true))
}

func (a *App) executeProjectTemplate(proj *config.Project, w io.Writer, content string) error {

	var tpl = template.New("file")
	tpl.Delims(
		proj.DelimLeft,
		proj.DelimRight,
	)

	if _, err := tpl.Parse(content); err != nil {
		return errors.Wrapf(
			err,
			"failed to parse template %s",
			content,
		)
	}

	return errors.Wrapf(
		tpl.Execute(w, proj),
		"failed to execute template %s",
		content,
	)
}

func GetQuickGoPath(p ...string) string {
	if len(p) == 0 {
		p = []string{quickGoConfigDir, config.QUICKGO_DIR}
	} else {
		p = append([]string{quickGoConfigDir, config.QUICKGO_DIR}, p...)
	}
	return filepath.ToSlash(filepath.Join(p...))
}

func GetProjectDirectoryPath(name string, absolute bool) string {
	var p string
	if absolute {
		p = filepath.ToSlash(GetQuickGoPath(
			config.PROJECTS_DIR,
			name,
		))
	} else {
		p = filepath.ToSlash(filepath.Join(
			config.PROJECTS_DIR,
			name,
		))
	}

	return strings.ReplaceAll(p, "\\", "/")
}
