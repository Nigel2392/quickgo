package config

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadYamlFS[T any](fileSys fs.FS, path string) (*T, error) {
	var f, err = fileSys.Open(path)
	if err != nil {
		return nil, err
	}

	return ReadYaml[T](f)
}

func LoadYaml[T any](path string) (*T, error) {
	var f, err = os.Open(path)
	if err != nil {
		return nil, err
	}

	return ReadYaml[T](f)
}

func ReadYaml[T any](r io.Reader) (*T, error) {
	var (
		data    = new(T)
		decoder = yaml.NewDecoder(r)
		err     error
	)
	if err = decoder.Decode(data); err != nil {
		return nil, err
	}

	if validator, ok := any(data).(Validator); ok {
		if err = validator.Validate(); err != nil {
			return nil, err
		}
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

	path = filepath.ToSlash(path)
	path = filepath.FromSlash(path)
	path = filepath.Clean(path)

	var (
		dir = filepath.Dir(path)
	)

	if _, err = os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return os.WriteFile(
		path, f,
		os.ModePerm,
	)
}
