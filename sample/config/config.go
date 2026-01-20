package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Redis RedisConfig `yaml:"redis"`
}

type RedisConfig struct {
	Username string    `yaml:"username"`
	Password string    `yaml:"password"`
	Masters  NodeGroup `yaml:"masters"`
	Replicas NodeGroup `yaml:"replicas"`
}

type NodeGroup struct {
	Addresses    []string `yaml:"addresses"`
	MaxThreshold float64  `yaml:"maxThreshold"`
}

func Load(path string) (*RedisConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg.Redis, nil
}
