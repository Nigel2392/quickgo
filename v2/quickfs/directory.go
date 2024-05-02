package quickfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	IsExcluded func(FileLike) bool
}

// Root returns the root directory.
func (d *FSDirectory) Root() *FSDirectory {
	return d.root
}

// String returns the directory in string format.
func (d *FSDirectory) String() string {
	var b strings.Builder
	b.WriteString(d.Name)
	b.WriteString(":\n")
	for _, dir := range d.Directories {
		b.WriteString("   ")
		b.WriteString(dir.GetName())
		b.WriteString("\n")
	}
	for _, f := range d.Files {
		b.WriteString("  ")
		b.WriteString(f.GetName())
		b.WriteString("\n")
	}
	return b.String()
}

// NewFSDirectory creates a new FSDirectory.
func NewDirectory(name, path string) Directory {
	return NewFSDirectory(name, path, nil)
}

func NewFSDirectory(name, dirPath string, root *FSDirectory) *FSDirectory {
	return &FSDirectory{
		Name:        name,
		Path:        dirPath,
		Files:       make(map[string]File),
		Directories: make(map[string]*FSDirectory),
		root:        root,
	}
}

func (f *FSDirectory) IsDir() bool {
	return true
}

func (d *FSDirectory) Load() error {

	if d.root != nil && d.root.IsExcluded != nil && d.root.IsExcluded(d) {
		return ErrFileLikeExcluded
	}

	var dirs, err = os.ReadDir(d.Path)
	if err != nil {
		return err
	}

	var root = d.root
	if root == nil {
		root = d
	}
loop:
	for _, dir := range dirs {
		var (
			n   = dir.Name()
			p   = path.Join(d.Path, n)
			f   *FSFile
			sub *FSDirectory
		)

		if dir.IsDir() {
			sub = NewFSDirectory(
				n, p, root,
			)
			err = sub.Load()
		} else {
			f, err = NewFSFile(
				n, p, root,
			)
		}
		if err != nil && errors.Is(err, ErrFileLikeExcluded) {
			continue loop
		} else if err != nil {
			return err
		}

		if sub != nil {
			d.Directories[n] = sub
		} else {
			d.Files[n] = f
		}
	}

	return nil
}

func (d *FSDirectory) GetName() string {
	return d.Name
}

func (d *FSDirectory) GetPath() string {
	return d.Path
}

func (d *FSDirectory) Find(path []string) (FileLike, error) {
	if path == nil {
		return nil, &fs.PathError{
			Op:   "find",
			Path: d.Path,
			Err:  os.ErrInvalid,
		}
	}

	if len(path) == 0 {
		return d, nil
	}

	var name = path[0]
	var dir, ok = d.Directories[name]
	if ok {
		return dir.Find(path[1:])
	}

	f, ok := d.Files[name]
	if ok && len(path) == 1 {
		return f, nil
	}

	return nil, os.ErrNotExist
}

func (d *FSDirectory) AddDirectory(dirPath string) {
	dirPath = filepath.ToSlash(dirPath)
	dirPath = filepath.FromSlash(dirPath)
	dirPath = filepath.Clean(dirPath)

	var (
		ok    bool
		_d    *FSDirectory
		root  *FSDirectory = d.root
		parts              = strings.Split(dirPath, string(os.PathSeparator))
		dir                = d
	)

	if root == nil {
		root = d
	}

	for _, part := range parts {
		if _d, ok = dir.Directories[part]; !ok {
			_d = NewFSDirectory(
				part, filepath.Join(dir.Path, part), root,
			)
			dir.Directories[part] = _d
		}

		dir = _d
	}

}

func (d *FSDirectory) AddFile(filePath string, reader io.ReadCloser) {
	filePath = filepath.ToSlash(filePath)
	filePath = filepath.FromSlash(filePath)
	filePath = filepath.Clean(filePath)
	var (
		parts = strings.Split(filePath, string(os.PathSeparator))
		dir   = d
	)
	for i := 0; i < len(parts)-1; i++ {
		var part = parts[i]
		if _, ok := dir.Directories[parts[i]]; !ok {
			dir.Directories[part] = NewFSDirectory(
				part, filepath.Join(dir.Path, part), d.root,
			)
		}
		dir = dir.Directories[part]
	}

	dir.Files[parts[len(parts)-1]] = &FSFile{
		Name:   parts[len(parts)-1],
		Path:   filePath,
		Reader: reader,
	}
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
