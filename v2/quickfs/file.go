package quickfs

import (
	"io"
	"os"
)

type FSFile struct {
	// Name of the file.
	Name string

	// Path of the file.
	Path string

	// If the file is all valid utf-8 text.
	IsText bool

	Reader io.ReadCloser
}

// NewFSFile creates a new FSFile.
func NewFile(name, path string) (File, error) {
	return NewFSFile(name, path, nil)
}

func NewFSFile(name, path string, root *FSDirectory) (*FSFile, error) {
	var f *FSFile = &FSFile{
		Name: name,
		Path: path,
	}

	if root != nil && root.IsExcluded != nil && root.IsExcluded(f) {
		return nil, ErrFileLikeExcluded
	}

	var osF, err = os.Open(path)
	if err != nil {
		return nil, err
	}

	f.Reader = osF

	return f, nil
}

func (f *FSFile) GetName() string {
	return f.Name
}

func (f *FSFile) GetPath() string {
	return f.Path
}

func (f *FSFile) Read(p []byte) (n int, err error) {
	return f.Reader.Read(p)
}

func (f *FSFile) Close() error {
	if f.Reader != nil {
		return f.Reader.Close()
	}
	return nil
}
