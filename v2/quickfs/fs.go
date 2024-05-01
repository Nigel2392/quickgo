package quickfs

import "io"

type (
	FileLike interface {
		// Name returns the name of the file.
		Name() string

		// Path returns the path of the file.
		Path() string
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
