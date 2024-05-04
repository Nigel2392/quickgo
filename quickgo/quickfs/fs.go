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

		// Find a directory by path.
		Find(path []string) (FileLike, error)

		// Traverse traverses the directory tree.
		// It will traverse into subdirectories.
		Traverse(fn func(FileLike) (cancel bool, err error)) (cancel bool, err error)

		// ForEach loops over all directories and files in this directory.
		// This will not traverse into subdirectories and is not recursive.
		// execRoot will let execute the function on the directory itself if true (as well as its direct children)
		ForEach(execRoot bool, fn func(FileLike) (cancel bool, err error)) (cancel bool, err error)

		// Load loads the directory content.
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
	count += 1 + d.Files.Len()

	for dir := d.Directories.Front(); dir != nil; dir = dir.Next() {
		fmt.Fprintf(w,
			"%s%s\n",
			strings.Repeat(indentString, indent*2),
			wrap(indent, dir.Value),
		)
		count += printDir(w, dir.Value, indent+1, indentString, wrap)
	}

	for f := d.Files.Front(); f != nil; f = f.Next() {
		fmt.Fprintf(w, "%s%s\n", strings.Repeat(indentString, indent*2), wrap(indent, f.Value))
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
