package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type AppConfig struct {
	Port      string          `yaml:"port"`
	Couchbase CouchbaseConfig `yaml:"couchbase"`
	Jaeger    JaegerConfig    `yaml:"jaeger"`
}

type CouchbaseConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Bucket   string `yaml:"bucket"`
}

type JaegerConfig struct {
	URL string `yaml:"url"`
}

func Read() *AppConfig {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$PWD/config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/config")
	viper.AddConfigPath("./config")
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
