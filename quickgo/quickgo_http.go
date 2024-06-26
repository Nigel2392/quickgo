package quickgo

import (
	"archive/zip"
	"bytes"
	"fmt"
	html_template "html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Nigel2392/goldcrest"
	"github.com/Nigel2392/quickgo/v2/quickgo/config"
	"github.com/Nigel2392/quickgo/v2/quickgo/logger"
	"github.com/Nigel2392/quickgo/v2/quickgo/quickfs"
	"github.com/pkg/errors"
)

func copyToZipfile(z *zip.Writer, name string, absPath string) error {
	var f, err = os.Open(absPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open '%s'", absPath)
	}
	defer f.Close()

	w, err := z.Create(name)
	if err != nil {
		return errors.Wrapf(err, "failed to create '%s'", name)
	}

	_, err = io.Copy(w, f)
	return err

}

func (a *App) HttpHandler() http.Handler {
	var mux = http.NewServeMux()
	mux.Handle("/", &LogHandler{
		Handler: http.HandlerFunc(a.serveIndex),
		Where:   "index",
		Level:   logger.InfoLevel,
	})
	mux.Handle("/projects/", &LogHandler{
		Handler: http.StripPrefix(
			"/projects/",
			http.HandlerFunc(a.serveProjects),
		),
		Where: "projects",
		Level: logger.InfoLevel,
	})
	mux.Handle("/config/", &LogHandler{
		Handler: http.StripPrefix(
			"/config/",
			http.HandlerFunc(a.serveProjectConfig),
		),
		Where: "project config",
		Level: logger.InfoLevel,
	})
	mux.Handle("/static/", &LogHandler{
		Handler: http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))),
		Where:   "static files",
		Level:   logger.DebugLevel,
	})
	mux.Handle("/favicon.ico", &LogHandler{
		Handler: http.HandlerFunc(a.serveFavicon),
		Where:   "favicon",
		Level:   logger.DebugLevel,
	})
	return a.middleware(mux)
}

func (a *App) serveIndex(w http.ResponseWriter, r *http.Request) {

	var projects, err = a.ListProjectObjects()
	if err != nil {
		logger.Errorf("Failed to list projects: %v", err)
		http.Error(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	var ctx = &ProjectTemplateContext{
		ObjectList: projects,
	}

	if err = a.executeServeTemplate(w, "index.tmpl", ctx); err != nil {
		logger.Errorf("Failed to render index: %v", err)
		http.Error(w, "Failed to render index", http.StatusInternalServerError)
	}
}

func (a *App) serveProjects(w http.ResponseWriter, r *http.Request) {
	var (
		proj     *config.Project
		parent   *quickfs.FSDirectory
		fileLike quickfs.FileLike
		err      error
	)

	var pathParts = strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if r.URL.Query().Get("download") != "" && len(pathParts) == 1 {

		var (
			projName   = pathParts[0]
			projPath   = GetProjectDirectoryPath(projName, true)
			configFile = filepath.Join(projPath, config.PROJECT_CONFIG_NAME)
			projectZip = filepath.Join(projPath, config.PROJECT_ZIP_NAME)
		)

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", projName))

		var zf = zip.NewWriter(w)
		defer zf.Close()

		logger.Debugf("Adding '%s' to zip for export", configFile)
		if err = copyToZipfile(zf, config.PROJECT_CONFIG_NAME, configFile); err != nil {
			logger.Errorf("Failed to copy '%s' to zip: %v", configFile, err)
			http.Error(w, "Failed to copy to zip", http.StatusInternalServerError)
			return
		}

		logger.Debugf("Adding '%s' to zip for export", projectZip)
		if err = copyToZipfile(zf, config.PROJECT_ZIP_NAME, projectZip); err != nil {
			logger.Errorf("Failed to copy '%s' to zip: %v", projectZip, err)
			http.Error(w, "Failed to copy to zip", http.StatusInternalServerError)
			return
		}

		logger.Infof("%s Downloaded project '%s'", r.RemoteAddr, projName)
		return
	}

	logger.Debugf("Reading project '%s'", pathParts[0])
	proj, closeFiles, err := a.ReadProjectConfig(pathParts[0])
	if err != nil {
		logger.Errorf("Failed to read project '%s': %v", pathParts[0], err)
		http.Error(w, "Invalid project", http.StatusBadRequest)
		return
	}

	defer func() {
		logger.Debugf("Closing project '%s'", proj.Name)
		closeFiles()
	}()

	parent, fileLike, err = proj.Root.FindWithParent(pathParts[1:])
	if err != nil {
		logger.Errorf("Failed to find file '%s': %v", path.Join(pathParts...), err)
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	var context = ProjectTemplateContext{
		Project: proj,
		Parent:  parent,
	}

	if fileLike.IsDir() {
		var dir = fileLike.(*quickfs.FSDirectory)
		var FileObjects = make([]quickfs.FileLike, 0)

		dir.ForEach(false, func(fl quickfs.FileLike) (cancel bool, err error) {
			FileObjects = append(FileObjects, fl)
			return false, nil
		})

		context.Dir = dir
		context.ObjectList = FileObjects

		logger.Debugf("Serving directory '%s' in project '%s'", dir.Name, proj.Name)

		if err = a.executeServeTemplate(w, "dir.tmpl", &context); err != nil {
			logger.Errorf("Failed to render directory object in project '%s': %v", proj.Name, err)
			http.Error(w, "Failed to render directory in project", http.StatusInternalServerError)
		}

		return
	}

	var b = new(bytes.Buffer)

	var file = fileLike.(*quickfs.FSFile)
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

	context.File = file
	context.Content = content

	logger.Debugf("Serving file '%s' in project '%s'", file.Name, proj.Name)

	if err = a.executeServeTemplate(w, "file.tmpl", &context); err != nil {
		logger.Errorf("Failed to render file object in project '%s': %v", proj.Name, err)
		http.Error(w, "Failed to render file in project", http.StatusInternalServerError)
	}
}

func (a *App) serveProjectConfig(w http.ResponseWriter, r *http.Request) {
	var pathParts = strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	var configFile, err = os.ReadFile(
		GetProjectDirectoryPath(path.Join(pathParts[0], config.PROJECT_CONFIG_NAME), true),
	)
	if err != nil {
		logger.Errorf("Failed to open project '%s': %v", pathParts[0], err)
		http.Error(w, "Failed to open project", http.StatusInternalServerError)
		return
	}

	var ctx = &ProjectTemplateContext{
		// Fake for breadcrumbs
		Dir: &quickfs.FSDirectory{
			Name: pathParts[0],
		},
		// Config YAML data
		Content: string(configFile),
	}

	if err = a.executeServeTemplate(w, "file.tmpl", ctx); err != nil {
		logger.Errorf("Failed to render project config: %v", err)
		http.Error(w, "Failed to render project config", http.StatusInternalServerError)
	}
}

func (a *App) serveFavicon(w http.ResponseWriter, r *http.Request) {
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
}

func (a *App) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		for _, hook := range goldcrest.Get[AppServeHook](HookQuickGoServer) {
			if served, err := hook(a, w, r); err != nil {
				logger.Errorf("Failed to serve: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			} else if served {
				logger.Debugf("'%s' was served and hijacked by a hook", r.URL.Path)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

type LogHandler struct {
	Handler http.Handler
	Level   logger.LogLevel
	Where   string
}

func (h *LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var s = fmt.Sprintf("Serving %s to %s on path '%s'", h.Where, r.RemoteAddr, r.URL.Path)
	switch h.Level {
	case logger.DebugLevel:
		logger.Debug(s)
	case logger.InfoLevel:
		logger.Info(s)
	case logger.WarnLevel:
		logger.Warn(s)
	case logger.ErrorLevel:
		logger.Error(s)
	}
	h.Handler.ServeHTTP(w, r)
}

type ProjectTemplateContext struct {
	Project    *config.Project
	File       *quickfs.FSFile
	Dir        *quickfs.FSDirectory
	Parent     *quickfs.FSDirectory
	ObjectList any
	Content    string
}

func (a *App) executeServeTemplate(w http.ResponseWriter, name string, context *ProjectTemplateContext) (err error) {
	var tpl = html_template.New("base")

	tpl = tpl.Funcs(html_template.FuncMap{
		"ObjectURL": func(fl quickfs.FileLike) string {
			return filepath.ToSlash(path.Join(
				"/projects",
				context.Project.Name,
				fl.GetPath(),
			))
		},
		"ProjectURL": func(project *config.Project) string {
			return filepath.ToSlash(path.Join(
				"/projects",
				project.Name,
			))
		},
		"ConfigURL": func(project *config.Project) string {
			return filepath.ToSlash(path.Join(
				"/config",
				project.Name,
			))
		},
		"FileSize": func(size int64) string {
			f_size := float64(size)
			if f_size < 1024 {
				return fmt.Sprintf("%d b", int(f_size))
			}
			f_size = f_size / 1024
			if f_size < 1024 {
				return fmt.Sprintf("%.1f KB", f_size)
			}
			f_size = f_size / 1024
			if f_size < 1024 {
				return fmt.Sprintf("%.1f MB", f_size)
			}
			f_size = f_size / 1024
			return fmt.Sprintf("%.1f GB", f_size)
		},
	})

	if tpl, err = tpl.ParseFS(embedFS, append(a.Patterns, fmt.Sprintf("_templates/%s", name))...); err != nil {
		return errors.Wrapf(err, "failed to parse template %s", name)
	}

	return tpl.ExecuteTemplate(w, name, context)
}
