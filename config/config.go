package config

import (
	"github.com/modeledge/cleanconfig"
)

type VMUserConfig struct {
	Elastic      Elastic      `toml:"Elastic"`
	Postgres     Postgres     `toml:"Database"`
	Turso        Turso        `toml:"Turso"`
	Server       Server       `toml:"Server"`
	LLM          LLM          `toml:"LLM"`
	LLMLibConfig LLMLibConfig `toml:"LLMLibConfig"`
}

func GetVMUserConfig(path string) *VMUserConfig {
	cfg, err := loadInstallerConfig(path)
	if err == nil {
		return cfg
	}
	return &VMUserConfig{}
}

func loadInstallerConfig(filename string) (*VMUserConfig, error) {
	var config VMUserConfig
	err := cleanconfig.ReadConfig(filename, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
