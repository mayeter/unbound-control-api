package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Unbound   UnboundConfig   `mapstructure:"unbound"`
	Security  SecurityConfig  `mapstructure:"security"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

type ServerConfig struct {
	Port     int    `mapstructure:"port"`
	Host     string `mapstructure:"host"`
	UseTLS   bool   `mapstructure:"use_tls"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type UnboundConfig struct {
	ControlSocket string `mapstructure:"control_socket"`
}

type SecurityConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type RateLimitConfig struct {
	RequestsPerSecond float64 `mapstructure:"requests_per_second"`
	BurstSize         float64 `mapstructure:"burst_size"`
}

type LoggingConfig struct {
	Level     string `mapstructure:"level"`
	UseSyslog bool   `mapstructure:"use_syslog"`
	AppName   string `mapstructure:"app_name"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
