package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func (f *File) SizeStr() string {
	size := len(f.Content)
	return sizeStr(size)
}

type Directory struct {
	Name     string      `json:"name"`
	Children []Directory `json:"directory"`
	Files    []File      `json:"files"`
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

func sizeStr[T int | int16 | int32 | int64](size T) string {
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
}

func FileToDir(file []byte) (Directory, error) {
	var dir Directory
	err := json.Unmarshal(file, &dir)
	if err != nil {
		return Directory{}, err
	}
	return dir, nil
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
