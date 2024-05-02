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

	readFrom bool
}

// NewFSFile creates a new FSFile.
func NewFile(name, path string) (File, error) {
	return NewFSFile(name, path, nil)
}

// NewFSFile creates a new FSFile.
// If root is not nil, it will check if the file is excluded.
// The file must always be closed after calling this function.
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

func (f *FSFile) IsDir() bool {
	return false
}

func (f *FSFile) GetName() string {
	return f.Name
}

func (f *FSFile) GetPath() string {
	return f.Path
}

func (f *FSFile) Read(p []byte) (n int, err error) {
	if f.readFrom {
		return 0, io.EOF
	}
	f.readFrom = true
	return f.Reader.Read(p)
}

func (f *FSFile) Close() error {
	if f.Reader != nil {
		return f.Reader.Close()
	}
	return nil
}
