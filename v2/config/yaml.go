package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadYamlFS[T any](fileSys fs.FS, path string) (*T, error) {
	var (
		data   = new(T)
		f, err = fs.ReadFile(fileSys, path)
	)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(f, data); err != nil {
		return nil, err
	}

	return data, nil
}

func WriteYaml(data interface{}, path string) error {
	var (
		f, err = yaml.Marshal(data)
	)
	if err != nil {
		return err
	}

	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.ReplaceAll(path, "/", "\\")
	path = filepath.Clean(path)

	var (
		dir = filepath.Dir(path)
	)

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(
		path, f,
		os.ModePerm,
	)
}
