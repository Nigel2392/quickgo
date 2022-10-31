package main

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed readme.md
//go:embed templates/*
var TemplateFS embed.FS

type Viewer struct {
	Dirs      []Directory
	bases     []string
	templates map[string]*template.Template
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

func NewViewer(str_dirs []string, raw bool) *Viewer {
	dirs := GetDirs(str_dirs, raw)
	v := &Viewer{
		Dirs: dirs,
		bases: []string{
			"templates/base.tmpl",
			"templates/dir_display.tmpl",
			"templates/go_back.tmpl",
		},
		templates: make(map[string]*template.Template),
	}
	tpls := []string{
		"index",
		"readme",
	}
	for _, tpl := range tpls {
		t, err := template.ParseFS(TemplateFS, append(v.bases, fmt.Sprintf("templates/%s.tmpl", tpl))...)
		if err != nil {
			panic(err)
		}
		v.templates[tpl] = t
	}
	return v
}

func (v *Viewer) TraverseDirFromPath(dir Directory, path []string) (Directory, File, bool, error) {
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
			return v.TraverseDirFromPath(child, path[1:])
		}
	}
	return Directory{}, File{}, false, fmt.Errorf("path not found")
}
