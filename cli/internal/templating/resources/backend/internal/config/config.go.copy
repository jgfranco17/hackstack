package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServiceSettings struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Load reads configuration from environment variables (prefix: APP_)
// with defaults of 0.0.0.0:8080.
func Load() (ServiceSettings, error) {
	v := viper.New()
	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", 8080)
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	var cfg ServiceSettings
	if err := v.Unmarshal(&cfg); err != nil {
		return ServiceSettings{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
