package quickfs

import "os"

type FSFile struct {
	// Name of the file.
	Name string

	// Path of the file.
	Path string

	// Content of the file.
	Content []byte

	// If the file is all valid utf-8 text.
	IsText bool

	f *os.File
}

// NewFSFile creates a new FSFile.
func NewFile(name, path string) (File, error) {
	return NewFSFile(name, path)
}

func NewFSFile(name, path string) (*FSFile, error) {
	var f *FSFile = &FSFile{
		Name: name,
		Path: path,
	}

	var file, err = os.Open(path)
	if err != nil {
		return nil, err
	}

	f.f = file

	return f, nil
}

func (f *FSFile) GetName() string {
	return f.Name
}

func (f *FSFile) GetPath() string {
	return f.Path
}

func (f *FSFile) Read(p []byte) (n int, err error) {
	return copy(p, f.Content), nil
}

func (f *FSFile) Size() int64 {
	return int64(len(f.Content))
}
