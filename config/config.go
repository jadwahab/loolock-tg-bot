package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AdminTimeout    int  `yaml:"adminTimeout"`
	KickDuration    int  `yaml:"kickDuration"`
	ResponseTimeout int  `yaml:"responseTimeout"`
	BotDebug        bool `yaml:"botDebug"`
}

func LoadConfig(filename string) (Config, error) {
	var config Config

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
