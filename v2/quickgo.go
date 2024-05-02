package quickgo

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Nigel2392/quickgo/v2/config"
	"github.com/Nigel2392/quickgo/v2/logger"
	"github.com/Nigel2392/quickgo/v2/quickfs"
	"github.com/pkg/errors"
)

const (
	QUICKGO_DIR         = ".quickgo"     // The directory for QuickGo files, resides in the executable directory.
	QUICKGO_CONFIG_NAME = "quickgo.yaml" // Config file for QuickGo, resides in the executable directory.
	PROJECT_CONFIG_NAME = "quickgo.yaml" // Config file for the project, resides in the project (working) directory.
	PROJECT_ZIP_NAME    = "project.zip"  // The name of the project zip file.

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

	// Check for the application config directory.
	var quickGoDir = filepath.Join(executableDir, QUICKGO_DIR)
	var projectDir = filepath.Join(executableDir, QUICKGO_DIR, "projects")
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
	}

	// Setup the global application to prevent multiple instances.
	cliApplication = app

	// Load the QuickGo configuration.
	var configPath = filepath.Join(
		executableDir, QUICKGO_DIR, QUICKGO_CONFIG_NAME,
	)

	cfg, err := config.LoadYamlFS[config.QuickGo](
		app.AppFS,
		QUICKGO_CONFIG_NAME,
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

	return app, nil
}

// Load the project configuration from the current working directory.
func (a *App) LoadProjectConfig() error {

	logger.Debugf("Loading project config: %s", PROJECT_CONFIG_NAME)

	var proj, err = config.LoadYamlFS[config.Project](
		a.ProjectFS,
		PROJECT_CONFIG_NAME,
	)
	if err != nil && os.IsNotExist(err) {
		logger.Error("Project config not found")
		return ErrProjectMissing
	} else if err != nil {
		return err
	}
	a.ProjectConfig = proj

	logger.Debugf("Loaded project config %s", a.ProjectConfig.Name)

	return nil
}

// Write an example configuration for the user.
func (a *App) WriteExampleProjectConfig() error {
	var example = config.ExampleProjectConfig()
	return config.WriteYaml(example, PROJECT_CONFIG_NAME)
}

func (a *App) WriteProject(proj *config.Project, directory string) error {

	var (
		cwd string
		err error
	)

	// The directory to copy the project files to.
	if directory == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return err
		}
	} else {
		cwd = directory
	}

	// Setup context for project templates.
	// Also setup the directory paths.
	var (
		context    = maps.Clone(proj.Context)
		projectDir = filepath.Join(cwd, proj.Name)
	)

	if !filepath.IsAbs(projectDir) {
		projectDir, err = filepath.Abs(projectDir)
		if err != nil {
			return errors.Wrapf(err, "failed to get absolute path for %s", projectDir)
		}
	}

	projectDir = strings.ReplaceAll(projectDir, "\\", "/")

	context["projectName"] = proj.Name
	context["projectPath"] = projectDir

	// Run commands before copying the project files.
	if err = a.ProjectConfig.BeforeCopy.Execute(context); err != nil {
		return errors.Wrap(err, "failed to execute before copy steps")
	}

	// Create the project directory.
	// This is located in <cwd>/<projectName>.
	if err = os.MkdirAll(projectDir, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create project directory %s", projectDir)
	}

	logger.Infof("Copying project files to %s", projectDir)

	// Loop over all files in the project.
	// This gets recursively called by subdirectories.
	_, err = proj.Root.ForEach(func(fl quickfs.FileLike) (cancel bool, err error) {
		var p = fl.GetPath()
		var b = new(bytes.Buffer)

		err = a.executeTemplate(
			"file", b, p,
		)
		if err != nil {
			return true, errors.Wrapf(err, "failed to execute template for filename %s", p)
		}

		var path = path.Join(projectDir, b.String())

		logger.Debugf("Copying %s to %s (isDir=%v)",
			pathForLog(fl.GetPath()),
			pathForLog(path),
			fl.IsDir(),
		)

		switch f := fl.(type) {
		case *quickfs.FSFile:
			// Copy the file content to the new file.
			// Replace any template variables.
			var dir = filepath.Dir(path)
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return true, errors.Wrapf(err, "failed to create directory %s", dir)
			}

			osFile, err := os.Create(path)
			if err != nil {
				return true, errors.Wrapf(err, "failed to create file %s", path)
			}
			defer osFile.Close()

			if err = a.CopyFileContent(osFile, f); err != nil {
				return true, errors.Wrapf(err, "failed to copy file content to %s", path)
			}

		case *quickfs.FSDirectory:
			// Create a new subdirectory inside of the project directory.
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return true, errors.Wrapf(err, "failed to create directory %s", path)
			}
		}

		logger.Debugf("Copied %s to %s",
			pathForLog(fl.GetPath()),
			pathForLog(path),
		)

		return false, nil
	})

	if err != nil {
		logger.Errorf("Failed to copy project files to %s", projectDir)
		return err
	}

	// Run commands after copying the project files.
	err = a.ProjectConfig.AfterCopy.Execute(context)
	if err != nil {
		return errors.Wrap(err, "failed to execute after copy steps")
	}

	logger.Infof("Finished copying project files to %s", projectDir)

	return nil

}

func (a *App) WriteProjectConfig(proj *config.Project) error {
	var (
		err     error
		file    *os.File
		dirPath = getProjectFilePath(proj.Name, true)
	)

	var path = filepath.Join(dirPath, PROJECT_CONFIG_NAME)
	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}

	logger.Infof("Writing project config to %s", path)

	err = config.WriteYaml(proj, path)
	if err != nil {
		return err
	}

	var zipPath = filepath.Join(dirPath, PROJECT_ZIP_NAME)

	if file, err = os.Create(zipPath); err != nil {
		return err
	}
	defer file.Close()

	var zf = zip.NewWriter(file)
	defer zf.Close()

	logger.Infof("Writing project files to %s", zipPath)

	_, err = a.ProjectConfig.Root.ForEach(func(fl quickfs.FileLike) (cancel bool, err error) {
		var p = fl.GetPath()

		logger.Debugf("Writing %s to %s (isDir=%v)", fl.GetPath(), p, fl.IsDir())

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

	return err
}

func (a *App) ReadProjectConfig(name string) (proj *config.Project, closeFiles func(), err error) {
	var (
		file       *os.File
		dirPath    = getProjectFilePath(name, false)
		absDirPath = getProjectFilePath(name, true)
	)
	proj, err = config.LoadYaml[config.Project](
		path.Join(absDirPath, PROJECT_CONFIG_NAME),
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to load YAML for project config %s", name)
	}

	proj.Root = quickfs.NewFSDirectory(
		proj.Name,
		dirPath,
		nil,
	)
	proj.Root.IsExcluded = proj.IsExcluded

	file, err = os.Open(
		path.Join(absDirPath, PROJECT_ZIP_NAME),
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

		if fInfo.IsDir() {
			proj.Root.AddDirectory(f.Name)

		} else {

			zipF, err := f.Open()
			if err != nil {
				return nil, closeFiles, err
			}

			zipFiles = append(zipFiles, zipF)

			proj.Root.AddFile(f.Name, zipF)
		}
	}

	return proj, closeFiles, nil
}

func (a *App) CopyFileContent(file *os.File, f *quickfs.FSFile) error {
	var b = new(bytes.Buffer)
	if _, err := io.Copy(b, f); err != nil {
		return err
	}

	f.IsText = quickfs.IsText(b.Bytes()) && !strings.HasSuffix(
		f.Name, PROJECT_CONFIG_NAME,
	)

	if !f.IsText {
		_, err := io.Copy(file, b)
		return err
	}

	content, err := io.ReadAll(b)
	if err != nil {
		return err
	}

	return a.executeTemplate(
		"file", file, string(content),
	)
}

func (a *App) executeTemplate(name string, w io.Writer, content string) error {

	var tpl = template.New(name)
	tpl.Delims(
		a.ProjectConfig.DelimLeft,
		a.ProjectConfig.DelimRight,
	)

	if _, err := tpl.Parse(content); err != nil {
		return errors.Wrapf(
			err,
			"failed to parse template %s",
			content,
		)
	}

	return errors.Wrapf(
		tpl.Execute(w, a.ProjectConfig),
		"failed to execute template %s",
		content,
	)
}

func (a *App) WriteFile(data []byte, path string) error {
	path = filepath.Join(executableDir, QUICKGO_DIR, path)
	return os.WriteFile(path, data, os.ModePerm)
}

func (a *App) ReadFile(path string) ([]byte, error) {
	path = filepath.Join(executableDir, QUICKGO_DIR, path)
	return os.ReadFile(path)
}

func pathForLog(p string) string {
	var parts = filepath.SplitList(p)
	if len(parts) < 3 {
		return p
	}
	if len(parts) == 3 {
		return fmt.Sprintf("%s/.../%s", parts[0], parts[2])
	}
	return fmt.Sprintf("%s/.../%s", parts[0], parts[len(parts)-1])
}

func getProjectFilePath(name string, absolute bool) string {
	var p string
	if absolute {
		p = path.Join(
			executableDir,
			QUICKGO_DIR,
			"projects",
			name,
		)
	} else {
		p = path.Join(
			"projects",
			name,
		)
	}

	return strings.ReplaceAll(p, "\\", "/")
}
