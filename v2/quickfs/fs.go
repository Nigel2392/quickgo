package quickfs

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"path/filepath"
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

func printDir(w io.Writer, d *FSDirectory, indent int) {
	for _, dir := range d.Directories {
		fmt.Fprintf(w, "%s%s%s\n", strings.Repeat(" ", indent), dir.GetName(), string(filepath.Separator))
		printDir(w, dir, indent+2)
	}

	for _, f := range d.Files {
		fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", indent), f.GetName())
	}
}

func PrintRoot(w io.Writer, root *FSDirectory) {
	fmt.Fprintf(w, "%s%s\n", root.GetName(), string(filepath.Separator))
	printDir(w, root, 2)
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
