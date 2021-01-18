package main

import (
	"bytes"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Whitelist  []string `yaml:"whitelist"`
	Blacklist  []string `yaml:"blacklist"`
	Exceptions []string `yaml:"exceptions"`
}

func ReadConfig(config []byte) (*Config, error) {

	t := Config{}
	var err error
	if err = yaml.NewDecoder(bytes.NewReader(config)).Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}
