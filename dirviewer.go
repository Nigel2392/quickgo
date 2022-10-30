package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"text/template"
)

//go:embed templates/*
var TemplateFS embed.FS

type Viewer struct {
	Dirs      []Directory
	templates []string
}

type TemplateData struct {
	Dirs        []Directory
	Dir         Directory
	DirEmpty    bool
	FileContent string
	IsRoot      bool
	IsFile      bool
	Datasize    string
}

func NewViewer(str_dirs []string) *Viewer {
	dirs := GetDirs(str_dirs)
	return &Viewer{
		Dirs: dirs,
		templates: []string{
			"templates/base.tmpl",
			"templates/dir_display.tmpl",
			"templates/go_back.tmpl",
			"templates/index.tmpl",
		},
	}
}

func (v *Viewer) serve(openBrowser bool) error {
	http.Handle("/static/", v.getStaticHandler())
	http.HandleFunc("/", v.http_DirBrowser)
	fmt.Println(Craft(CMD_BRIGHT_Blue, "Serving on http://"+AppConfig.Host+":"+AppConfig.Port))
	// Open browser to localhost:8000
	if openBrowser {
		err := OpenBrowser("http://" + AppConfig.Host + ":" + AppConfig.Port)
		if err != nil {
			fmt.Println(Craft(CMD_BRIGHT_Red, "Error opening browser: "+err.Error()))
		}
	}
	return http.ListenAndServe(AppConfig.Host+":"+AppConfig.Port, nil)
}

func (v *Viewer) GetIndexTemplate() (*template.Template, error) {
	return template.ParseFS(TemplateFS, v.templates...)
}

func (v *Viewer) getStaticHandler() http.Handler {
	static_fs, _ := fs.Sub(fs.FS(TemplateFS), "templates")
	return http.StripPrefix("/static/", http.FileServer(http.FS(static_fs)))
}

func (v *Viewer) http_DirBrowser(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon") {
		return
	}
	var url string = r.URL.Path
	HTML_TEMPLATE, err := v.GetIndexTemplate()
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
	}
	if url == "/" {
		dirs := SortDirs(v.Dirs)
		HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{Dirs: dirs, IsRoot: true})
		return
	}
	url = strings.Trim(url, "/")
	path := strings.Split(url, "/")
	var dir Directory
	for _, l_dir := range v.Dirs {
		if l_dir.Name == path[0] {
			dir = l_dir
			break
		}
	}
	dir, file, file_found, err := v.http_TraverseDir(dir, path[1:])
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	if !file_found && len(dir.Children) == 0 && len(dir.Files) == 0 {
		HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{DirEmpty: true})
		return
	}
	if file_found {
		if strings.Contains(http.DetectContentType([]byte(file.Content)), "text/plain") {
			HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{FileContent: file.Content, IsFile: true, Datasize: sizeStr(len(file.Content))})
			return
		} else {
			fmt.Fprint(w, file.Content)
			return
		}
	}
	dir.Children = SortDirs(dir.Children)
	dir.Files = SortFiles(dir.Files)
	HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{Dir: dir, Datasize: dir.SizeStr()})
}

func (v *Viewer) http_TraverseDir(dir Directory, path []string) (Directory, File, bool, error) {
	if len(path) == 0 {
		return dir, File{}, false, nil
	} else if len(path) > 0 {
		for _, file := range dir.Files {
			if strings.EqualFold(file.Name, path[0]) {
				return dir, file, true, nil
			}
		}
	}
	for _, child := range dir.Children {
		if strings.EqualFold(child.Name, path[0]) {
			for _, file := range child.Files {
				if strings.EqualFold(file.Name, path[0]) {
					return child, file, true, nil
				}
			}
			return v.http_TraverseDir(child, path[1:])
		}
	}
	return Directory{}, File{}, false, fmt.Errorf("path not found")
}
