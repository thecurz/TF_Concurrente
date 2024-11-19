package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Server struct {
		Port       string `yaml:"port"`
		MaxClients int    `yaml:"maxClients"`
	} `yaml:"server"`
	Dataset struct {
		Path       string `yaml:"path"`
		Partitions int    `yaml:"partitions"`
	} `yaml:"dataset"`
}

func LoadServerConfig() ServerConfig {
	var cfg ServerConfig
	data, err := ioutil.ReadFile("../config/server_config.yaml")
	if err != nil {
		log.Fatalf("Error reading server config file: %v", err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Error parsing server config file: %v", err)
	}
	return cfg
}
