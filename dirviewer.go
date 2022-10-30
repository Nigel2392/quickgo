package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
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
	http.Handle("/static/", getStaticHandler())
	http.HandleFunc("/", v.http_DirBrowser)
	fmt.Println(Craft(CMD_BRIGHT_Blue, "Serving on http://localhost:8000"))
	// Open browser to localhost:8000
	if openBrowser {
		err := OpenBrowser("http://localhost:8000")
		if err != nil {
			fmt.Println(Craft(CMD_BRIGHT_Red, "Error opening browser: "+err.Error()))
		}
	}
	return http.ListenAndServe("127.0.0.1:8000", nil)
}

func (v *Viewer) GetIndexTemplate() (*template.Template, error) {
	return template.ParseFS(TemplateFS, v.templates...)
}

func getStaticHandler() http.Handler {
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
			HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{FileContent: file.Content, IsFile: true})
			return
		} else {
			fmt.Fprint(w, file.Content)
			return
		}
	}
	dir.Children = SortDirs(dir.Children)
	dir.Files = SortFiles(dir.Files)
	HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{Dir: dir})
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

func GetDirs(str_dirs []string) []Directory {
	var dirs []Directory
	var wg sync.WaitGroup
	var out_mu sync.Mutex

	wg.Add(len(str_dirs))
	for _, dir := range str_dirs {
		go func(dirname string, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			proj_name := strings.SplitN(dirname, ".", 2)[0]
			proj_name = strings.ToUpper(proj_name)
			dir, err := GetDir(dirname, proj_name)
			if err != nil {
				fmt.Println(err)
				return
			}
			mu.Lock()
			dirs = append(dirs, dir)
			mu.Unlock()
		}(dir, &wg, &out_mu)
	}
	wg.Wait()

	return dirs
}

func OpenBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func SortDirs(dirs []Directory) []Directory {
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name < dirs[j].Name
	})
	return dirs
}

func SortFiles(files []File) []File {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	return files
}
