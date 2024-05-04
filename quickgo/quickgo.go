package quickgo

import (
	"archive/zip"
	"bytes"
	"fmt"
	html_template "html/template"
	"io"
	"io/fs"
	"maps"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Nigel2392/goldcrest"
	"github.com/Nigel2392/quickgo/v2/quickgo/config"
	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
	"github.com/Nigel2392/quickgo/v2/quickgo/quickfs"
	"github.com/pkg/errors"
)

var (
	cliApplication *App
)

type (
	App struct {
		Config        *config.QuickGo `yaml:"config"`        // The configuration for QuickGo.
		ProjectConfig *config.Project `yaml:"projectConfig"` // The configuration for the project.
		Patterns      []string        `yaml:"patterns"`      // The patterns for the templates.
		AppFS         fs.FS           `yaml:"-"`             // The file system for the app, resides in the executable directory.
		ProjectFS     fs.FS           `yaml:"-"`             // The file system for the project, resides in the project (working) directory.
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
	var quickGoDir = filepath.Join(executableDir, config.QUICKGO_DIR)
	var projectDir = filepath.Join(executableDir, config.QUICKGO_DIR, "projects")
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
			"_templates/readme.tmpl",
			"_templates/index.tmpl",
		},
	}

	// Setup the global application to prevent multiple instances.
	cliApplication = app

	// Load the QuickGo configuration.
	var configPath = filepath.Join(
		executableDir, config.QUICKGO_DIR, config.QUICKGO_CONFIG_NAME,
	)

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

// Load the project configuration from the current working directory.
func (a *App) LoadProjectConfig(directory string) (err error) {

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
		return err
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

	logger.Debugf("Checking if project is locked in '%s'", cwd)

	if err = config.IsLocked(cwd); err != nil {
		return err
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
	_, err = proj.Root.ForEach(func(fl quickfs.FileLike) (cancel bool, err error) {
		var p = fl.GetPath()
		var b = new(bytes.Buffer)

		err = a.executeTemplate(
			proj, b, p,
		)
		if err != nil {
			return true, errors.Wrapf(err, "failed to execute template for filename %s", p)
		}

		var path = filepath.Join(projectDir, b.String())

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
		dirPath = getProjectFilePath(proj.Name, true)
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

	_, err = proj.Root.ForEach(func(fl quickfs.FileLike) (cancel bool, err error) {
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
		absDirPath = getProjectFilePath(name, true)
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

			proj.Root.AddFile(f.Name, zipF)
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

	return a.executeTemplate(
		proj, file, string(content),
	)
}

func (a *App) ListProjects() ([]string, error) {
	var (
		projects = make([]string, 0)
		dirPath  = getProjectFilePath("", true)
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
		if _, err = os.Stat(configName); err != nil {
			continue
		}

		projects = append(projects, d.Name())
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

func (a *App) WriteFile(data []byte, path string) error {
	path = filepath.Join(executableDir, config.QUICKGO_DIR, path)
	return os.WriteFile(path, data, os.ModePerm)
}

func (a *App) ReadFile(path string) ([]byte, error) {
	path = filepath.Join(executableDir, config.QUICKGO_DIR, path)
	return os.ReadFile(path)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	go func() {
		if err := recover(); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()

	var pathParts = strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var primary = strings.ToLower(pathParts[0])

	for _, hook := range goldcrest.Get[AppServeHook](HookQuickGoServer) {
		if served, err := hook(a, w, r); err != nil {
			logger.Errorf("Failed to serve: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		} else if served {
			logger.Debugf("'%s' was served and hijacked by a hook", r.URL.Path)
		}
	}

	switch {
	case primary == "projects":

		a.serveProjects(w, r, pathParts[1:])

	case primary == "favicon.ico" && len(pathParts) == 1:

		// Write out the favicon file.
		// No need to log the happy path here, quite boring.
		var f, err = staticFS.Open("quickgo.png")
		if err != nil {
			logger.Errorf("Failed to open 'quickgo.png': %v", err)
			http.Error(w, "Failed to open 'quickgo.png'", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		w.Header().Set("Content-Type", "image/x-icon")
		if _, err = io.Copy(w, f); err != nil {
			logger.Errorf("Failed to read 'quickgo.png': %v", err)
			http.Error(w, "Failed to read 'quickgo.png'", http.StatusInternalServerError)
		}

	case primary == "static":

		// Serve static files.
		logger.Debugf("Serving static file %s", r.URL.Path)
		// Serve from embedded file system.
		var handler = http.FileServer(http.FS(staticFS))
		handler = http.StripPrefix("/static/", handler)
		handler.ServeHTTP(w, r)

	default:

		logger.Debugf("Invalid request path '%s'", r.URL.Path)
		http.Error(w, "Invalid path", http.StatusBadRequest)
	}
}

type ProjectTemplateContext struct {
	Project *config.Project
	File    *quickfs.FSFile
	Dir     *quickfs.FSDirectory
	Content string
}

func (a *App) serveProjects(w http.ResponseWriter, r *http.Request, pathParts []string) {
	var (
		proj     *config.Project
		parent   *quickfs.FSDirectory
		fileLike quickfs.FileLike
		err      error
	)

	if len(pathParts) == 0 {
		logger.Debugf("Invalid request path '%s'", r.URL.Path)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	proj, closeFiles, err := a.ReadProjectConfig(pathParts[0])
	if err != nil {
		logger.Errorf("Failed to read project '%s': %v", pathParts[0], err)
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	defer closeFiles()

	parent, fileLike, err = proj.Root.FindWithParent(pathParts[1:])
	if err != nil {
		logger.Errorf("Failed to find file '%s': %v", path.Join(pathParts...), err)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	if fileLike.IsDir() {
		// logger.Debugf("Invalid request path '%s', cannot open directory.", r.URL.Path)
		// http.Error(w, "Invalid path", http.StatusBadRequest)

		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Cache-Control", "no-cache")
		var dir = fileLike.(*quickfs.FSDirectory)

		if dir.Root() != nil {
			fmt.Fprintf(w,
				"<a href='/%s'>../%s</a><br>",
				path.Join(
					"projects",
					proj.Name,
					parent.GetPath(),
				),
				parent.GetName(),
			)
		}

		quickfs.PrintRootFn(w, dir, "&nbsp;", func(indent int, fl quickfs.FileLike) string {
			var (
				p   = fl.GetPath()
				url = filepath.ToSlash(
					path.Join("projects", proj.Name, p),
				)
				name = fl.GetName()
			)

			if !strings.HasPrefix(url, "/") {
				url = "/" + url
			}

			return fmt.Sprintf("<a href='%s' data-indent='%d'>%s</a><br>", url, indent, name)
		})

		return
	}

	var (
		tpl  *html_template.Template
		file = fileLike.(*quickfs.FSFile)
		b    = new(bytes.Buffer)
	)

	if _, err = io.Copy(b, file); err != nil {
		logger.Errorf("Failed to read file '%s': %v", file.GetPath(), err)
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	var content = b.String()
	file.IsText = quickfs.IsText(content)
	if !file.IsText {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename="+file.GetPath())
		if _, err = io.Copy(w, b); err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
		}
		return
	}

	if tpl, err = html_template.ParseFS(embedFS, a.Patterns...); err != nil {
		logger.Errorf("Failed to parse template: %v", err)
		http.Error(w, "Failed to parse file", http.StatusInternalServerError)
		return
	}

	var context = ProjectTemplateContext{
		Project: proj,
		File:    file,
		Dir:     parent,
		Content: content,
	}

	if err = tpl.ExecuteTemplate(w, "base", context); err != nil {
		logger.Errorf("Failed to execute template: %v", err)
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}

func (a *App) executeTemplate(proj *config.Project, w io.Writer, content string) error {

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

func getProjectFilePath(name string, absolute bool) string {
	var p string
	if absolute {
		p = path.Join(
			executableDir,
			config.QUICKGO_DIR,
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
