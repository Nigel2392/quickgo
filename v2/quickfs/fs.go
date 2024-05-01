package quickfs

import (
	"encoding/gob"
	"io"
)

type (
	FileLike interface {
		// Name returns the name of the file.
		GetName() string

		// Path returns the path of the file.
		GetPath() string
	}

	Directory interface {
		FileLike

		Find(path []string) (FileLike, error)

		ForEach(func(FileLike) (cancel bool, err error)) (cancel bool, err error)
	}

	File interface {
		FileLike

		// Read reads the file content.
		io.Reader
	}
)

func init() {
	gob.Register(&FSDirectory{})
	gob.Register(&FSFile{})
}
