package main

import (
	"embed"
	"fmt"
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
	FileContent string
	Datasize    string
	DirEmpty    bool
	IsRoot      bool
	IsFile      bool
	Raw         bool
	ShowPreview bool
	Name        string
}

func NewViewer(str_dirs []string, raw bool) *Viewer {
	dirs := GetDirs(str_dirs, raw)
	v := &Viewer{
		Dirs: dirs,
		bases: []string{
			"templates/base.tmpl",
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

func FindDir(dir *Directory, path []string) (*Directory, error) {
	if len(path) == 0 {
		return dir, nil
	}
	for _, child := range dir.Children {
		if child.Name == path[0] {
			return FindDir(&child, path[1:])
		}
	}
	return nil, fmt.Errorf("Directory not found")
}
