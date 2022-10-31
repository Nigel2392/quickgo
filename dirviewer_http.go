package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

var DEFAULT_XML_START = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
var DEFAULT_XML_END = `</urlset>`

func (v *Viewer) serve(openBrowser bool) error {
	http.Handle("/static/", v.getStaticHandler())
	http.HandleFunc("/readme.md", v.readmeHandler)
	http.HandleFunc("/favicon.ico", v.iconHandler)
	http.HandleFunc("/sitemap.xml", v.SitemapHandler)
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

func (v *Viewer) SitemapHandler(w http.ResponseWriter, r *http.Request) {
	// Get max depth to calculate the lowest possible priority ahead of time.
	depth := GetDepth(v.Dirs)
	minus := float32(1) / float32(depth)
	// Set content type to xml
	w.Header().Set("Content-Type", "application/xml")
	// Write xml start
	fmt.Fprint(w, DEFAULT_XML_START)
	// Recursively iterate through directories and write to sitemap
	SitemapWriter(w, v.Dirs, 0, 1, minus, "http://"+AppConfig.Host+":"+AppConfig.Port+"/")
	// Write xml end
	fmt.Fprint(w, DEFAULT_XML_END)
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
	md, err := Markdownify(readme)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	tpl.ExecuteTemplate(w, "readme.tmpl", TemplateData{FileContent: md})
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
	n_dir, file, file_found, err := TraverseDirFromPath(dir, path[1:])
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
		return
	}
	if !file_found && len(n_dir.Children) == 0 && len(n_dir.Files) == 0 {
		HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{DirEmpty: true})
		return
	}
	if file_found {
		if strings.Contains(http.DetectContentType([]byte(file.Content)), "text/plain") {
			var content string = file.Content
			var err error
			if strings.HasSuffix(strings.ToLower(file.Name), ".md") {
				content, err = Markdownify([]byte(content))
				if err != nil {
					fmt.Fprint(w, err.Error())
					return
				}
			}
			HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{FileContent: content, IsFile: true, Datasize: sizeStr(len(file.Content))})
			return
		} else {
			w.Header().Set("Content-Type", http.DetectContentType([]byte(file.Content)))
			fmt.Fprint(w, file.Content)
			return
		}
	}
	n_dir.Children = SortDirs(n_dir.Children)
	n_dir.Files = SortFiles(n_dir.Files)
	HTML_TEMPLATE.ExecuteTemplate(w, "index.tmpl", TemplateData{Dir: n_dir, Datasize: n_dir.SizeStr()})
}

// Generate a sitemap from directories
func SitemapWriter(w http.ResponseWriter, dirs []Directory, margin int, priority, minus float32, path string) {
	priority -= minus
	for _, dir := range dirs {
		fmt.Fprintf(w, `<url>
	<loc>%s</loc>
	<size>%s</size>
	<priority>%.2f</priority>
</url>`, path+dir.Name+"/", dir.SizeStr(), priority)
		if len(dir.Children) > 0 {
			SitemapWriter(w, dir.Children, margin+20, priority, minus, path+dir.Name+"/")
		}
		for _, file := range dir.Files {
			fmt.Fprintf(w, `<url>
	<loc>%s</loc>
	<size>%s</size>
	<priority>%.2f</priority>
</url>`, path+dir.Name+"/"+file.Name+"/", sizeStr(len(file.Content)), priority)
		}
	}
}
