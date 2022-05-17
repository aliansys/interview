package config

import (
	"github.com/aliansys/interview/config"
	"github.com/aliansys/interview/helpers/os"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

const (
	defaultPath = "./config/config.yml"
)

func Parse(fromPath string) (config.Config, error) {
	if fromPath == "" {
		fromPath = defaultPath
	}

	filename, err := filepath.Abs(fromPath)
	if err != nil {
		return config.Config{}, err
	}

	cfg := config.Config{
		Api: struct {
			Address string `yaml:"address"`
		}{
			Address: ":3000",
		},
		ClickHouse: struct {
			DSN       string `yaml:"dsn"`
			BatchSize int    `yaml:"batch_size"`
		}{
			DSN:       "tcp://localhost:9001/interview",
			BatchSize: 1000,
		},
	}

	ok, err := os.Exists(filename)
	if err != nil {
		return config.Config{}, err
	}

	if !ok {
		return cfg, nil
	}
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return config.Config{}, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}
