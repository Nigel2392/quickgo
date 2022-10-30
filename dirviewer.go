package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"text/template"
)

//go:embed templates/*
var TemplateFS embed.FS

type Viewer struct {
	Dirs []Directory
}

type TemplateData struct {
	Dirs            []Directory
	Dir             Directory
	DirEmpty        bool
	FileContent     string
	HasMultipleDirs bool
	IsFile          bool
}

func NewViewer(str_dirs []string) *Viewer {
	dirs := GetDirs(str_dirs)
	return &Viewer{Dirs: dirs}
}

func GetIndexTemplate() (*template.Template, error) {
	return template.ParseFS(TemplateFS, "templates/index.tmpl", "templates/base.tmpl")
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
	HTML_TEMPLATE, err := GetIndexTemplate()
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
	}
	if url == "/" {
		//fmt.Fprint(w, "<h1 style=\"color:#9200ff;font-style: helvetica;\">Templates: </h1>")
		//for _, dir := range v.Dirs {
		//	fmt.Fprintf(w, "<a href='%s/' style=\"font-size:1.5em;font-weight:bold;text-decoration:none;color:#ab22ff;\">%s</a><br>", dir.Name, dir.Name)
		//}
		// Load index.html
		HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{Dirs: v.Dirs, HasMultipleDirs: true})
		return
	} else {
		url = strings.Trim(url, "/")
	}
	path := strings.Split(url, "/")
	if path[0] == "static" {
		getStaticHandler().ServeHTTP(w, r)
		return
	}
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
		//if strings.Contains(file.Content, "<html>") && strings.Contains(file.Content, "</html>") {
		//	data = fmt.Sprintf(bare_html, file.Name, "", html.EscapeString(file.Content))
		//}
		if strings.Contains(http.DetectContentType([]byte(file.Content)), "text/plain") {
			HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{FileContent: file.Content, IsFile: true})
			return
		} else {
			fmt.Fprint(w, file.Content)
			return
		}
	}
	//for _, child := range dir.Children {
	//	fmt.Fprintf(w, "<a href=\"%s/\" style=\"font-size:1.2em;font-weight:bold;text-decoration:none;color:#9200ff;\">%s</a><br>", child.Name, child.Name)
	//}
	//for _, file := range dir.Files {
	//	fmt.Fprintf(w, "<a href=\"%s/\" style=\"font-size:1em;font-weight:bold;text-decoration:none;\">%s</a><br>", file.Name, file.Name)
	//}
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
