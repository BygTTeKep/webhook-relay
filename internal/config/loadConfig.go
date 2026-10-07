package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type DBConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Name     string `mapstructure:"name"`
	Password string `mapstructure:"password"`
}
type AppConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic string `mapstructure:"topic"`
}

type Config struct {
	DBCfg  DBConfig `mapstructure:"database"`
	AppCfg AppConfig `mapstructure:"app"`
	KafkaCfg KafkaConfig `mapstructure:"kafka"`
}

func LoadConfig(path string) (c *Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}