package main

import (
	"fmt"
	"strings"
	"sync"
)

type FileSystem interface {
	GetName() string
}

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (f *File) SizeStr() string {
	size := len(f.Content)
	return sizeStr(size)
}

func (f *File) GetName() string {
	return f.Name
}

type Directory struct {
	Name     string      `json:"name"`
	Children []Directory `json:"directory"`
	Files    []File      `json:"files"`
}

func (d *Directory) GetName() string {
	return d.Name
}

func (d *Directory) GetSortedFiles() []Directory {
	data := SortDirs(d.Children)
	return data
}

func (d *Directory) GetSortedChildren() []File {
	data := SortFiles(d.Files)
	return data
}

func (d *Directory) Size() int64 {
	var size int64
	for _, file := range d.Files {
		size += int64(len(file.Content))
	}
	for _, child := range d.Children {
		size += child.Size()
	}
	return int64(size)
}
func (d *Directory) SizeStr() string {
	size := d.Size()
	return sizeStr(size)
}

func GetDirs(str_dirs []string, raw bool) []Directory {
	var dirs []Directory
	var wg sync.WaitGroup
	var out_mu sync.Mutex

	wg.Add(len(str_dirs))
	for _, str_dir := range str_dirs {
		go func(dirname string, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			proj_name := strings.SplitN(dirname, ".", 2)[0]
			proj_name = strings.ToUpper(proj_name)
			dir, err := GetDir(dirname, proj_name, raw)
			if err != nil {
				fmt.Println(err)
				return
			}
			mu.Lock()
			dir.Name = dirname
			dirs = append(dirs, dir)
			mu.Unlock()
		}(str_dir, &wg, &out_mu)
	}
	wg.Wait()

	return dirs
}

func TraverseDirFromPath(dir Directory, path []string) (Directory, File, bool, error) {
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
			return TraverseDirFromPath(child, path[1:])
		}
	}
	return Directory{}, File{}, false, fmt.Errorf("path not found")
}

func RenameDirData(dir Directory, project_name string) Directory {
	if project_name == "" {
		project_name = dir.Name
	}
	dir.Name = ReplaceNamesString(dir.Name, project_name)
	for i, file := range dir.Files {
		dir.Files[i].Name = ReplaceNamesString(file.Name, project_name)
		dir.Files[i].Content = ReplaceNamesString(file.Content, project_name)
	}
	for i, child := range dir.Children {
		dir.Children[i] = RenameDirData(child, project_name)
	}
	return dir
}
