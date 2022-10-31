package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Configuration struct {
	Host    string `json:"host"`
	Port    string `json:"port"`
	Encoder string `json:"encoder"`
}

func (c *Configuration) IsEncType(enctype string) bool {
	return strings.EqualFold(c.Encoder, enctype)
}

func (c *Configuration) IsGob() bool {
	return c.IsEncType("gob")
}

func (c *Configuration) IsJSON() bool {
	return c.IsEncType("json")
}

func (c *Configuration) GetName(name string) string {
	if c.IsJSON() {
		return name + ".json"
	} else if c.IsGob() {
		return name + ".gob"
	}
	return name
}

func (c *Configuration) GetConfig(path string) (*Configuration, error) {
	conf, err := os.ReadFile(path)
	if err != nil {
		c = &Configuration{
			Host:    "127.0.0.1", // Default host
			Port:    "8080",      // Default port
			Encoder: "json",      // Default encoder
		}
		confJSON, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, confJSON, 0644); err != nil {
			return nil, err
		}
	} else {
		err = json.Unmarshal(conf, &c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *Configuration) Serialize(dir Directory, path string) error {
	if c.IsJSON() {
		return WriteJSONConfig(dir, path+".json")
	} else if c.IsGob() {
		return WriteGOBConfig(dir, path+".gob")
	}
	return fmt.Errorf("invalid encoder")

}

func (c *Configuration) Deserialize(data []byte) (Directory, error) {
	if c.IsJSON() {
		return jsonDecode(data)
	} else if c.IsGob() {
		return gobDecode(data)
	} else {
		return Directory{}, fmt.Errorf("invalid encoder")
	}
}
