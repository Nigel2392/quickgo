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
	//return compress(data)
}

func gobDecode(data []byte) (Directory, error) {
	var dir Directory
	//data, err := decompress(data)
	//if err != nil {
	//	return dir, err
	//}
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

// When using gob encoder, also compress the data to save space
//func compress(data []byte) ([]byte, error) {
//	var buf bytes.Buffer
//	w := zlib.NewWriter(&buf)
//	_, err := w.Write(data)
//	if err != nil {
//		return nil, err
//	}
//	err = w.Close()
//	if err != nil {
//		return nil, err
//	}
//	return buf.Bytes(), nil
//}
//
//func decompress(data []byte) ([]byte, error) {
//	buf := bytes.NewBuffer(data)
//	r, err := zlib.NewReader(buf)
//	if err != nil {
//		return nil, err
//	}
//	defer r.Close()
//	var out bytes.Buffer
//	_, err = io.Copy(&out, r)
//	if err != nil {
//		return nil, err
//	}
//	return out.Bytes(), nil
//}
//
