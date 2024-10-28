package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		ConnString string `mapstructure:"conn_string"`
	} `mapstructure:"database"`

	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
