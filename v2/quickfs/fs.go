package quickfs

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"io"
	"unicode/utf8"
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

		Load() error
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

func IsText(data []byte) bool {
	var (
		fileReader  = bytes.NewReader(data)
		fileScanner = bufio.NewScanner(fileReader)
	)

	fileScanner.Split(bufio.ScanLines)
	fileScanner.Scan()

	return utf8.Valid(fileScanner.Bytes())
}
