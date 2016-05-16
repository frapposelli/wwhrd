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
	defer f.Close()
	if err != nil {
		return nil, err
	}

	t := Config{}
	err = yaml.NewDecoder(f).Decode(&t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
