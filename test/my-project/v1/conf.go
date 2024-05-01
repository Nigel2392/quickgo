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

func (c *Configuration) SetEncoder(encoder string) error {
	switch strings.TrimSpace(strings.ToLower(encoder)) {
	case "json":
		AppConfig.Encoder = "json"
	case "gob":
		AppConfig.Encoder = "gob"
	default:
		return fmt.Errorf("invalid serialization method")
	}
	return nil
}

func (c *Configuration) Serialize(dir Directory, path string) error {
	if c.IsJSON() {
		return WriteJSONConfig(dir, c.GetName(path))
	} else if c.IsGob() {
		return WriteGOBConfig(dir, c.GetName(path))
	}
	return fmt.Errorf("invalid serialization")

}

func (c *Configuration) Deserialize(data []byte) (Directory, error) {
	if c.IsJSON() {
		return jsonDecode(data)
	} else if c.IsGob() {
		return gobDecode(data)
	} else {
		return Directory{}, fmt.Errorf("invalid deserialization method")
	}
}

func (c *Configuration) Save(path string) error {
	confJSON, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, confJSON, 0644); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) Copy(conf Configuration) Configuration {
	if c.Host != conf.Host && conf.Host != "" {
		c.Host = conf.Host
	}
	if c.Port != conf.Port && conf.Port != "" {
		c.Port = conf.Port
	}
	if c.Encoder != conf.Encoder && conf.Encoder != "" {
		c.Encoder = conf.Encoder
	}
	return *c
}

func (c *Configuration) GetConfig(path string) (*Configuration, error) {
	conf, err := os.ReadFile(path)
	if err != nil {
		c = &Configuration{
			Host:    "127.0.0.1", // Default host
			Port:    "8080",      // Default port
			Encoder: "json",      // Default encoder
		}
		if err := c.Save(path); err != nil {
			return nil, err
		}
		return c, nil
	} else {
		err = json.Unmarshal(conf, &c)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}
