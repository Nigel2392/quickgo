package main

import (
	"fmt"
	"html"
	"net/http"
	"strings"
	"sync"
)

var bare_html = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>%s</title>
	<style>%s</style>
</head>
<body style="white-space: pre;">
<pre>%s</pre>
</body>
</html>
`

type Viewer struct {
	Dirs []Directory
}

func NewViewer(str_dirs []string) *Viewer {
	dirs := GetDirs(str_dirs)
	return &Viewer{Dirs: dirs}
}

func (v *Viewer) http_DirBrowser(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "favicon") {
		return
	}
	var url string = r.URL.Path
	if url == "/" {
		for _, dir := range v.Dirs {
			fmt.Fprintf(w, "<a href='%s/' style=\"font-size:1.5em;font-weight:bold;text-decoration:none;\">%s</a><br>", dir.Name, dir.Name)
		}
		return
	} else {
		url = strings.Trim(url, "/")
	}
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
		fmt.Fprintf(w, "Sorry, this directory is empty!")
		return
	}
	if file_found {
		var data string = string(file.Content)
		if strings.Contains(file.Content, "<html>") && strings.Contains(file.Content, "</html>") {
			data = fmt.Sprintf(bare_html, file.Name, "", html.EscapeString(file.Content))
		}
		fmt.Fprint(w, data)

		fmt.Fprint(w, data)
		return
	}
	for _, child := range dir.Children {
		fmt.Fprintf(w, "<a href=\"%s/\" style=\"font-size:1.2em;font-weight:bold;text-decoration:none;color:#9200ff;\">%s</a><br>", child.Name, child.Name)
	}
	for _, file := range dir.Files {
		fmt.Fprintf(w, "<a href=\"%s/\" style=\"font-size:1em;font-weight:bold;text-decoration:none;\">%s</a><br>", file.Name, file.Name)
	}
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

func (v *Viewer) serve() error {
	http.HandleFunc("/", v.http_DirBrowser)
	fmt.Println(Craft(CMD_BRIGHT_Blue, "Serving on port http://localhost:8000"))
	return http.ListenAndServe("127.0.0.1:8000", nil)
}

func GetDirs(str_dirs []string) []Directory {
	var dirs []Directory
	var wg sync.WaitGroup
	var murw sync.Mutex

	wg.Add(len(str_dirs))
	for _, dir := range str_dirs {
		go func(dirname string, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			dir, err := GetDir(dirname, "")
			if err != nil {
				fmt.Println(err)
				return
			}
			murw.Lock()
			dirs = append(dirs, dir)
			murw.Unlock()
		}(dir, &wg, &murw)
	}
	wg.Wait()

	return dirs
}
