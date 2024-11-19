package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type ClientConfig struct {
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
	Computation struct {
		SimilarityMetric string `yaml:"similarityMetric"`
	} `yaml:"computation"`
}

func LoadClientConfig() ClientConfig {
	var cfg ClientConfig
	data, err := ioutil.ReadFile("/app/config/client_config.yaml")
	if err != nil {
		log.Fatalf("Error reading client config file: %v", err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Error parsing client config file: %v", err)
	}
	return cfg
}
