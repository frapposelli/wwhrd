package main

import (
	"os"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
)

type Config struct {
	Whitelist  []string `yaml:"whitelist"`
	Blacklist  []string `yaml:"blacklist"`
	Exceptions []string `yaml:"exceptions"`
}

func ReadConfig(path string) (*Config, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t := Config{}
	if err = yaml.NewDecoder(f).Decode(&t); err != nil {
		return nil, err
	}

	return &t, nil
}
