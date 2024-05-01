package quickgo

import (
	"encoding/gob"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Nigel2392/quickgo/v2/config"
	"github.com/Nigel2392/quickgo/v2/quickfs"
	"github.com/pkg/errors"
)

const (
	QUICKGO_DIR         = ".quickgo"     // The directory for QuickGo files, resides in the executable directory.
	QUICKGO_CONFIG_NAME = "quickgo.yaml" // Config file for QuickGo, resides in the executable directory.
	PROJECT_CONFIG_NAME = "quickgo.yaml" // Config file for the project, resides in the project (working) directory.

	// Error messages.
	ErrProjectMissing = ErrorStr("project config not found")
)

var cliApplication *App

type (
	App struct {
		Config        *config.QuickGo `yaml:"config"`        // The configuration for QuickGo.
		ProjectConfig *config.Project `yaml:"projectConfig"` // The configuration for the project.
		AppFS         fs.FS           `yaml:"-"`             // The file system for the app, resides in the executable directory.
		ProjectFS     fs.FS           `yaml:"-"`             // The file system for the project, resides in the project (working) directory.
	}

	ErrorStr string
)

func (e ErrorStr) Error() string {
	return string(e)
}

func LoadApp() (*App, error) {
	if cliApplication != nil {
		return cliApplication, nil
	}

	var wd, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	var quickGoDir = filepath.Join(executableDir, QUICKGO_DIR)
	var projectDir = filepath.Join(executableDir, QUICKGO_DIR, "projects")
	_, err = os.Stat(projectDir)

	if err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(projectDir, os.ModePerm); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	var app = &App{
		AppFS:     os.DirFS(quickGoDir),
		ProjectFS: os.DirFS(wd),
	}

	cliApplication = app

	var configPath = filepath.Join(
		executableDir, QUICKGO_DIR, QUICKGO_CONFIG_NAME,
	)

	cfg, err := config.LoadYamlFS[config.QuickGo](
		app.AppFS,
		QUICKGO_CONFIG_NAME,
	)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "failed to load config file %s", configPath)
		}

		cfg = &config.QuickGo{
			Host: "localhost",
			Port: "8080",
		}

		if err = config.WriteYaml(cfg, configPath); err != nil {
			return nil, errors.Wrapf(err, "failed to write config file %s", configPath)
		}
	}

	app.Config = cfg

	return app, nil
}

func (a *App) LoadProjectConfig(projectDirectory string) error {
	proj, err := config.LoadYamlFS[config.Project](
		a.ProjectFS,
		PROJECT_CONFIG_NAME,
	)
	if err != nil && os.IsNotExist(err) {
		return ErrProjectMissing
	} else if err != nil {
		return err
	}
	a.ProjectConfig = proj

	a.ProjectConfig.Root = quickfs.NewFSDirectory(
		"$$PROJECT_NAME$$",
		projectDirectory,
		nil,
	)

	a.ProjectConfig.Root.IsExcluded = a.ProjectConfig.IsExcluded

	return a.ProjectConfig.Root.Load()
}

func (a *App) WriteExampleProjectConfig() error {
	var example = config.ExampleProjectConfig()
	return config.WriteYaml(example, PROJECT_CONFIG_NAME)
}

func getProjectFilePath(name string) string {
	return filepath.Join(
		executableDir,
		QUICKGO_DIR,
		"projects",
		fmt.Sprintf(
			"%s.quickgo",
			strings.ToLower(name),
		),
	)
}

func (a *App) WriteProjectConfig(proj *config.Project) error {
	var (
		err  error
		file *os.File
		path = getProjectFilePath(proj.Name)
	)

	if file, err = os.Create(path); err != nil {
		return err
	}
	defer file.Close()

	return gob.NewEncoder(file).Encode(a.ProjectConfig)
}

func (a *App) ReadProjectConfig(name string) (*config.Project, error) {
	var (
		err  error
		file *os.File
		path = getProjectFilePath(name)
		proj = new(config.Project)
	)

	if file, err = os.Open(path); err != nil {
		return nil, err
	}

	defer file.Close()

	if err = gob.NewDecoder(file).Decode(proj); err != nil {
		return nil, err
	}

	return proj, nil
}

func (a *App) WriteProject(proj *config.Project, directory string) error {

	var (
		cwd string
		err error
	)

	if directory == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return err
		}
	} else {
		cwd = directory
	}

	var (
		context    = maps.Clone(proj.Context)
		projectDir = filepath.Join(cwd, proj.Name)
	)

	context["projectName"] = proj.Name
	context["projectDir"] = projectDir

	if err = a.ProjectConfig.BeforeCopy.Execute(context); err != nil {
		return err
	}

	if err = os.MkdirAll(projectDir, os.ModePerm); err != nil {
		return err
	}

	_, err = proj.Root.ForEach(func(fl quickfs.FileLike) (cancel bool, err error) {
		var p = fl.GetPath()
		if strings.Contains(p, "$$PROJECT_NAME$$") {
			p = strings.ReplaceAll(p, "$$PROJECT_NAME$$", proj.Name)
		}

		var path = filepath.Join(projectDir, p)
		switch f := fl.(type) {
		case *quickfs.FSFile:
			osFile, err := os.Create(path)
			if err != nil {
				return true, err
			}
			defer osFile.Close()

			if err = a.CopyFileContent(osFile, f); err != nil {
				return true, err
			}

		case *quickfs.FSDirectory:
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return true, err
			}
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return a.ProjectConfig.AfterCopy.Execute(context)
}

func (a *App) CopyFileContent(file *os.File, f *quickfs.FSFile) error {
	if !f.IsText {
		_, err := file.Write(f.Content)
		return err
	}

	var tpl = template.New("file")

	if _, err := tpl.Parse(string(f.Content)); err != nil {
		return err
	}

	return tpl.Execute(file, a.ProjectConfig)
}

func (a *App) WriteFile(data []byte, path string) error {
	path = filepath.Join(executableDir, QUICKGO_DIR, path)
	return os.WriteFile(path, data, os.ModePerm)
}

func (a *App) ReadFile(path string) ([]byte, error) {
	path = filepath.Join(executableDir, QUICKGO_DIR, path)
	return os.ReadFile(path)
}
