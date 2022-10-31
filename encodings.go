package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

func WriteJSONConfig(dir Directory, path string) error {
	json_data, err := json.MarshalIndent(dir, "", "  ")
	if err != nil {
		return err
	}
	err = WriteConf(path, json_data)
	if err != nil {
		return err
	}
	return nil
}

func WriteGOBConfig(dir Directory, path string) error {
	gob_data, err := gobEncode(dir)
	if err != nil {
		return err
	}
	err = WriteConf(path, gob_data)
	if err != nil {
		return err
	}
	return nil
}

func gobEncode(dir Directory) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(dir)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gobDecode(data []byte) (Directory, error) {
	var dir Directory
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&dir)
	if err != nil {
		return Directory{}, err
	}
	return dir, nil
}

func jsonDecode(file []byte) (Directory, error) {
	var dir Directory
	err := json.Unmarshal(file, &dir)
	if err != nil {
		return Directory{}, err
	}
	return dir, nil
}
