package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

func (v *Viewer) serve(openBrowser bool) error {
	http.Handle("/static/", v.getStaticHandler())
	http.HandleFunc("/readme.md", v.readmeHandler)
	http.HandleFunc("/favicon.ico", v.iconHandler)
	http.HandleFunc("/", v.directoryHandler)
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

func (v *Viewer) getStaticHandler() http.Handler {
	static_fs, _ := fs.Sub(fs.FS(TemplateFS), "templates")
	return http.StripPrefix("/static/", http.FileServer(http.FS(static_fs)))
}

func (v *Viewer) iconHandler(w http.ResponseWriter, r *http.Request) {
	ico, err := TemplateFS.ReadFile("templates/quickgo.png")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	w.Write(ico)
}

func (v *Viewer) readmeHandler(w http.ResponseWriter, r *http.Request) {
	readme, err := TemplateFS.ReadFile("readme.md")
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	tpl, err := v.templates["readme"].Clone()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	// Markdown
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.DefinitionList),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(readme, &buf); err != nil {
		panic(err)
	}
	tpl.ExecuteTemplate(w, "readme.tmpl", TemplateData{FileContent: buf.String()})
}

func (v *Viewer) directoryHandler(w http.ResponseWriter, r *http.Request) {
	var url string = r.URL.Path
	HTML_TEMPLATE, err := v.templates["index"].Clone()
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
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
	dir, file, file_found, err := v.TraverseDirFromPath(dir, path[1:])
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
