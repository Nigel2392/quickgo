package quickfs

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type (
	FileLike interface {
		// Name returns the name of the file.
		GetName() string

		// Path returns the path of the file.
		GetPath() string

		// IsDir returns true if the file is a directory.
		IsDir() bool
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

func printDir(w io.Writer, d *FSDirectory, indent int, indentString string, wrap func(int, FileLike) string) (count int) {
	count += 1 + len(d.Files)

	for _, dir := range d.Directories {
		fmt.Fprintf(w,
			"%s%s\n",
			strings.Repeat(indentString, indent*2),
			wrap(indent, dir),
		)
		count += printDir(w, dir, indent+1, indentString, wrap)
	}

	for _, f := range d.Files {
		fmt.Fprintf(w, "%s%s\n", strings.Repeat(indentString, indent*2), wrap(indent, f))
	}

	return count
}

func PrintRoot(w io.Writer, root *FSDirectory) int {
	return PrintRootFn(w, root, " ", func(_ int, fl FileLike) string {
		return fl.GetName()
	})
}

func PrintRootFn(w io.Writer, root *FSDirectory, indentString string, wrap func(int, FileLike) string) int {
	fmt.Fprintf(w, "%s\n", wrap(0, root))
	return 1 + printDir(w, root, 1, indentString, wrap)
}

func IsTextReader(r io.Reader) bool {
	var (
		fileScanner = bufio.NewScanner(r)
	)

	fileScanner.Split(bufio.ScanLines)
	fileScanner.Scan()

	return utf8.Valid(fileScanner.Bytes())
}

func IsText[T string | []byte](data T) bool {
	return IsTextReader(bytes.NewReader([]byte(data)))
}
