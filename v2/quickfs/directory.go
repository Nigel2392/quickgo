package quickfs

import (
	"errors"
	"os"
	"path"
)

var (
	ErrFileLikeExcluded = errors.New("file like object was excluded")
)

type FSDirectory struct {
	// Name of the directory.
	Name string

	// Path of the directory.
	Path string

	// Files in the directory.
	Files map[string]File

	// Directories in the directory.
	Directories map[string]*FSDirectory

	// Root directory.
	root *FSDirectory

	// IsExcluded returns true if the directory is excluded.
	// It should only be set on the root directory.
	IsExcluded func(*FSDirectory) bool
}

// NewFSDirectory creates a new FSDirectory.
func NewDirectory(name, path string) (Directory, error) {
	return NewFSDirectory(name, path, nil)
}

func NewFSDirectory(name, filepath string, root *FSDirectory) (*FSDirectory, error) {
	var d *FSDirectory = &FSDirectory{
		Name:        name,
		Path:        filepath,
		Files:       make(map[string]File),
		Directories: make(map[string]*FSDirectory),
		root:        root,
	}

	if root == nil {
		root = d
	}

	if root.IsExcluded != nil && root.IsExcluded(d) {
		return nil, ErrFileLikeExcluded
	}

	var dirs, err = os.ReadDir(filepath)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		var (
			n   = dir.Name()
			f   *FSFile
			sub *FSDirectory
		)
		if dir.IsDir() {
			sub, err = NewFSDirectory(
				n, path.Join(filepath, n),
				root,
			)
		} else {
			f, err = NewFSFile(
				n, path.Join(filepath, n),
			)
		}
		if err != nil && errors.Is(err, ErrFileLikeExcluded) {
			continue
		} else if err != nil {
			return nil, err
		}

		if sub != nil {
			d.Directories[n] = sub
		} else {
			d.Files[n] = f
		}
	}

	return d, nil
}

func (d *FSDirectory) GetName() string {
	return d.Name
}

func (d *FSDirectory) GetPath() string {
	return d.Path
}

func (d *FSDirectory) Find(path []string) (FileLike, error) {
	if len(path) == 0 {
		return d, nil
	}

	var (
		dir = d
		ok  bool
	)

	for _, name := range path {
		if dir, ok = dir.Directories[name]; ok {
			return dir.Find(path[1:])
		} else if f, ok := dir.Files[name]; ok && len(path) == 1 {
			return f, nil
		}
	}

	return nil, os.ErrNotExist
}

func (d *FSDirectory) ForEach(fn func(FileLike) (cancel bool, err error)) (cancel bool, err error) {
	cancel, err = fn(d)
	if cancel || err != nil {
		return
	}

	for _, dir := range d.Directories {
		cancel, err = dir.ForEach(fn)
		if cancel || err != nil {
			return
		}
	}

	for _, f := range d.Files {
		cancel, err = fn(f)
		if cancel || err != nil {
			return
		}
	}

	return
}
