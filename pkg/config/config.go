package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Port string `yaml:"port"`
}

func Read() *AppConfig {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$PWD/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var appConfig AppConfig
	err = viper.Unmarshal(&appConfig)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	return &appConfig
}
