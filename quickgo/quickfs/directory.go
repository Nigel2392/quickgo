package quickfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
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
	Files *orderedmap.OrderedMap[string, *FSFile]

	// Directories in the directory.
	Directories *orderedmap.OrderedMap[string, *FSDirectory]

	// Root directory.
	root *FSDirectory

	// Size of the directory.
	size int64

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
	// for _, dir := range d.Directories {
	// 	b.WriteString("   ")
	// 	b.WriteString(dir.GetName())
	// 	b.WriteString("\n")
	// }
	// for _, f := range d.Files {
	// 	b.WriteString("  ")
	// 	b.WriteString(f.GetName())
	// 	b.WriteString("\n")
	// }
	for el := d.Directories.Front(); el != nil; el = el.Next() {
		b.WriteString("   ")
		b.WriteString(el.Value.GetName())
		b.WriteString("\n")
	}
	for el := d.Files.Front(); el != nil; el = el.Next() {
		b.WriteString("  ")
		b.WriteString(el.Value.GetName())
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
		Files:       orderedmap.NewOrderedMap[string, *FSFile](),
		Directories: orderedmap.NewOrderedMap[string, *FSDirectory](),
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
			d.Directories.Set(n, sub)
		} else {
			d.Files.Set(n, f)
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
	var _, fl, err = d.find(nil, path)
	return fl, err
}

func (d *FSDirectory) FindWithParent(path []string) (*FSDirectory, FileLike, error) {
	return d.find(nil, path)
}

func (d *FSDirectory) find(parent *FSDirectory, path []string) (*FSDirectory, FileLike, error) {
	if path == nil {
		return nil, nil, &fs.PathError{
			Op:   "find",
			Path: d.Path,
			Err:  os.ErrInvalid,
		}
	}

	if len(path) == 0 {
		return parent, d, nil
	}

	var name = path[0]
	var dir, ok = d.Directories.Get(name)
	if ok {
		return dir.find(d, path[1:])
	}

	f, ok := d.Files.Get(name)
	if ok && len(path) == 1 {
		return d, f, nil
	}

	return nil, nil, os.ErrNotExist
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
		if _d, ok = dir.Directories.Get(part); !ok {
			_d = NewFSDirectory(
				part, filepath.Join(dir.Path, part), root,
			)
			dir.Directories.Set(part, _d)
		}

		dir = _d
	}
}

func (d *FSDirectory) Size() int64 {
	if d.size > 0 {
		return d.size
	}
	for el := d.Directories.Front(); el != nil; el = el.Next() {
		d.size += el.Value.Size()
	}
	for el := d.Files.Front(); el != nil; el = el.Next() {
		d.size += el.Value.Size
	}
	return d.size

}

func (d *FSDirectory) AddFile(filePath string, reader io.ReadCloser) *FSFile {
	filePath = filepath.ToSlash(filePath)
	filePath = filepath.FromSlash(filePath)
	filePath = filepath.Clean(filePath)
	var (
		parts = strings.Split(filePath, string(os.PathSeparator))
		dir   = d
		ok    bool
	)
	for i := 0; i < len(parts)-1; i++ {
		var part = parts[i]
		if d, ok = dir.Directories.Get(part); !ok {
			d = NewFSDirectory(
				part, filepath.Join(dir.Path, part), dir.root,
			)
			dir.Directories.Set(part, d)
		}

		dir = d
	}

	var f = &FSFile{
		Name:   parts[len(parts)-1],
		Path:   filePath,
		Reader: reader,
	}
	// dir.Files[parts[len(parts)-1]] = f
	dir.Files.Set(f.Name, f)
	return f
}

// Traverse traverses the directory tree.
// It will traverse into subdirectories.
func (d *FSDirectory) Traverse(fn func(FileLike) (cancel bool, err error)) (cancel bool, err error) {
	if cancel, err = fn(d); cancel || err != nil {
		return
	}
	for el := d.Directories.Front(); el != nil; el = el.Next() {
		if cancel, err = el.Value.Traverse(fn); cancel || err != nil {
			return
		}
	}
	for el := d.Files.Front(); el != nil; el = el.Next() {
		if cancel, err = fn(el.Value); cancel || err != nil {
			return
		}
	}
	return
}

// ForEach loops over all directories and files in this directory.
// This will not traverse into subdirectories and is not recursive.
func (d *FSDirectory) ForEach(execRoot bool, fn func(FileLike) (cancel bool, err error)) (cancel bool, err error) {
	if execRoot {
		if cancel, err = fn(d); err != nil || cancel {
			return
		}
	}

	for el := d.Directories.Front(); el != nil; el = el.Next() {
		if cancel, err = fn(el.Value); err != nil || cancel {
			return
		}
	}

	for el := d.Files.Front(); el != nil; el = el.Next() {
		if cancel, err = fn(el.Value); err != nil || cancel {
			return
		}
	}
	return
}
