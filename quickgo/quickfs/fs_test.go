package quickfs_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nigel2392/quickgo/v2/quickgo/quickfs"
	"github.com/elliotchance/orderedmap/v2"
)

func makeFiles(path string, count int) *orderedmap.OrderedMap[string, *quickfs.FSFile] {
	var files = orderedmap.NewOrderedMap[string, *quickfs.FSFile]()
	for i := 1; i < count+1; i++ {
		var n = fmt.Sprintf("file%d", i)
		files.Set(n, &quickfs.FSFile{
			Name: n,
			Path: filepath.Join(path, n),
		})
	}
	return files
}

func getDirectoryMap() *orderedmap.OrderedMap[string, *quickfs.FSDirectory] {
	var m = orderedmap.NewOrderedMap[string, *quickfs.FSDirectory]()
	var dir = &quickfs.FSDirectory{
		Name:  "dir1", // 4
		Path:  filepath.Join("root", "dir1"),
		Files: makeFiles(filepath.Join("root", "dir1"), 2), // 5, 6
		//Directories: map[string]*quickfs.FSDirectory{
		//	"dir2": {
		//		Name:        "dir2", // 7
		//		Path:        filepath.Join("root", "dir1", "dir2"),
		//		Files:       makeFiles(filepath.Join("root", "dir1", "dir2"), 2), // 8, 9
		//		Directories: make(map[string]*quickfs.FSDirectory),
		//	},
		//},
	}
	m.Set("dir1", dir)
	var inner = orderedmap.NewOrderedMap[string, *quickfs.FSDirectory]()
	inner.Set("dir2", &quickfs.FSDirectory{
		Name:        "dir2", // 7
		Path:        filepath.Join("root", "dir1", "dir2"),
		Files:       makeFiles(filepath.Join("root", "dir1", "dir2"), 2), // 8, 9
		Directories: orderedmap.NewOrderedMap[string, *quickfs.FSDirectory](),
	})
	dir.Directories = inner
	return m
}

var FileTree = &quickfs.FSDirectory{
	Name:        "root", // 1
	Path:        "root",
	Files:       makeFiles("root", 2), // 2, 3
	Directories: getDirectoryMap(),
}

func TestForEach(t *testing.T) {
	var count int
	FileTree.ForEach(func(fl quickfs.FileLike) (cancel bool, err error) {
		t.Logf("file: %s", fl.GetPath())
		count++
		return
	})
	if count != 9 {
		t.Errorf("expected 9, got %d", count)
	}
}

func TestFind(t *testing.T) {
	var file, err = FileTree.Find([]string{"dir1", "dir2", "file2"})
	if err != nil {
		t.Fatal(err)
	}
	if file.GetPath() != filepath.Join("root", "dir1", "dir2", "file2") {
		t.Errorf("expected %s, got %s", filepath.Join("root", "dir1", "dir2", "file2"), file.GetPath())
	}

	file, err = FileTree.Find([]string{"dir1", "file2"})
	if err != nil {
		t.Fatal(err)
	}
	if file.GetPath() != filepath.Join("root", "dir1", "file2") {
		t.Errorf("expected %s, got %s", filepath.Join("root", "dir1", "file2"), file.GetPath())
	}

	file, err = FileTree.Find([]string{"file2"})
	if err != nil {
		t.Fatal(err)
	}
	if file.GetPath() != filepath.Join("root", "file2") {
		t.Errorf("expected %s, got %s", filepath.Join("root", "file2"), file.GetPath())
	}

	file, err = FileTree.Find([]string{})
	if err != nil {
		t.Fatal(err)
	}

	if file.GetPath() != "root" {
		t.Errorf("expected root, got %s", file.GetPath())
	}
}

func TestFindError(t *testing.T) {
	var err error
	_, err = FileTree.Find(nil)
	if err == nil {
		t.Error("expected error, got nil")
	}

	_, err = FileTree.Find([]string{"dir1", "dir2", "file3"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestAddDirectories(t *testing.T) {
	var dirs = []string{
		"dir1/dir2",
		"dir1/dir3",
		"dir2/dir3",
		"dir2/dir4/dir5",
	}

	var newRoot = &quickfs.FSDirectory{
		Name:        "root",
		Path:        "root",
		Files:       orderedmap.NewOrderedMap[string, *quickfs.FSFile](),
		Directories: orderedmap.NewOrderedMap[string, *quickfs.FSDirectory](),
	}

	for _, dir := range dirs {
		newRoot.AddDirectory(dir)
	}

	var count int
	newRoot.ForEach(func(fl quickfs.FileLike) (cancel bool, err error) {
		t.Logf("dir: %s", fl.GetPath())
		count++
		return
	})

	if count != 8 {
		t.Errorf("expected %d, got %d", 8, count)
	}

	_, err := newRoot.Find([]string{"dir1", "dir2"})
	if err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}

	for _, dir := range dirs {
		var split = strings.Split(
			filepath.Clean(dir), string(os.PathSeparator),
		)
		dI, err := newRoot.Find(split)
		if err != nil {
			t.Fatalf("expected nil, got: %v", err)
		}

		var d = dI.(*quickfs.FSDirectory)

		if d.Name != split[len(split)-1] {
			t.Errorf("expected %s, got %s", split[len(split)-1], d.Name)
		}

		if d.Path != filepath.Join("root", dir) {
			t.Errorf("expected %s, got %s", filepath.Join("root", dir), d.Path)
		}
	}
}

type ByteReader bytes.Reader

func (r *ByteReader) Read(p []byte) (n int, err error) {
	return (*bytes.Reader)(r).Read(p)
}

func (r *ByteReader) Close() error {
	return nil
}

func TestAddFiles(t *testing.T) {
	var files = []string{
		"file1",
		"dir1/file2",
		"dir1/dir2/file3",
		"dir1/dir2/file4",
		"dir2/file5",
		"dir2/dir3/file6",
		"dir2/dir3/file7",
	}

	var newRoot = &quickfs.FSDirectory{
		Name:        "root",
		Path:        "root",
		Files:       orderedmap.NewOrderedMap[string, *quickfs.FSFile](),
		Directories: orderedmap.NewOrderedMap[string, *quickfs.FSDirectory](),
	}

	for _, file := range files {
		var r = bytes.NewReader([]byte(file))
		var rd = (*ByteReader)(r)
		newRoot.AddFile(file, rd)
	}

	var count int
	newRoot.ForEach(func(fl quickfs.FileLike) (cancel bool, err error) {
		t.Logf("file: %s", fl.GetPath())
		count++
		return
	})

	if count != 12 {
		t.Errorf("expected %d, got %d", 12, count)
	}

	for _, file := range files {
		var split = strings.Split(
			filepath.Clean(file), string(os.PathSeparator),
		)
		fI, err := newRoot.Find(split)
		if err != nil {
			t.Fatalf("expected nil, got: %v", err)
		}

		var f = fI.(*quickfs.FSFile)

		if f.Name != split[len(split)-1] {
			t.Errorf("expected %s, got %s", split[len(split)-1], f.Name)
		}

		file = strings.ReplaceAll(file, string(os.PathSeparator), "/")
		path := strings.ReplaceAll(f.Path, string(os.PathSeparator), "/")

		if path != file {
			t.Errorf("expected %s, got %s", file, path)
		}

		var buf = make([]byte, len(file))
		_, err = f.Read(buf)
		if err != nil {
			t.Fatalf("expected nil, got: %v", err)
		}

		if string(buf) != file {
			t.Errorf("expected %s, got %s", file, string(buf))
		}
	}

}
